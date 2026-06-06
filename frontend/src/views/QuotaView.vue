<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'
import { useI18n } from 'vue-i18n'

const store = useConsoleStore()
const { t, locale } = useI18n()

// Currently selected Provider ID
const selectedProviderId = ref<number | null>(null)

// Currently selected Provider object
const selectedProvider = computed(() => {
  if (selectedProviderId.value === null) return null
  return store.providers.find(p => p.id === selectedProviderId.value) ?? null
})

// Current provider's mountId (if created)
const currentMountId = computed(() => {
  if (selectedProviderId.value === null) return null
  return store.mountIdByProvider[selectedProviderId.value] ?? null
})

// Whether mount exists
const hasMount = computed(() => currentMountId.value !== null)

// Quota mode selection
const quotaMode = ref<'real' | 'inherit' | 'virtual'>('real')
const virtualTotalInput = ref('')
const virtualTotalMB = computed(() => {
  const n = parseInt(virtualTotalInput.value, 10)
  return isNaN(n) || n <= 0 ? undefined : n
})

// inherit 模式：可选父挂载点（仅 real 模式且不是当前 provider 的 mount）
const inheritParentId = ref<number | null>(null)
const availableParents = computed(() => {
  return store.mounts.filter(m =>
    m.mode === 'real' && m.providerId !== selectedProviderId.value
  )
})

// Status toast
const statusMessage = ref('')
const statusIsError = ref(false)
let statusTimer: ReturnType<typeof setTimeout> | null = null
function showStatus(msg: string, isError = false) {
  if (statusTimer) clearTimeout(statusTimer)
  statusMessage.value = msg
  statusIsError.value = isError
  statusTimer = setTimeout(() => { statusMessage.value = '' }, isError ? 5000 : 2500)
}

// Format bytes to human-readable
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// Backend returns quota in MB, convert to bytes then format
function formatQuotaMB(mb: number): string {
  return formatBytes(mb * 1024 * 1024)
}

// Mock providers store values in GB but backend reports as MB
function displayQuotaMB(mb: number): string {
  const scale = selectedProvider.value?.provider_type === 'mock' ? 1024 : 1
  return formatQuotaMB(mb * scale)
}

// Locale-aware date formatting
function formatTime(t: string | null | undefined): string {
  if (!t) return '—'
  return new Date(t).toLocaleString(locale.value)
}

// Create Mount
async function handleCreateMount() {
  const p = selectedProvider.value
  if (!p) return
  // For local storage, the path is stored in account_id
  const rootPath = p.net_disk === 'local' ? p.account_id : undefined
  const mount = await store.createMountForProvider(
    p.id, p.name, p.provider_type, rootPath,
    quotaMode.value, virtualTotalMB.value,
    quotaMode.value === 'inherit' ? inheritParentId.value ?? undefined : undefined
  )
  if (mount) {
    await store.queryQuotaByMount(mount.id)
    showStatus(t('quota.mount_created'))
  } else {
    showStatus(t('quota.mount_failed'), true)
  }
}

	// Sync quota (call remote API)
async function handleSync() {
  if (currentMountId.value === null) return
  const res = await store.syncQuotaByMount(currentMountId.value)
  if (res && res.code === 1000) {
    showStatus(t('quota.sync_success'))
  } else {
    showStatus(t('quota.sync_failed') + ' ' + (res?.msg || 'check console'), true)
  }
}

// Auto-select first provider when list changes
watch(() => store.providers, (list) => {
  if (list.length > 0 && selectedProviderId.value === null) {
    selectedProviderId.value = list[0].id
  }
}, { immediate: true })

// 切换服务商时自动读取缓存配额
watch(selectedProviderId, (id) => {
  if (id === null) return
  const mountId = store.mountIdByProvider[id]
  if (mountId) {
    store.queryQuotaByMount(mountId)
  } else {
    store.currentQuota = null
    store.currentQuotaMode = null
    store.currentQuotaExtra = null
  }
})

onMounted(async () => {
  await store.fetchProviders()
  // Auto-sync once providers are loaded
  if (selectedProviderId.value !== null) {
    const mountId = store.mountIdByProvider[selectedProviderId.value]
    if (mountId) {
      await store.syncQuotaByMount(mountId)
    }
  }
})
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('quota.title')"
      :description="t('quota.description')"
    />

    <!-- Provider select + actions -->
    <div class="quota-controls">
      <select
        v-model="selectedProviderId"
        class="provider-select"
      >
        <option value="" disabled>{{ t('quota.select_provider') }}</option>
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
          :disabled="store.mountCreating || (quotaMode === 'virtual' && !virtualTotalMB) || (quotaMode === 'inherit' && !inheritParentId)"
          @click="handleCreateMount"
        >
          {{ store.mountCreating ? t('quota.creating_mount') : t('quota.create_mount') }}
        </button>

        <template v-else-if="selectedProvider && hasMount">
          <button
            class="btn btn--primary"
            :disabled="store.quotaLoading"
            @click="handleSync"
          >
            {{ store.quotaLoading ? t('quota.syncing') : t('quota.sync_quota') }}
          </button>
        </template>
      </div>
    </div>

    <!-- Quota mode selection (shown before mount is created) -->
    <div v-if="selectedProvider && !hasMount" class="mode-section">
      <div class="mode-options">
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="real" />
          <div class="mode-option__body">
            <span class="mode-option__title">{{ t('quota.real') }}</span>
            <span class="mode-option__desc">{{ t('quota.real_desc') }}</span>
          </div>
        </label>
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="virtual" />
          <div class="mode-option__body">
            <span class="mode-option__title">{{ t('quota.virtual') }}</span>
            <span class="mode-option__desc">{{ t('quota.virtual_desc') }}</span>
          </div>
        </label>
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="inherit" />
          <div class="mode-option__body">
            <span class="mode-option__title">{{ t('quota.inherit') }}</span>
            <span class="mode-option__desc">{{ t('quota.inherit_desc') }}</span>
          </div>
        </label>
      </div>
      <div v-if="quotaMode === 'inherit' && availableParents.length === 0" class="mode-hint">
        {{ t('quota.inherit_no_parents') }}
      </div>
      <div v-if="quotaMode === 'inherit' && availableParents.length > 0" class="inherit-select">
        <label>{{ t('quota.inherit_parent') }}</label>
        <select v-model="inheritParentId" class="inherit-select__field">
          <option :value="null" disabled>{{ t('quota.inherit_select_hint') }}</option>
          <option v-for="mp in availableParents" :key="mp.id" :value="mp.id">
            {{ mp.providerName }}
          </option>
        </select>
      </div>
      <div v-if="quotaMode === 'virtual'" class="virtual-input">
        <label>{{ t('quota.virtual_total') }}</label>
        <input
          v-model="virtualTotalInput"
          type="number"
          min="1"
          :placeholder="t('quota.placeholder_gb')"
          class="virtual-input__field"
        />
      </div>
    </div>

    <!-- Status feedback -->
    <transition name="fade">
      <div v-if="statusMessage" class="status-toast" :class="{ 'status-toast--error': statusIsError }">{{ statusMessage }}</div>
    </transition>

    <!-- Quota card -->
    <div v-if="store.currentQuota" class="quota-card">
      <div class="quota-card__header">
        <h3>{{ selectedProvider?.name || store.currentQuota.provider }}</h3>
        <div class="quota-card__header-right">
          <span v-if="store.currentQuotaMode" class="mode-badge" :class="`mode-badge--${store.currentQuotaMode}`">
            {{ t('quota.' + store.currentQuotaMode) }}
          </span>
          <span v-if="store.currentQuotaExtra?.inherit_chain?.length" class="inherit-chain">
            ← {{ store.currentQuotaExtra.inherit_chain.join(', ') }}
          </span>
          <span class="quota-card__time">
            {{ t('quota.updated') }} {{ formatTime(store.currentQuota.updated_at) }}
          </span>
        </div>
      </div>

      <div class="quota-stats">
        <div class="quota-stat">
          <span class="quota-stat__label">{{ t('quota.total') }}</span>
          <span class="quota-stat__value">{{ displayQuotaMB(store.currentQuota.total) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">{{ t('quota.used') }}</span>
          <span class="quota-stat__value">{{ displayQuotaMB(store.currentQuota.used) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">{{ t('quota.available') }}</span>
          <span class="quota-stat__value quota-stat__value--available">
            {{ displayQuotaMB(store.currentQuota.available) }}
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
      <p v-if="!selectedProvider">{{ t('quota.empty_no_provider') }}</p>
      <p v-else-if="!hasMount">{{ t('quota.empty_create_mount') }}</p>
      <p v-else>{{ t('quota.empty_no_data') }}</p>
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

.quota-card__header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.quota-card__time {
  font-size: 13px;
  color: #6b7280;
}

.mode-badge {
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}
.mode-badge--real { background: #d1fae5; color: #065f46; }
.mode-badge--inherit { background: #dbeafe; color: #1e40af; }
.mode-badge--virtual { background: #fef3c7; color: #92400e; }

.inherit-chain {
  font-size: 12px;
  color: #6b7280;
  font-family: 'SFMono-Regular', Consolas, monospace;
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

/* ── Quota mode selection ── */
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

.inherit-select {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.inherit-select label {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
}
.inherit-select__field {
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  width: 240px;
  transition: border-color 0.2s;
}
.inherit-select__field:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.mode-hint {
  margin-top: 16px;
  padding: 10px 16px;
  background: #fef3c7;
  color: #92400e;
  border-radius: 8px;
  font-size: 13px;
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