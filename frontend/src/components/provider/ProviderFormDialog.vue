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
    label: '通用（推荐）',
    desc: '通过 OpenList API 转发所有操作，适用于你在 OpenList 中配置的任何网盘'
  },
  {
    value: 'baidu',
    label: '百度网盘（直连，需 token）',
    desc: '百度网盘专用后端，可绕过 OpenList 直连下载，需要配置百度 access token'
  },
  {
    value: 'local',
    label: '本地存储',
    desc: '读取 Windows 本地磁盘的真实容量，需指定盘符路径'
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
  // 本地存储时，将路径存入 account_id 字段（后端不要求它，仅用于创建 Mount 时读取）
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
        <h3>{{ provider ? '编辑 Provider' : '注册 Provider' }}</h3>
        <button class="dialog__close" @click="handleClose">&times;</button>
      </div>

      <div class="dialog__body">
        <div class="form-group">
          <label>名称 *</label>
          <input v-model="formData.name" type="text" placeholder="如：我的夸克网盘" />
        </div>

        <div class="form-group">
          <label>存储后端类型</label>
          <select v-model="formData.net_disk">
            <option v-for="opt in netDiskOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
          <small class="type-hint">
            {{ netDiskOptions.find(o => o.value === formData.net_disk)?.desc }}
          </small>
        </div>

        <!-- 本地存储路径 -->
        <div v-if="formData.net_disk === 'local'" class="form-group">
          <label>本地文件夹路径 *</label>
          <input v-model="localPath" type="text" placeholder="C:\" />
          <small>
            LocalWindowsProvider 会调用 <code>GetDiskFreeSpaceEx</code> 读取该路径所在磁盘的
            真实总容量、已用和可用空间
          </small>
        </div>

        <!-- 百度 access_token -->
        <div v-if="formData.net_disk === 'baidu'" class="form-group">
          <label>百度 Access Token *</label>
          <input v-model="formData.access_token" type="password" placeholder="粘贴百度 access_token" />
          <small>
            获取方式：打开 <code>https://pan.baidu.com</code> → F12 → Network → 过滤 <code>quota</code>，
            在请求 URL 中找到 <code>access_token=xxx</code>，复制 <code>xxx</code> 部分
          </small>
        </div>

        <!-- 通用 / 百度 模式显示账户标识 -->
        <div v-if="formData.net_disk !== 'local'" class="form-group">
          <label>账户标识（可选）</label>
          <input v-model="formData.account_id" type="text" placeholder="仅用于备注" />
          <small>方便你区分同一类型下的多个账户，不填也行</small>
        </div>

        <div v-if="provider" class="form-group">
          <label>状态</label>
          <select v-model="formData.status">
            <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <div class="notice">
          <strong>关于配额</strong>
          <p v-if="formData.net_disk === 'baidu'">
            选择百度网盘直连后，创建 Mount → 同步配额，BaiduProvider 会使用你提供的 access_token
            直接调百度 API 获取真实容量数据。
          </p>
          <p v-else-if="formData.net_disk === 'local'">
            选择本地存储后，创建 Mount → 同步配额，LocalWindowsProvider 会直接读取你指定的
            磁盘路径的真实容量。这是当前唯一能显示真实配额的选项。
          </p>
          <p v-else>
            注册完成后，为该 Provider 创建 Mount（挂载点），然后在 Mount 页面同步配额即可。
            通用模式下显示的是测试数据。
          </p>
        </div>
      </div>

      <div class="dialog__footer">
        <button class="btn btn--secondary" @click="handleClose">取消</button>
        <button class="btn btn--primary" @click="handleSubmit">
          {{ provider ? '保存' : '注册' }}
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
