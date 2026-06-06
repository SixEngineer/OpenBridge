<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ProviderRecord } from '@/types/provider'

const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  provider?: ProviderRecord | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'submit', data: Partial<ProviderRecord>): void
}>()

const localPath = ref('C:\\')

const formData = ref<Partial<ProviderRecord>>({
  name: '',
  net_disk: 'mock',
  account_id: '',
  access_token: '',
  auth_cookie: '',
  status: 'active',
  total_quota: 0,
  used_quota: 0,
  available_quota: 0
})

const netDiskOptions = [
  {
    value: 'mock',
    labelKey: 'provider_form.options.generic',
    descKey: 'provider_form.options.generic_desc',
  },
  {
    value: 'baidu',
    labelKey: 'provider_form.options.baidu',
    descKey: 'provider_form.options.baidu_desc',
  },
  {
    value: 'quark',
    labelKey: 'provider_form.options.quark',
    descKey: 'provider_form.options.quark_desc',
  },
  {
    value: 'local',
    labelKey: 'provider_form.options.local',
    descKey: 'provider_form.options.local_desc',
  }
]

const statusOptions = [
  { value: 'active', labelKey: 'provider_form.status_options.active' },
  { value: 'disabled', labelKey: 'provider_form.status_options.disabled' }
]

watch(() => props.provider, (val) => {
  if (val) {
    formData.value = { ...val }
    localPath.value = val.account_id || 'C:\\'
  } else {
    formData.value = {
      name: '',
      net_disk: 'mock',
      account_id: '',
      access_token: '',
      auth_cookie: '',
      status: 'active',
      total_quota: 0,
      used_quota: 0,
      available_quota: 0
    }
    localPath.value = 'C:\\'
  }
}, { immediate: true })

function handleClose() {
  emit('update:visible', false)
}

function handleSubmit() {
  // For local storage, store the path in account_id (used when creating Mount)
  const data: Partial<ProviderRecord> = {
    ...formData.value,
    provider_type: formData.value.net_disk || 'mock',
  }
  if (data.net_disk === 'local') {
    data.account_id = localPath.value
  }
  emit('submit', data)
  handleClose()
}
</script>

<template>
  <div v-if="visible" class="dialog-overlay" @click.self="handleClose">
    <div class="dialog">
      <div class="dialog__header">
        <h3>{{ provider ? t('provider_form.title_edit') : t('provider_form.title_add') }}</h3>
        <button class="dialog__close" @click="handleClose">&times;</button>
      </div>

      <div class="dialog__body">
        <div class="form-group">
          <label>{{ t('provider_form.name') }}</label>
          <input v-model="formData.name" type="text" :placeholder="t('provider_form.name_placeholder')" />
        </div>

        <div class="form-group">
          <label>{{ t('provider_form.backend_type') }}</label>
          <select v-model="formData.net_disk">
            <option v-for="opt in netDiskOptions" :key="opt.value" :value="opt.value">
              {{ t(opt.labelKey) }}
            </option>
          </select>
          <small class="type-hint">
            {{ t(netDiskOptions.find(o => o.value === formData.net_disk)?.descKey || '') }}
          </small>
        </div>

        <!-- Local storage path -->
        <div v-if="formData.net_disk === 'local'" class="form-group">
          <label>{{ t('provider_form.local_path') }}</label>
          <input v-model="localPath" type="text" :placeholder="t('provider_form.local_path_placeholder')" />
          <small v-html="t('provider_form.local_path_hint')"></small>
        </div>

        <!-- Baidu access token -->
        <div v-if="formData.net_disk === 'baidu'" class="form-group">
          <label>{{ t('provider_form.baidu_token') }}</label>
          <input v-model="formData.access_token" type="password" :placeholder="t('provider_form.baidu_token_placeholder')" />
          <small v-html="t('provider_form.baidu_token_hint')"></small>
        </div>

        <!-- Quark cookie -->
        <div v-if="formData.net_disk === 'quark'" class="form-group">
          <label>{{ t('provider_form.quark_cookie') }}</label>
          <input v-model="formData.auth_cookie" type="password" :placeholder="t('provider_form.quark_cookie_placeholder')" />
          <small v-html="t('provider_form.quark_cookie_hint')"></small>
        </div>

        <!-- Account ID (for non-local types) -->
        <div v-if="formData.net_disk !== 'local'" class="form-group">
          <label>{{ t('provider_form.account_id') }}</label>
          <input v-model="formData.account_id" type="text" :placeholder="t('provider_form.account_id_placeholder')" />
          <small>{{ t('provider_form.account_id_hint') }}</small>
        </div>

        <div v-if="provider" class="form-group">
          <label>{{ t('provider_form.status') }}</label>
          <select v-model="formData.status">
            <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ t(opt.labelKey) }}
            </option>
          </select>
        </div>

        <div class="notice">
          <strong>{{ t('provider_form.about_quota') }}</strong>
          <p v-if="formData.net_disk === 'baidu'" v-html="t('provider_form.baidu_notice')"></p>
          <p v-else-if="formData.net_disk === 'quark'" v-html="t('provider_form.quark_notice')"></p>
          <p v-else-if="formData.net_disk === 'local'" v-html="t('provider_form.local_notice')"></p>
          <p v-else v-html="t('provider_form.generic_notice')"></p>
        </div>
      </div>

      <div class="dialog__footer">
        <button class="btn btn--secondary" @click="handleClose">{{ t('provider_form.cancel') }}</button>
        <button class="btn btn--primary" @click="handleSubmit">
          {{ provider ? t('provider_form.save') : t('provider_form.register') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 12px;
  width: 540px;
  max-width: 90vw;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.dialog__header {
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dialog__header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.dialog__close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6b7280;
  padding: 0;
  line-height: 1;
}

.dialog__close:hover {
  color: #374151;
}

.dialog__body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 6px;
  color: #374151;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group small {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #6b7280;
}

.form-group code {
  background: #f3f4f6;
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 12px;
}

.type-hint {
  font-style: italic;
}

.notice {
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 14px 16px;
  margin-top: 8px;
}

.notice strong {
  font-size: 13px;
  color: #1e40af;
}

.notice p {
  margin: 8px 0 0;
  font-size: 13px;
  color: #1e40af;
  line-height: 1.5;
}

.dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
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

.btn--secondary {
  background: white;
  color: #374151;
  border: 1px solid #d1d5db;
}

.btn--secondary:hover {
  background: #f9fafb;
}
</style>
