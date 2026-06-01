<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'

const store = useConsoleStore()

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

// Create Mount
async function handleCreateMount() {
  const p = selectedProvider.value
  if (!p) return
  // For local storage, the path is stored in account_id
  const rootPath = p.net_disk === 'local' ? p.account_id : undefined
  const mount = await store.createMountForProvider(
    p.id, p.name, p.provider_type, rootPath,
    quotaMode.value, virtualTotalMB.value
  )
  if (mount) {
    await store.queryQuotaByMount(mount.id)
    showStatus('Mount created successfully, click Sync Quota to get data')
  } else {
    showStatus('Failed to create mount', true)
  }
}

// Query quota (read local cache)
async function handleQuery() {
  if (currentMountId.value === null) return
  const res = await store.queryQuotaByMount(currentMountId.value)
  if (res && res.code === 1000) {
    showStatus('Local cache quota read')
  } else {
    showStatus('Query failed: ' + (res?.msg || 'check console'), true)
  }
}

// Sync quota (call remote API)
async function handleSync() {
  if (currentMountId.value === null) return
  const res = await store.syncQuotaByMount(currentMountId.value)
  if (res && res.code === 1000) {
    showStatus('Quota sync succeeded')
  } else {
    showStatus('Sync failed: ' + (res?.msg || 'check console'), true)
  }
}

// Auto-select first provider when list changes
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
      description="Query and sync drive quota through Mount points"
    />

    <!-- Provider select + actions -->
    <div class="quota-controls">
      <select
        v-model="selectedProviderId"
        class="provider-select"
      >
        <option value="" disabled>Select a provider</option>
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
          {{ store.mountCreating ? 'Creating...' : 'Create Mount & Query Quota' }}
        </button>

        <template v-else-if="selectedProvider && hasMount">
          <button
            class="btn btn--secondary"
            :disabled="store.quotaLoading"
            @click="handleQuery"
          >
            Query Quota
          </button>
          <button
            class="btn btn--primary"
            :disabled="store.quotaLoading"
            @click="handleSync"
          >
            {{ store.quotaLoading ? 'Syncing...' : 'Sync Quota' }}
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
            <span class="mode-option__title">Real</span>
            <span class="mode-option__desc">Real capacity (from cloud/disk)</span>
          </div>
        </label>
        <label class="mode-option">
          <input type="radio" v-model="quotaMode" value="virtual" />
          <div class="mode-option__body">
            <span class="mode-option__title">Virtual</span>
            <span class="mode-option__desc">Virtual capacity (specify total manually)</span>
          </div>
        </label>
        <label class="mode-option mode-option--disabled" title="Requires an existing Mount as parent, not available yet">
          <input type="radio" disabled />
          <div class="mode-option__body">
            <span class="mode-option__title">Inherit</span>
            <span class="mode-option__desc">Inherit parent quota (not available)</span>
          </div>
        </label>
      </div>
      <div v-if="quotaMode === 'virtual'" class="virtual-input">
        <label>Virtual Total Capacity (MB)</label>
        <input
          v-model="virtualTotalInput"
          type="number"
          min="1"
          placeholder="e.g. 102400 (100 GB)"
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
        <h3>{{ store.currentQuota.provider }}</h3>
        <span class="quota-card__time">
          Updated: {{ new Date(store.currentQuota.updated_at).toLocaleString() }}
        </span>
      </div>

      <div class="quota-stats">
        <div class="quota-stat">
          <span class="quota-stat__label">Total</span>
          <span class="quota-stat__value">{{ formatQuotaMB(store.currentQuota.total) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">Used</span>
          <span class="quota-stat__value">{{ formatQuotaMB(store.currentQuota.used) }}</span>
        </div>
        <div class="quota-stat">
          <span class="quota-stat__label">Available</span>
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
      <p v-if="!selectedProvider">Register a provider on the Providers page first</p>
      <p v-else-if="!hasMount">Click "Create Mount & Query Quota"</p>
      <p v-else>No quota data yet. Click "Query Quota" or "Sync Quota"</p>
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