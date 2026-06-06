<script setup lang="ts">
import { onMounted, ref } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ProviderFormDialog from '@/components/provider/ProviderFormDialog.vue'
import { useConsoleStore } from '@/stores/console'
import type { ProviderRecord } from '@/types/provider'
import { registerProvider, updateProvider } from '@/api/provider'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const store = useConsoleStore()

// Dialog state
const dialogVisible = ref(false)
const editingProvider = ref<ProviderRecord | null>(null)

// Toast notification
const toast = ref({ show: false, message: '', type: 'success' as 'success' | 'error' })
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(message: string, type: 'success' | 'error' = 'success') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { show: true, message, type }
  toastTimer = setTimeout(() => {
    toast.value.show = false
  }, 2500)
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Backend stores quota in MB, convert to bytes for formatting
function formatProviderQuota(mb: number): string {
  return formatBytes(mb * 1024 * 1024)
}

// Mock providers store values in GB but backend reports as MB
function displayQuota(mb: number, providerType: string): string {
  const scale = providerType === 'mock' ? 1024 : 1
  return formatProviderQuota(mb * scale)
}

function openAddDialog() {
  editingProvider.value = null
  dialogVisible.value = true
}

function openEditDialog(provider: ProviderRecord) {
  editingProvider.value = provider
  dialogVisible.value = true
}

async function handleSubmit(data: Partial<ProviderRecord>) {
  // 检查同名（编辑时排除自身）
  if (!editingProvider.value?.id) {
    const exists = store.providers.some(p => p.name === data.name)
    if (exists) {
      showToast(t('providers.name_exists'), 'error')
      return
    }
  }
  try {
    let res
    if (editingProvider.value?.id) {
      res = await updateProvider({ ...data, id: editingProvider.value.id })
    } else {
      res = await registerProvider(data)
    }

    if (res.code === 1000 || res.code === 0) {
      showToast(editingProvider.value ? t('providers.updated_success') : t('providers.registered_success'))
      dialogVisible.value = false
      editingProvider.value = null
      await store.fetchProviders()
    } else {
      showToast('Failed: ' + (res.msg), 'error')
    }
  } catch (error: any) {
    console.error('Operation failed', error)
    showToast(error?.message || t('common.operation_failed'), 'error')
  }
}

async function handleDelete(provider: ProviderRecord) {
  if (!confirm(`${t('providers.delete_confirm')} "${provider.name}"?`)) {
    return
  }

  const success = await store.removeProvider(provider.id)
  if (success) {
    showToast(t('providers.delete_success'))
  } else {
    showToast(t('providers.delete_failed'), 'error')
  }
}

onMounted(() => {
  store.fetchProviders()
})
</script>

<template>
  <section class="page">
    <PageHeader :title="t('providers.title')" :description="t('providers.description')">
      <template #actions>
        <button class="btn btn--primary" @click="openAddDialog">
          {{ t('providers.register_btn') }}
        </button>
      </template>
    </PageHeader>

    <div v-if="store.providers.length === 0" class="empty-state">
      <p>{{ t('providers.empty') }}</p>
    </div>

    <div v-else class="provider-grid">
      <article v-for="provider in store.providers" :key="provider.id" class="provider-card">
        <div class="provider-card__header">
          <div>
            <p class="provider-card__name">{{ provider.name }}</p>
            <p class="provider-card__id">{{ provider.provider_type }} · {{ provider.net_disk }}</p>
          </div>
          <div class="provider-card__header-right">
            <StatusBadge :state="provider.status" />
            <button
              class="provider-card__edit"
              @click="openEditDialog(provider)"
              :title="t('providers.edit_title')"
            >
              ✏️
            </button>
            <button
              class="provider-card__delete"
              @click="handleDelete(provider)"
              :title="t('providers.delete_title')"
            >
              🗑️
            </button>
          </div>
        </div>

        <template v-if="provider.net_disk === 'local'">
          <p class="provider-card__section-title">{{ t('providers.local_path') }}</p>
          <p class="provider-card__text">{{ provider.account_id || t('providers.not_set') }}</p>
        </template>
        <div v-else class="provider-card__placeholder"></div>

        <p class="provider-card__section-title">{{ t('providers.quota_usage') }}</p>
        <p class="provider-card__text">
          {{ t('providers.total') }} {{ displayQuota(provider.total_quota, provider.provider_type) }}<br>
          {{ t('providers.used') }} {{ displayQuota(provider.used_quota, provider.provider_type) }}<br>
          {{ t('providers.available') }} {{ displayQuota(provider.available_quota, provider.provider_type) }}
        </p>
        
        <p class="provider-card__section-title" v-if="provider.last_error">Last Error</p>
        <p class="provider-card__text provider-card__text--error" v-if="provider.last_error">
          {{ provider.last_error }}
        </p>
      </article>
    </div>

    <ProviderFormDialog
      v-model:visible="dialogVisible"
      :provider="editingProvider"
      @submit="handleSubmit"
    />

    <transition name="toast-fade">
      <div v-if="toast.show" class="toast" :class="`toast--${toast.type}`">
        {{ toast.message }}
      </div>
    </transition>
  </section>
</template>

<style scoped>
.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.provider-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
  transition: all 0.2s;
}

.provider-card:hover {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border-color: #d1d5db;
}

.provider-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 16px;
}

.provider-card__header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-card__edit {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  opacity: 0.5;
  transition: all 0.2s;
}

.provider-card__edit:hover {
  opacity: 1;
  background: #e0e7ff;
}

.provider-card__delete {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  opacity: 0.5;
  transition: all 0.2s;
}

.provider-card__delete:hover {
  opacity: 1;
  background: #fee2e2;
}

.provider-card__name {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: #111827;
}

.provider-card__id {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

.provider-card__section-title {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #6b7280;
  margin: 16px 0 4px 0;
}

.provider-card__text {
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #374151;
}

.provider-card__text--error {
  color: #ef4444;
}

.provider-card__placeholder {
  height: 43px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
  font-size: 16px;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
}

.btn {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--primary:hover {
  background: #2563eb;
}

/* ── Toast ── */
.toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  pointer-events: none;
}
.toast--success {
  background: #d1fae5;
  color: #065f46;
  border: 1px solid #a7f3d0;
}
.toast--error {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.toast-fade-enter-active { transition: all 0.3s ease; }
.toast-fade-leave-active { transition: all 0.3s ease; }
.toast-fade-enter-from { opacity: 0; transform: translateX(-50%) translateY(-12px); }
.toast-fade-leave-to { opacity: 0; transform: translateX(-50%) translateY(-12px); }
</style>