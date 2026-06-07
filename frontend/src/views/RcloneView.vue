<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import RcloneProfileDialog from '@/components/rclone/RcloneProfileDialog.vue'
import { useConsoleStore } from '@/stores/console'
import { useI18n } from 'vue-i18n'
import {
  applyRcloneProfile,
  createRcloneProfile,
  deleteRcloneProfile,
  listRcloneProfiles,
  mountRcloneProfile,
  updateRcloneProfile,
} from '@/api/rclone'
import type { RcloneProfile, RcloneProfileInput } from '@/types/rclone'

const { t, locale } = useI18n()
const store = useConsoleStore()

const profiles = ref<RcloneProfile[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const editingProfile = ref<RcloneProfile | null>(null)
const applyingId = ref<number | null>(null)
const mountingId = ref<number | null>(null)

const toast = ref({ show: false, message: '', type: 'success' as 'success' | 'error' })
let toastTimer: ReturnType<typeof setTimeout> | null = null

const availableMounts = computed(() => store.allMounts.filter(m => m.enabled))

function showToast(message: string, type: 'success' | 'error' = 'success') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { show: true, message, type }
  toastTimer = setTimeout(() => {
    toast.value.show = false
  }, 2400)
}

async function fetchProfiles() {
  loading.value = true
  try {
    const res = await listRcloneProfiles()
    profiles.value = res.data
  } catch (error: any) {
    showToast(error?.message || t('common.request_error'), 'error')
  } finally {
    loading.value = false
  }
}

function openAddDialog() {
  editingProfile.value = null
  dialogVisible.value = true
}

function openEditDialog(profile: RcloneProfile) {
  editingProfile.value = profile
  dialogVisible.value = true
}

async function handleSubmit(data: RcloneProfileInput) {
  try {
    if (editingProfile.value) {
      await updateRcloneProfile(editingProfile.value.id, data)
    } else {
      await createRcloneProfile(data)
    }
    dialogVisible.value = false
    await fetchProfiles()
    showToast(t('rclone.save_success'))
  } catch (error: any) {
    showToast(error?.message || t('rclone.save_failed'), 'error')
  }
}

async function handleDelete(profile: RcloneProfile) {
  if (!confirm(t('rclone.delete_confirm', { name: profile.name }))) {
    return
  }
  try {
    await deleteRcloneProfile(profile.id)
    await fetchProfiles()
    showToast(t('rclone.delete_success'))
  } catch (error: any) {
    showToast(error?.message || t('rclone.delete_failed'), 'error')
  }
}

async function handleApply(profile: RcloneProfile) {
  applyingId.value = profile.id
  try {
    await applyRcloneProfile(profile.id)
    await fetchProfiles()
    showToast(t('rclone.apply_success'))
  } catch (error: any) {
    showToast(error?.message || t('rclone.apply_failed'), 'error')
  } finally {
    applyingId.value = null
  }
}

async function handleMount(profile: RcloneProfile) {
  mountingId.value = profile.id
  try {
    await mountRcloneProfile(profile.id)
    await fetchProfiles()
    showToast(t('rclone.mount_success'))
  } catch (error: any) {
    showToast(error?.message || t('rclone.mount_failed'), 'error')
  } finally {
    mountingId.value = null
  }
}

async function copyText(value: string) {
  await navigator.clipboard.writeText(value)
  showToast(t('rclone.copied'))
}

function formatTime(value?: string | null) {
  if (!value) return '—'
  return new Date(value).toLocaleString(locale.value)
}

function modeLabel(mode: string) {
  const key = `rclone.${mode}`
  return t(key)
}

onMounted(async () => {
  await store.fetchAllMounts()
  await fetchProfiles()
})
</script>

<template>
  <section class="page">
    <PageHeader :title="t('rclone.title')" :description="t('rclone.description')">
      <template #actions>
        <button v-if="store.isAdmin" class="btn btn--primary" @click="openAddDialog">
          {{ t('rclone.add') }}
        </button>
      </template>
    </PageHeader>

    <div v-if="loading" class="empty-state">{{ t('common.refresh') }}...</div>
    <div v-else-if="profiles.length === 0" class="empty-state">{{ t('rclone.empty') }}</div>

    <div v-else class="profile-grid">
      <article v-for="profile in profiles" :key="profile.id" class="profile-card">
        <div class="profile-card__header">
          <div>
            <h3>{{ profile.name }}</h3>
            <p>{{ modeLabel(profile.mode) }} · {{ profile.target_path }}</p>
          </div>
          <div class="profile-card__actions" v-if="store.isAdmin">
            <button class="icon-btn" @click="openEditDialog(profile)">{{ t('rclone.edit') }}</button>
            <button class="icon-btn icon-btn--danger" @click="handleDelete(profile)">{{ t('rclone.delete') }}</button>
          </div>
        </div>

        <div class="meta-grid">
          <div>
            <span class="meta-label">{{ t('rclone.mount_ids') }}</span>
            <span class="meta-value">{{ profile.mount_ids.join(', ') }}</span>
          </div>
          <div>
            <span class="meta-label">{{ t('rclone.username') }}</span>
            <span class="meta-value">{{ profile.username }}</span>
          </div>
          <div>
            <span class="meta-label">{{ t('rclone.saved_password') }}</span>
            <span class="meta-value">{{ profile.password_saved ? t('common.status_active') : t('common.status_disabled') }}</span>
          </div>
          <div>
            <span class="meta-label">{{ t('rclone.last_applied') }}</span>
            <span class="meta-value">{{ profile.last_applied_at ? formatTime(profile.last_applied_at) : t('rclone.not_applied') }}</span>
          </div>
          <div>
            <span class="meta-label">{{ t('rclone.last_mounted') }}</span>
            <span class="meta-value">{{ profile.last_mounted_at ? formatTime(profile.last_mounted_at) : t('rclone.not_mounted') }}</span>
          </div>
        </div>

        <div v-if="profile.last_error" class="error-box">{{ profile.last_error }}</div>

        <div class="command-block">
          <p class="command-title">{{ t('rclone.config_command') }}</p>
          <code v-for="command in profile.apply_commands" :key="command" class="command-line">{{ command }}</code>
          <button class="btn btn--ghost" @click="copyText(profile.apply_commands.join('\n'))">{{ t('rclone.copy') }}</button>
        </div>

        <div class="command-block">
          <p class="command-title">{{ t('rclone.mount_command') }}</p>
          <code class="command-line">{{ profile.mount_command }}</code>
          <button class="btn btn--ghost" @click="copyText(profile.mount_command)">{{ t('rclone.copy') }}</button>
        </div>

        <div class="footer-actions" v-if="store.isAdmin">
          <button class="btn btn--secondary" :disabled="applyingId === profile.id" @click="handleApply(profile)">
            {{ applyingId === profile.id ? t('rclone.apply') + '...' : t('rclone.apply') }}
          </button>
          <button class="btn btn--primary" :disabled="mountingId === profile.id" @click="handleMount(profile)">
            {{ mountingId === profile.id ? t('rclone.mount') + '...' : t('rclone.mount') }}
          </button>
        </div>
      </article>
    </div>

    <RcloneProfileDialog
      v-model:visible="dialogVisible"
      :profile="editingProfile"
      :mounts="availableMounts"
      @submit="handleSubmit"
    />

    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="`toast--${toast.type}`">{{ toast.message }}</div>
    </transition>
  </section>
</template>

<style scoped>
.profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 18px;
}

.profile-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 18px;
  box-shadow: var(--shadow);
}

.profile-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.profile-card__header h3 {
  margin: 0 0 4px;
  color: var(--text);
}

.profile-card__header p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
}

.profile-card__actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  border-radius: 10px;
  padding: 8px 10px;
  cursor: pointer;
}

.icon-btn--danger {
  color: #dc2626;
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}

.meta-label {
  display: block;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
  margin-bottom: 4px;
}

.meta-value {
  color: var(--text);
  font-size: 14px;
  word-break: break-word;
}

.command-block {
  margin-top: 14px;
  padding: 12px;
  border-radius: 12px;
  background: var(--surface-strong);
  border: 1px solid var(--border);
}

.command-title {
  margin: 0 0 8px;
  font-size: 12px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.command-line {
  display: block;
  padding: 9px 10px;
  margin-bottom: 8px;
  border-radius: 10px;
  background: var(--surface);
  color: var(--text);
  font-family: Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.footer-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.btn {
  border: none;
  border-radius: 10px;
  padding: 10px 14px;
  font-size: 14px;
  cursor: pointer;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--secondary,
.btn--ghost {
  background: var(--surface);
  border: 1px solid var(--border);
  color: var(--text);
}

.btn--ghost {
  padding: 8px 12px;
}

.empty-state {
  text-align: center;
  padding: 52px 20px;
  color: var(--muted);
  background: var(--surface);
  border: 1px dashed var(--border);
  border-radius: 16px;
}

.error-box {
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.22);
  color: #dc2626;
  border-radius: 12px;
  padding: 10px 12px;
  font-size: 13px;
}

.toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 20px;
  border-radius: 10px;
  z-index: 9999;
}

.toast--success {
  background: rgba(42, 167, 106, 0.15);
  color: #2aa76a;
  border: 1px solid rgba(42, 167, 106, 0.3);
}

.toast--error {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.3s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-12px);
}

@media (max-width: 860px) {
  .profile-grid {
    grid-template-columns: 1fr;
  }

  .meta-grid {
    grid-template-columns: 1fr;
  }

  .footer-actions {
    flex-direction: column;
  }
}
</style>
