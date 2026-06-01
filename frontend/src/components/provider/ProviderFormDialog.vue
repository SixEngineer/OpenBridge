<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ProviderRecord } from '@/types/provider'

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
  status: 'active',
  total_quota: 0,
  used_quota: 0,
  available_quota: 0
})

const netDiskOptions = [
  {
    value: 'mock',
    label: 'Generic (Recommended)',
    desc: 'Routes all operations through the OpenList API. Works with any drive configured in OpenList.'
  },
  {
    value: 'baidu',
    label: 'Baidu Netdisk (Direct, needs token)',
    desc: 'Dedicated Baidu backend with direct download support. Requires a Baidu access token.'
  },
  {
    value: 'local',
    label: 'Local Storage',
    desc: 'Reads real disk capacity from the local filesystem. Requires a drive path.'
  }
]

const statusOptions = [
  { value: 'active', label: 'Active' },
  { value: 'disabled', label: 'Disabled' }
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
        <h3>{{ provider ? 'Edit Provider' : 'Register Provider' }}</h3>
        <button class="dialog__close" @click="handleClose">&times;</button>
      </div>

      <div class="dialog__body">
        <div class="form-group">
          <label>Name *</label>
          <input v-model="formData.name" type="text" placeholder="e.g. My Cloud Drive" />
        </div>

        <div class="form-group">
          <label>Backend Type</label>
          <select v-model="formData.net_disk">
            <option v-for="opt in netDiskOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
          <small class="type-hint">
            {{ netDiskOptions.find(o => o.value === formData.net_disk)?.desc }}
          </small>
        </div>

        <!-- Local storage path -->
        <div v-if="formData.net_disk === 'local'" class="form-group">
          <label>Local Folder Path *</label>
          <input v-model="localPath" type="text" placeholder="C:\" />
          <small>
            Uses <code>GetDiskFreeSpaceEx</code> to read the real total, used, and available
            capacity of the disk containing this path.
          </small>
        </div>

        <!-- Baidu access token -->
        <div v-if="formData.net_disk === 'baidu'" class="form-group">
          <label>Baidu Access Token *</label>
          <input v-model="formData.access_token" type="password" placeholder="Paste Baidu access_token" />
          <small>
            How to get: Open <code>https://pan.baidu.com</code> → F12 → Network → filter <code>quota</code>,
            find <code>access_token=xxx</code> in the request URL, copy the <code>xxx</code> part.
          </small>
        </div>

        <!-- Account ID (for non-local types) -->
        <div v-if="formData.net_disk !== 'local'" class="form-group">
          <label>Account ID (optional)</label>
          <input v-model="formData.account_id" type="text" placeholder="For reference only" />
          <small>Helps distinguish multiple accounts of the same type. Can be left blank.</small>
        </div>

        <div v-if="provider" class="form-group">
          <label>Status</label>
          <select v-model="formData.status">
            <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <div class="notice">
          <strong>About Quota</strong>
          <p v-if="formData.net_disk === 'baidu'">
            After registration, create a Mount → sync quota. The BaiduProvider will use your access_token
            to fetch real capacity data from the Baidu API.
          </p>
          <p v-else-if="formData.net_disk === 'local'">
            After registration, create a Mount → sync quota. The LocalWindowsProvider will read the
            real disk capacity of the specified path.
          </p>
          <p v-else>
            After registration, create a Mount, then sync quota on the Quota page.
            Generic mode shows test data.
          </p>
        </div>
      </div>

      <div class="dialog__footer">
        <button class="btn btn--secondary" @click="handleClose">Cancel</button>
        <button class="btn btn--primary" @click="handleSubmit">
          {{ provider ? 'Save' : 'Register' }}
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
