<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import LocalPathInput from '@/components/common/LocalPathInput.vue'
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
const netDiskOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

function toggleNetDisk() {
  netDiskOpen.value = !netDiskOpen.value
}

function selectNetDisk(value: string) {
  formData.value.net_disk = value
  netDiskOpen.value = false
}

// Click outside to close
function onDocumentClick(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    netDiskOpen.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => {
  document.removeEventListener('click', onDocumentClick)
  document.body.style.overflow = ''
})

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

// 每次打开新增弹窗时重置表单（provider 为 null 时 watch 不会重复触发）
watch(() => props.visible, (v) => {
  if (v) {
    document.body.style.overflow = 'hidden'
    if (!props.provider) {
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
  } else {
    document.body.style.overflow = ''
  }
})

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
  <div v-if="visible" class="dialog-overlay" @mousedown.self="handleClose">
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
          <div ref="dropdownRef" class="cdropdown">
            <button
              class="cdropdown__trigger"
              :class="{ 'cdropdown__trigger--open': netDiskOpen }"
              type="button"
              @click.prevent="toggleNetDisk"
            >
              <span class="provider-type-tag" :class="`provider-type-tag--${formData.net_disk}`">
                {{ t(netDiskOptions.find(o => o.value === formData.net_disk)?.labelKey || '') }}
              </span>
              <svg class="cdropdown__arrow" :class="{ 'cdropdown__arrow--open': netDiskOpen }" width="14" height="14" viewBox="0 0 16 16" fill="none">
                <path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
            <transition name="fade">
              <div v-if="netDiskOpen" class="cdropdown__panel">
                <button
                  v-for="opt in netDiskOptions"
                  :key="opt.value"
                  class="cdropdown__item"
                  :class="{ 'cdropdown__item--active': formData.net_disk === opt.value }"
                  type="button"
                  @click.prevent="selectNetDisk(opt.value)"
                >
                  <span class="provider-type-tag" :class="`provider-type-tag--${opt.value}`">
                    {{ t(opt.labelKey) }}
                  </span>
                  <span class="cdropdown__item-desc">{{ t(opt.descKey) }}</span>
                </button>
              </div>
            </transition>
          </div>
          <small class="type-hint">
            {{ t(netDiskOptions.find(o => o.value === formData.net_disk)?.descKey || '') }}
          </small>
        </div>

        <!-- Local storage path -->
        <div v-if="formData.net_disk === 'local'" class="form-group">
          <label>{{ t('provider_form.local_path') }}</label>
          <LocalPathInput
            v-model="localPath"
            mode="directory"
            :placeholder="t('provider_form.local_path_placeholder')"
            :title="t('provider_form.local_path')"
          />
          <small v-html="t('provider_form.local_path_hint')"></small>
        </div>

        <!-- Baidu access token -->
        <div v-if="formData.net_disk === 'baidu'" class="form-group">
          <label>{{ t('provider_form.baidu_token') }}</label>
          <div class="token-input-row">
            <input v-model="formData.access_token" type="password" :placeholder="t('provider_form.baidu_token_placeholder')" />
            <a href="https://api.oplist.org" target="_blank" class="btn btn--sm btn--link">{{ t('provider_form.get_token') }}</a>
          </div>
          <small v-html="t('provider_form.baidu_token_hint')"></small>
        </div>

        <!-- Quark cookie -->
        <div v-if="formData.net_disk === 'quark'" class="form-group">
          <label>{{ t('provider_form.quark_cookie') }}</label>
          <div class="token-input-row">
            <input v-model="formData.auth_cookie" type="password" :placeholder="t('provider_form.quark_cookie_placeholder')" />
            <a href="https://pan.quark.cn" target="_blank" class="btn btn--sm btn--link">{{ t('provider_form.get_cookie') }}</a>
          </div>
          <small v-html="t('provider_form.quark_cookie_hint')"></small>
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
  background: var(--surface-strong);
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
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dialog__header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
}

.dialog__close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--muted);
  padding: 0;
  line-height: 1;
}

.dialog__close:hover {
  color: var(--text);
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
  color: var(--text);
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
  background: var(--surface);
  color: var(--text);
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(91, 192, 190, 0.15);
}

.token-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}
.token-input-row input {
  flex: 1;
}
.btn--link {
  display: inline-flex;
  align-items: center;
  padding: 10px 16px;
  background: #3b82f6;
  color: white;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.2s;
}
.btn--link:hover {
  background: #2563eb;
  color: white;
}

.form-group small {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--muted);
}

.form-group code {
  background: var(--surface);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 12px;
}

.type-hint {
  font-style: italic;
}

/* ── Custom dropdown ── */
.cdropdown {
  position: relative;
}

.cdropdown__trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text);
  transition: all 0.2s;
  text-align: left;
  -webkit-appearance: none;
  appearance: none;
}

.cdropdown__trigger:hover,
.cdropdown__trigger--open {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(91, 192, 190, 0.15);
}

.cdropdown__arrow {
  flex-shrink: 0;
  color: var(--muted);
  margin-left: auto;
  transition: transform 0.25s ease;
}

.cdropdown__arrow--open {
  transform: rotate(180deg);
}

.cdropdown__panel {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  z-index: 100;
}

.cdropdown__item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  text-align: left;
  color: var(--text);
  transition: background 0.15s;
}

.cdropdown__item:hover {
  background: rgba(59, 130, 246, 0.06);
}

.cdropdown__item--active {
  background: rgba(59, 130, 246, 0.1);
}

.cdropdown__item-desc {
  font-size: 12px;
  color: var(--muted);
  flex: 1;
  text-align: right;
}

/* Provider type tags */
.provider-type-tag {
  font-size: 12px;
  padding: 2px 10px;
  border-radius: 20px;
  font-weight: 600;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}
.provider-type-tag--mock {
  background: rgba(156, 163, 175, 0.15);
  color: #6b7280;
}
.provider-type-tag--baidu {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
}
.provider-type-tag--quark {
  background: rgba(251, 146, 60, 0.12);
  color: #f97316;
}
.provider-type-tag--local {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

[data-theme="dark"] .provider-type-tag--mock {
  background: rgba(156, 163, 175, 0.2);
  color: #9ca3af;
}
[data-theme="dark"] .provider-type-tag--baidu {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}
[data-theme="dark"] .provider-type-tag--quark {
  background: rgba(251, 146, 60, 0.2);
  color: #fb923c;
}
[data-theme="dark"] .provider-type-tag--local {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

/* Fade transition */
.fade-enter-active {
  transition: all 0.15s ease-out;
}
.fade-leave-active {
  transition: all 0.1s ease-in;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.notice {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 14px 16px;
  margin-top: 8px;
}

.notice strong {
  font-size: 13px;
  color: var(--accent);
}

.notice p {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--accent);
  line-height: 1.5;
}

.dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border);
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
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
}

.btn--secondary:hover {
  background: var(--surface-strong);
}

@media (max-width: 860px) {
  .dialog-overlay {
    align-items: flex-end;
  }

  .dialog {
    width: 100%;
    max-width: 100%;
    max-height: 92vh;
    border-radius: 16px 16px 0 0;
    margin-top: auto;
  }

  .dialog__header {
    padding: 16px 20px;
    position: sticky;
    top: 0;
    background: var(--surface-strong);
    z-index: 1;
  }

  .dialog__body {
    padding: 16px 20px;
  }

  .dialog__footer {
    padding: 16px 20px;
    position: sticky;
    bottom: 0;
    background: var(--surface-strong);
    flex-direction: column;
  }

  .dialog__footer .btn {
    width: 100%;
  }

  .token-input-row {
    flex-direction: column;
    align-items: stretch;
  }

  .btn--link {
    justify-content: center;
  }
}
</style>
