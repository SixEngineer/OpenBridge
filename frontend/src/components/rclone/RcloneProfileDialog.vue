<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LocalPathInput from '@/components/common/LocalPathInput.vue'
import type { MountPoint } from '@/types/mount'
import type { RcloneMode, RcloneProfile, RcloneProfileInput } from '@/types/rclone'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  profile?: RcloneProfile | null
  mounts: MountPoint[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'submit', data: RcloneProfileInput): void
}>()

const form = ref<RcloneProfileInput>({
  name: '',
  mode: 'ordinary',
  mount_ids: [],
  username: '',
  password: '',
  target_path: '',
})
const error = ref('')

const modeOptions: { value: RcloneMode; label: string; hint: string }[] = [
  { value: 'ordinary', label: 'rclone.ordinary', hint: 'rclone.mode_hint_ordinary' },
  { value: 'union', label: 'rclone.union', hint: 'rclone.mode_hint_union' },
  { value: 'combine', label: 'rclone.combine', hint: 'rclone.mode_hint_combine' },
]

const selectedCount = computed(() => form.value.mount_ids.length)

watch(() => props.profile, (profile) => {
  error.value = ''
  if (profile) {
    form.value = {
      name: profile.name,
      mode: profile.mode,
      mount_ids: [...profile.mount_ids],
      username: profile.username,
      password: '',
      target_path: profile.target_path,
    }
    return
  }

  form.value = {
    name: '',
    mode: 'ordinary',
    mount_ids: [],
    username: '',
    password: '',
    target_path: '',
  }
}, { immediate: true })

watch(() => form.value.mode, (mode) => {
  if (mode === 'ordinary' && form.value.mount_ids.length > 1) {
    form.value.mount_ids = form.value.mount_ids.slice(0, 1)
  }
})

function handleClose() {
  emit('update:visible', false)
}

function toggleMount(mountId: number) {
  if (form.value.mode === 'ordinary') {
    form.value.mount_ids = form.value.mount_ids[0] === mountId ? [] : [mountId]
    return
  }

  const set = new Set(form.value.mount_ids)
  if (set.has(mountId)) set.delete(mountId)
  else set.add(mountId)
  form.value.mount_ids = Array.from(set)
}

function handleSubmit() {
  error.value = ''
  if (!form.value.name.trim() || !form.value.username.trim() || !form.value.target_path.trim()) {
    error.value = t('common.request_error')
    return
  }
  if (!props.profile && !form.value.password?.trim()) {
    error.value = t('rclone.password_required')
    return
  }
  if (form.value.mode === 'ordinary' && form.value.mount_ids.length !== 1) {
    error.value = t('rclone.mount_select_hint')
    return
  }
  if ((form.value.mode === 'union' || form.value.mode === 'combine') && form.value.mount_ids.length < 2) {
    error.value = t('rclone.mount_select_hint')
    return
  }
  emit('submit', {
    ...form.value,
    name: form.value.name.trim(),
    username: form.value.username.trim(),
    password: form.value.password?.trim() || '',
    target_path: form.value.target_path.trim(),
  })
}
</script>

<template>
  <div v-if="visible" class="dialog-overlay" @mousedown.self="handleClose">
    <div class="dialog">
      <div class="dialog__header">
        <h3>{{ profile ? t('rclone.dialog_edit') : t('rclone.dialog_add') }}</h3>
        <button class="dialog__close" @click="handleClose">&times;</button>
      </div>

      <div class="dialog__body">
        <div class="form-group">
          <label>{{ t('rclone.config_name') }}</label>
          <input v-model="form.name" type="text" :placeholder="t('rclone.name_placeholder')" />
        </div>

        <div class="form-group">
          <label>{{ t('rclone.mode') }}</label>
          <div class="mode-grid">
            <button
              v-for="option in modeOptions"
              :key="option.value"
              type="button"
              class="mode-card"
              :class="{ 'mode-card--active': form.mode === option.value }"
              @click="form.mode = option.value"
            >
              <span class="mode-card__title">{{ t(option.label) }}</span>
              <span class="mode-card__hint">{{ t(option.hint) }}</span>
            </button>
          </div>
        </div>

        <div class="form-group">
          <label>{{ t('rclone.mount_ids') }}</label>
          <small>{{ t('rclone.mount_select_hint') }}</small>
          <div v-if="mounts.length > 0" class="mount-list">
            <button
              v-for="mount in mounts"
              :key="mount.id"
              type="button"
              class="mount-chip"
              :class="{ 'mount-chip--active': form.mount_ids.includes(mount.id) }"
              @click="toggleMount(mount.id)"
            >
              <strong>{{ mount.name }}</strong>
              <span>{{ t('rclone.mount_label', { id: mount.id }) }}</span>
            </button>
          </div>
          <p v-else class="empty-hint">{{ t('rclone.no_mounts') }}</p>
          <div v-if="selectedCount > 0" class="selected-hint">{{ t('rclone.select_mounts') }}: {{ selectedCount }}</div>
        </div>

        <div class="form-group">
          <label>{{ t('rclone.username') }}</label>
          <input v-model="form.username" type="text" :placeholder="t('rclone.username_placeholder')" />
        </div>

        <div class="form-group">
          <label>{{ t('rclone.password') }}</label>
          <input v-model="form.password" type="password" :placeholder="t('rclone.password_placeholder')" />
          <small>{{ t('rclone.password_hint') }}</small>
        </div>

        <div class="form-group">
          <label>{{ t('rclone.target_path') }}</label>
          <LocalPathInput
            v-model="form.target_path"
            mode="directory"
            :placeholder="t('rclone.target_path_placeholder')"
            :title="t('rclone.target_path')"
          />
          <small>{{ t('rclone.target_path_hint') }}</small>
        </div>

        <p v-if="error" class="form-error">{{ error }}</p>
      </div>

      <div class="dialog__footer">
        <button class="btn btn--secondary" @click="handleClose">{{ t('rclone.cancel') }}</button>
        <button class="btn btn--primary" @click="handleSubmit">{{ t('rclone.save') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  width: min(760px, 92vw);
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  background: var(--surface-strong);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.18);
}

.dialog__header,
.dialog__footer {
  padding: 18px 22px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dialog__footer {
  border-top: 1px solid var(--border);
  border-bottom: 0;
  justify-content: flex-end;
  gap: 10px;
}

.dialog__close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--muted);
  cursor: pointer;
}

.dialog__body {
  padding: 22px;
  overflow-y: auto;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-group label {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
  color: var(--text);
  font-size: 14px;
  box-sizing: border-box;
}

.form-group small,
.selected-hint,
.empty-hint {
  font-size: 12px;
  color: var(--muted);
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.mode-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  text-align: left;
  padding: 14px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
}

.mode-card--active {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.12);
}

.mode-card__title {
  font-weight: 700;
}

.mode-card__hint {
  font-size: 12px;
  color: var(--muted);
}

.mount-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mount-chip {
  min-width: 180px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.mount-chip--active {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.08);
}

.mount-chip span {
  font-size: 12px;
  color: var(--muted);
}

.form-error {
  margin: 0;
  color: #dc2626;
  font-size: 13px;
}

.btn {
  padding: 10px 16px;
  border-radius: 10px;
  border: none;
  font-size: 14px;
  cursor: pointer;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--secondary {
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
}

@media (max-width: 860px) {
  .dialog {
    width: 100%;
    max-height: 95vh;
    border-radius: 18px 18px 0 0;
    align-self: flex-end;
  }

  .mode-grid {
    grid-template-columns: 1fr;
  }

  .mount-chip {
    width: 100%;
  }

  .dialog__footer {
    flex-direction: column;
  }

  .dialog__footer .btn {
    width: 100%;
  }
}
</style>
