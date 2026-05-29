<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'

const store = useConsoleStore()

// 当前选中的 Provider ID
const selectedProviderId = ref<number | null>(null)

// 当前选中的 Provider 对象
const selectedProvider = computed(() => {
  if (selectedProviderId.value === null) return null
  return store.providers.find(p => p.id === selectedProviderId.value) ?? null
})

// 当前 provider 的 mountId（如果已创建）
const currentMountId = computed(() => {
  if (selectedProviderId.value === null) return null
  return store.mountIdByProvider[selectedProviderId.value] ?? null
})

// 是否有 mount
const hasMount = computed(() => currentMountId.value !== null)

// 配额模式选择
const quotaMode = ref<'real' | 'inherit' | 'virtual'>('real')
const virtualTotalInput = ref('')
const virtualTotalMB = computed(() => {
  const n = parseInt(virtualTotalInput.value, 10)
  return isNaN(n) || n <= 0 ? undefined : n
})

// 操作状态提示
const statusMessage = ref('')
const statusIsError = ref(false)
let statusTimer: ReturnType<typeof setTimeout> | null = null
function showStatus(msg: string, isError = false) {
  if (statusTimer) clearTimeout(statusTimer)
  statusMessage.value = msg
  statusIsError.value = isError
  statusTimer = setTimeout(() => { statusMessage.value = '' }, isError ? 5000 : 2500)
}

// 格式化字节
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 后端以 MB 为单位返回配额，先转成字节再格式化
function formatQuotaMB(mb: number): string {
  return formatBytes(mb * 1024 * 1024)
}

// 创建 Mount
async function handleCreateMount() {
  const p = selectedProvider.value
  if (!p) return
  // 本地存储时，路径存在 account_id 中
  const rootPath = p.net_disk === 'local' ? p.account_id : undefined
  const mount = await store.createMountForProvider(
    p.id, p.name, p.provider_type, rootPath,
    quotaMode.value, virtualTotalMB.value
  )
  if (mount) {
    await store.queryQuotaByMount(mount.id)
    showStatus('Mount 创建成功，请点击"同步配额"获取数据')
  } else {
    showStatus('创建 Mount 失败', true)
  }
}

// 查询配额（读本地缓存）
async function handleQuery() {
  if (currentMountId.value === null) return
  const res = await store.queryQuotaByMount(currentMountId.value)
  if (res && res.code === 1000) {
    showStatus('已读取本地缓存配额')
  } else {
    showStatus('查询失败：' + (res?.msg || '请查看控制台日志'), true)
  }
}

// 同步配额（调远端 API）
async function handleSync() {
  if (currentMountId.value === null) return
  const res = await store.syncQuotaByMount(currentMountId.value)
  if (res && res.code === 1000) {
    showStatus('配额同步成功')
  } else {
    showStatus('同步失败：' + (res?.msg || '请查看控制台日志'), true)
  }
}

// Provider 列表变化时默认选中第一个
watch(() => store.providers, (list) => {
  if (list.length > 0 && selectedProviderId.value === null) {
    selectedProviderId.value = list[0].id
  }
}, { immediate: true })

onMounted(() => {
  store.fetchProviders()
})
</script>

<template>
  <section class="page">
    <PageHeader
      title="Quota Management"
      description="通过 Mount 挂载点查询和同步网盘配额"
    />

    <!-- Provider 选择 + 操作区 -->
    <div class="quota-controls">
      <select
        v-model="selectedProviderId"
        class="provider-select"
      >
        <option value="" disabled>请选择 Provider</option>
        <option
          v-for="p in store.providers"
          :key="p.id"
          :value="p.id"
        >
          {{ p.name }} ({{ p.provider_type }})
        </option>
      </select>

      <div class="button-group">
        <button
          v-if="selectedProvider && !hasMount"
          class="btn btn--primary"
          :disabled="store.mountCreating || (quotaMode === 'virtual' && !virtualTotalMB)"
          @click="handleCreateMount"
        >
          {{ store.mountCreating ? '创建中...' : '创建 Mount 并查询配额' }}
        </button>

        <template v-else-if="selectedProvider && hasMount">
          <button
            class="btn btn--secondary"
            :disabled="store.quotaLoading"
            @click="handleQuery"
          >
            查询配额
          </button>
          <button
            class="btn btn--primary"
            :disabled="store.quotaLoading"
            @click="handleSync"
          >
            {{ store.quotaLoading ? '同步中...' : '同步配额' }}
          </button>
        </template>
      </div>
    </div>

    <!-- 配额模式选择（创建 mount 前显示） -->
    <div v-if="selectedProvider && !hasMount" class="mode-section">
      <div class="mode-options">
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="real" />
          <div class="mode-option__body">
            <span class="mode-option__title">Real</span>
            <span class="mode-option__desc">真实容量（从网盘/磁盘获取）</span>
          </div>
        </label>
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="virtual" />
          <div class="mode-option__body">
            <span class="mode-option__title">Virtual</span>
            <span class="mode-option__desc">虚拟容量（手动指定总量）</span>
          </div>
        </label>
        <label class="mode-option mode-option--disabled" title="需要已有的 Mount 作为父级，暂不可用">
          <input type="radio" disabled />
          <div class="mode-option__body">
            <span class="mode-option__title">Inherit</span>
            <span class="mode-option__desc">继承父级配额（暂不可用）</span>
          </div>
        </label>
      </div>
      <div v-if="quotaMode === 'virtual'" class="virtual-input">
        <label>虚拟总容量 (MB)</label>
        <input
          v-model="virtualTotalInput"
          type="number"
          min="1"
          placeholder="e.g. 102400 (100 GB)"
          class="virtual-input__field"
        />
      </div>
    </div>

    <!-- 操作反馈 -->
    <transition name="fade">
      <div v-if="statusMessage" class="status-toast" :class="{ 'status-toast--error': statusIsError }">{{ statusMessage }}</div>
    </transition>

    <!-- 配额卡片 -->
    <div v-if="store.currentQuota" class="quota-card">
      <div class="quota-card__header">
        <h3>{{ store.currentQuota.provider }}</h3>
        <span class="quota-card__time">
          更新时间: {{ new Date(store.currentQuota.updated_at).toLocaleString() }}
        </span>
      </div>

      <div class="quota-stats">
        <div class="quota-stat">
          <span class="quota-stat__label">总配额</span>
          <span class="quota-stat__value">{{ formatQuotaMB(store.currentQuota.total) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">已使用</span>
          <span class="quota-stat__value">{{ formatQuotaMB(store.currentQuota.used) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">可用</span>
          <span class="quota-stat__value quota-stat__value--available">
            {{ formatQuotaMB(store.currentQuota.available) }}
          </span>
        </div>
      </div>

      <div class="quota-progress">
        <div
          class="quota-progress__bar"
          :style="{ width: `${(store.currentQuota.used / store.currentQuota.total) * 100}%` }"
        ></div>
      </div>
    </div>

    <div v-else class="empty-state">
      <p v-if="!selectedProvider">请先在 Providers 页面注册一个 Provider</p>
      <p v-else-if="!hasMount">请点击"创建 Mount 并查询配额"</p>
      <p v-else>暂无配额数据，请点击"查询配额"或"同步配额"</p>
    </div>
  </section>
</template>

<style scoped>
.quota-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 30px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.provider-select {
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  min-width: 200px;
}

.button-group {
  display: flex;
  gap: 12px;
  margin-left: auto;
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

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn--secondary {
  background: white;
  color: #374151;
  border: 1px solid #d1d5db;
}

.btn--secondary:hover:not(:disabled) {
  background: #f9fafb;
}

.quota-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  border: 1px solid #e5e7eb;
}

.quota-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.quota-card__header h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  text-transform: capitalize;
}

.quota-card__time {
  font-size: 13px;
  color: #6b7280;
}

.quota-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 24px;
}

.quota-stat {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.quota-stat__label {
  font-size: 14px;
  color: #6b7280;
}

.quota-stat__value {
  font-size: 28px;
  font-weight: 600;
  color: #111827;
}

.quota-stat__value--available {
  color: #10b981;
}

.quota-progress {
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  overflow: hidden;
}

.quota-progress__bar {
  height: 100%;
  background: #3b82f6;
  border-radius: 4px;
  transition: width 0.3s;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
}

/* ── 配额模式选择 ── */
.mode-section {
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.mode-options {
  display: flex;
  gap: 12px;
}

.mode-option {
  flex: 1;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 14px;
  border: 1px solid #d1d5db;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}
.mode-option:has(input:checked) {
  border-color: #3b82f6;
  background: #eff6ff;
}
.mode-option:hover:not(.mode-option--disabled) {
  border-color: #9ca3af;
  background: #f9fafb;
}
.mode-option--disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #f9fafb;
}
.mode-option input[type="radio"] {
  margin-top: 3px;
}

.mode-option__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.mode-option__title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.mode-option__desc {
  font-size: 12px;
  color: #6b7280;
}

.virtual-input {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.virtual-input label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
}

.virtual-input__field {
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  width: 240px;
  transition: border-color 0.2s;
}
.virtual-input__field:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.status-toast {
  padding: 10px 16px;
  border-radius: 8px;
  background: #3b82f6;
  color: white;
  font-size: 14px;
  text-align: center;
}

.status-toast--error {
  background: #dc2626;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>