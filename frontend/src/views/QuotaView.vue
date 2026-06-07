<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, reactive } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'
import { useI18n } from 'vue-i18n'
import type { MountPoint } from '@/types/mount'
import type { QuotaInfo } from '@/types/quota'
import { getProviderList } from '@/api/provider'

const store = useConsoleStore()
const { t, locale } = useI18n()

const backendStatus = ref<'active' | 'error' | 'disabled'>('active')

async function checkBackend() {
  try {
    const res = await getProviderList()
    backendStatus.value = (res.code === 1000 || res.code === 0) ? 'active' : 'error'
  } catch {
    backendStatus.value = 'error'
  }
}

// ── Provider selection ──
const selectedProviderId = ref<number | null>(null)
const providerDropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const selectedProvider = computed(() => {
  if (selectedProviderId.value === null) return null
  return store.providers.find(p => p.id === selectedProviderId.value) ?? null
})

function selectProvider(id: number) {
  selectedProviderId.value = id
  providerDropdownOpen.value = false
}

function toggleDropdown() {
  if (backendStatus.value !== 'error') {
    providerDropdownOpen.value = !providerDropdownOpen.value
  }
}

// Click outside to close
function onDocumentClick(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    providerDropdownOpen.value = false
  }
}
onMounted(() => document.addEventListener('click', onDocumentClick))
onUnmounted(() => document.removeEventListener('click', onDocumentClick))

// Auto-select first provider when list changes
watch(() => store.providers, (list) => {
  if (list.length > 0 && selectedProviderId.value === null) {
    selectedProviderId.value = list[0].id
  }
}, { immediate: true })

// ── Mounts for selected provider ──
const providerMounts = computed(() => {
  if (selectedProviderId.value === null) return []
  return store.getMountsByProvider(selectedProviderId.value)
})

// ── Create Mount mode ──
const showCreateForm = ref(false)
const quotaMode = ref<'real' | 'inherit' | 'virtual'>('real')
const virtualTotalInput = ref('')
const virtualTotalMB = computed(() => {
  const n = parseInt(virtualTotalInput.value, 10)
  return isNaN(n) || n <= 0 ? undefined : n
})
const inheritParentId = ref<number | null>(null)
const availableParents = computed(() => {
  return store.mounts.filter(m =>
    m.mode === 'real' && m.providerId !== selectedProviderId.value
  )
})

// ── Per-mount quota cache ──
interface MountQuotaCache {
  quota: QuotaInfo | null
  mode: string | null
  extra: { inherit_chain?: number[]; virtual_config?: Record<string, number> } | null
  loading: boolean
}
const mountQuotaMap = reactive<Record<number, MountQuotaCache>>({})

function getMountCache(mountId: number): MountQuotaCache {
  if (!mountQuotaMap[mountId]) {
    mountQuotaMap[mountId] = { quota: null, mode: null, extra: null, loading: false }
  }
  return mountQuotaMap[mountId]
}

// ── Edit mount ──
const editingMountId = ref<number | null>(null)
const editForm = reactive<{
  name: string
  quota_mode: 'real' | 'inherit' | 'virtual'
  virtual_total: number
  virtual_used: number
  inherit_from_id: number | null
}>({
  name: '',
  quota_mode: 'real',
  virtual_total: 0,
  virtual_used: 0,
  inherit_from_id: null,
})

// 编辑时可选的继承父挂载点（排除自己和子挂载点链）
const editAvailableParents = computed(() => {
  return store.mounts.filter(m =>
    m.mode === 'real' && m.id !== editingMountId.value
  )
})

function startEdit(mount: MountPoint) {
  editingMountId.value = mount.id
  editForm.name = mount.name
  editForm.quota_mode = mount.quota_mode
  editForm.virtual_total = mount.virtual_total
  editForm.virtual_used = mount.virtual_used
  editForm.inherit_from_id = mount.inherit_from_id ?? null
}

// 编辑中切换模式到 virtual 时，自动填入 real 的当前配额值
watch(() => editForm.quota_mode, (newMode) => {
  if (newMode !== 'virtual' || editingMountId.value === null) return
  const cache = getMountCache(editingMountId.value)
  if (cache.quota) {
    editForm.virtual_total = cache.quota.total
    editForm.virtual_used = cache.quota.used
  }
})

function cancelEdit() {
  editingMountId.value = null
}

async function handleEditSave(mountId: number) {
  const mount = store.allMounts.find(m => m.id === mountId)
  if (!mount) return
  const payload: Partial<MountPoint> = { name: editForm.name }
  // 判断模式是否改变
  const modeChanged = editForm.quota_mode !== mount.quota_mode
  if (modeChanged) {
    payload.quota_mode = editForm.quota_mode
  }
  if (editForm.quota_mode === 'virtual') {
    payload.virtual_total = editForm.virtual_total
    payload.virtual_used = editForm.virtual_used
  } else if (editForm.quota_mode === 'inherit') {
    payload.inherit_from_id = editForm.inherit_from_id ?? undefined
  }
  const result = await store.updateMountById(mountId, payload)
  if (result) {
    // 模式切换后重新查询配额
    if (modeChanged) {
      await handleSyncMount(mountId)
    }
    showStatus(t('quota.edit_success'))
    editingMountId.value = null
  } else {
    showStatus(t('quota.edit_failed'), true)
  }
}

// ── Delete mount ──
const deleteConfirmId = ref<number | null>(null)

function confirmDelete(mountId: number) {
  deleteConfirmId.value = mountId
}

function cancelDelete() {
  deleteConfirmId.value = null
}

async function handleDelete(mountId: number) {
  const ok = await store.deleteMountById(mountId)
  if (ok) {
    showStatus(t('quota.delete_success'))
    deleteConfirmId.value = null
  } else {
    showStatus(t('quota.delete_failed'), true)
  }
}

// ── Sync / Query per mount ──
async function handleSyncMount(mountId: number) {
  const cache = getMountCache(mountId)
  if (cache.loading) return
  cache.loading = true
  try {
    const res = await store.syncQuotaByMount(mountId)
    if (res && res.code === 1000) {
      cache.quota = res.data.quota
      cache.mode = res.data.mode
      cache.extra = {
        inherit_chain: res.data.inherit_chain,
        virtual_config: res.data.virtual_config,
      }
      showStatus(t('quota.sync_success'))
    } else {
      showStatus((t('quota.sync_failed') + ' ' + (res?.msg || '')), true)
    }
  } catch {
    showStatus(t('quota.sync_failed'), true)
  } finally {
    cache.loading = false
  }
}

async function handleQueryMount(mountId: number) {
  const cache = getMountCache(mountId)
  if (cache.loading) return // 避免重复请求
  cache.loading = true
  try {
    const res = await store.queryQuotaByMount(mountId)
    if (res && res.code === 1000) {
      cache.quota = res.data.quota
      cache.mode = res.data.mode
      cache.extra = {
        inherit_chain: res.data.inherit_chain,
        virtual_config: res.data.virtual_config,
      }
    }
    // res.code !== 1000 或 res === null: 静默处理（mount 可能不存在，不弹错误 toast）
  } finally {
    cache.loading = false
  }
}

// ── Create Mount ──
async function handleCreateMount() {
  const p = selectedProvider.value
  if (!p) return
  const rootPath = p.net_disk === 'local' ? p.account_id : undefined
  const mount = await store.createMountForProvider(
    p.id, p.name, p.provider_type, rootPath,
    quotaMode.value, virtualTotalMB.value,
    quotaMode.value === 'inherit' ? inheritParentId.value ?? undefined : undefined
  )
  if (mount) {
    showStatus(t('quota.mount_created'))
    showCreateForm.value = false
    // Auto-query the new mount's quota
    await handleQueryMount(mount.id)
  } else {
    showStatus(t('quota.mount_failed'), true)
  }
}

// ── Status toast ──
const statusMessage = ref('')
const statusIsError = ref(false)
let statusTimer: ReturnType<typeof setTimeout> | null = null
function showStatus(msg: string, isError = false) {
  if (statusTimer) clearTimeout(statusTimer)
  statusMessage.value = msg
  statusIsError.value = isError
  statusTimer = setTimeout(() => { statusMessage.value = '' }, isError ? 5000 : 2500)
}

// ── Formatting helpers ──
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatQuotaMB(mb: number): string {
  return formatBytes(mb * 1024 * 1024)
}

function displayQuotaMB(mb: number, mountMode?: string | null): string {
  // 虚拟模式值就是 MB，不应用 mock 缩放
  // 继承模式配额来自父 provider，值已经是正确单位，不额外缩放
  if (mountMode === 'virtual' || mountMode === 'inherit') {
    return formatQuotaMB(mb)
  }
  const scale = selectedProvider.value?.provider_type === 'mock' ? 1024 : 1
  return formatQuotaMB(mb * scale)
}

function formatTime(t: string | null | undefined): string {
  if (!t) return '—'
  return new Date(t).toLocaleString(locale.value)
}

// ── Loading state ──
const initialLoading = ref(true)
const mountsLoaded = ref(false) // 防止 watch 在初始加载时触发查询

// ── Lifecycle ──
onMounted(async () => {
  await checkBackend()
  await store.fetchProviders()
  await store.fetchAllMounts()
  mountsLoaded.value = true
  // 只查询当前选中 provider 的挂载点配额，并且并行执行
  if (providerMounts.value.length > 0) {
    await Promise.all(
      providerMounts.value.map(mount => handleQueryMount(mount.id))
    )
  }
  initialLoading.value = false
})

// When provider changes, try to load existing mount quotas
watch(selectedProviderId, async () => {
  // 初始加载期间不做查询 — 由 onMounted 统一处理
  if (!mountsLoaded.value) return
  const promises = providerMounts.value
    .filter(mount => !getMountCache(mount.id).quota)
    .map(mount => handleQueryMount(mount.id))
  if (promises.length > 0) {
    await Promise.all(promises)
  }
})
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('quota.title')"
      :description="t('quota.description')"
    />

    <!-- Backend offline warning -->
    <div v-if="backendStatus === 'error'" class="offline-banner">
      {{ t('dashboard.backend_api_disconnected') }}
    </div>

    <!-- Provider select + Create button -->
    <div class="quota-controls">
      <div ref="dropdownRef" class="provider-dropdown">
        <button
          class="dropdown-trigger"
          :class="{ 'dropdown-trigger--disabled': backendStatus === 'error' }"
          @click="toggleDropdown"
        >
          <template v-if="selectedProvider">
            <span class="dropdown-trigger__name">{{ selectedProvider.name }}</span>
            <span class="provider-type-tag" :class="`provider-type-tag--${selectedProvider.provider_type}`">
              {{ selectedProvider.provider_type }}
            </span>
          </template>
          <span v-else class="dropdown-trigger__placeholder">{{ t('quota.select_provider') }}</span>
          <svg class="dropdown-arrow" :class="{ 'dropdown-arrow--open': providerDropdownOpen }" width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <transition name="dropdown-fade">
          <div v-if="providerDropdownOpen" class="dropdown-panel">
            <button
              v-for="p in store.providers"
              :key="p.id"
              class="dropdown-item"
              :class="{ 'dropdown-item--active': selectedProviderId === p.id }"
              @click="selectProvider(p.id)"
            >
              <div class="dropdown-item__info">
                <span class="dropdown-item__name">{{ p.name }}</span>
                <span class="dropdown-item__desc">{{ p.net_disk === 'local' ? p.account_id : '' }}</span>
              </div>
              <span class="provider-type-tag" :class="`provider-type-tag--${p.provider_type}`">
                {{ p.provider_type }}
              </span>
            </button>
          </div>
        </transition>
      </div>

      <div class="button-group">
        <button
          v-if="selectedProvider && store.isAdmin"
          class="btn btn--primary"
          :disabled="backendStatus === 'error'"
          @click="showCreateForm = !showCreateForm"
        >
          {{ showCreateForm ? t('quota.cancel') : t('quota.create_new_mount') }}
        </button>
      </div>
    </div>

    <!-- Create Mount form -->
    <div v-if="selectedProvider && showCreateForm" class="mode-section">
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
      <div class="create-actions">
        <button
          class="btn btn--primary"
          :disabled="store.mountCreating || backendStatus === 'error' || (quotaMode === 'virtual' && !virtualTotalMB) || (quotaMode === 'inherit' && !inheritParentId)"
          @click="handleCreateMount"
        >
          {{ store.mountCreating ? t('quota.creating_mount') : t('quota.create_mount') }}
        </button>
      </div>
    </div>

    <!-- Initial loading spinner -->
    <div v-if="initialLoading" class="loading-state">
      <span class="loading-spinner"></span>
      <p>{{ t('quota.loading_mounts') }}</p>
    </div>

    <!-- Status toast -->
    <transition name="fade">
      <div v-if="statusMessage" class="status-toast" :class="{ 'status-toast--error': statusIsError }">{{ statusMessage }}</div>
    </transition>

    <!-- Mount cards list -->
    <div v-if="!initialLoading && selectedProvider && providerMounts.length > 0" class="mount-list">
      <div
        v-for="mount in providerMounts"
        :key="mount.id"
        class="mount-card"
        :class="{ 'mount-card--stale': backendStatus === 'error' }"
      >
        <!-- Card header -->
        <div class="mount-card__header">
          <div class="mount-card__header-left">
            <h3 class="mount-card__title">{{ mount.name }}</h3>
            <span class="mode-badge" :class="`mode-badge--${mount.quota_mode}`">
              {{ t('quota.' + mount.quota_mode) }}
            </span>
            <span v-if="getMountCache(mount.id).extra?.inherit_chain?.length" class="inherit-chain">
              &larr; {{ getMountCache(mount.id).extra!.inherit_chain!.join(', ') }}
            </span>
          </div>
          <div class="mount-card__header-right">
            <button v-if="store.isAdmin" class="btn btn--small btn--secondary" @click="startEdit(mount)" :disabled="backendStatus === 'error'">
              {{ t('quota.edit_mount') }}
            </button>
            <button v-if="store.isAdmin" class="btn btn--small btn--danger" @click="confirmDelete(mount.id)" :disabled="backendStatus === 'error'">
              {{ t('quota.delete_mount') }}
            </button>
          </div>
        </div>

        <!-- Quota stats (when not editing) -->
        <template v-if="editingMountId !== mount.id">
          <div v-if="getMountCache(mount.id).quota" class="quota-stats">
            <div class="quota-stat">
              <span class="quota-stat__label">{{ t('quota.total') }}</span>
              <span class="quota-stat__value">{{ displayQuotaMB(getMountCache(mount.id).quota!.total, getMountCache(mount.id).mode) }}</span>
            </div>
            <div class="quota-stat">
              <span class="quota-stat__label">{{ t('quota.used') }}</span>
              <span class="quota-stat__value">{{ displayQuotaMB(getMountCache(mount.id).quota!.used, getMountCache(mount.id).mode) }}</span>
            </div>
            <div class="quota-stat">
              <span class="quota-stat__label">{{ t('quota.available') }}</span>
              <span class="quota-stat__value quota-stat__value--available">
                {{ displayQuotaMB(getMountCache(mount.id).quota!.available, getMountCache(mount.id).mode) }}
              </span>
            </div>
          </div>
          <div v-else class="quota-stats quota-stats--empty">
            <span class="quota-empty-text">{{ t('quota.empty_no_data') }}</span>
          </div>

          <!-- Progress bar -->
          <div v-if="getMountCache(mount.id).quota" class="quota-progress">
            <div
              class="quota-progress__bar"
              :style="{ width: `${Math.min((getMountCache(mount.id).quota!.used / getMountCache(mount.id).quota!.total) * 100, 100)}%` }"
            ></div>
          </div>

          <!-- Card footer with sync and update time -->
          <div class="mount-card__footer">
            <span class="quota-card__time">
              {{ t('quota.updated') }} {{ formatTime(getMountCache(mount.id).quota?.updated_at) }}
            </span>
            <button
              class="btn btn--primary"
              :disabled="getMountCache(mount.id).loading || backendStatus === 'error'"
              @click="handleSyncMount(mount.id)"
            >
              {{ getMountCache(mount.id).loading ? t('quota.syncing') : t('quota.sync_quota') }}
            </button>
          </div>
        </template>

        <!-- Edit form -->
        <div v-else class="edit-form">
          <div class="edit-field">
            <label>{{ t('quota.edit_name') }}</label>
            <input v-model="editForm.name" type="text" class="edit-input" />
          </div>

          <!-- Mode selector -->
          <div class="edit-field">
            <label>{{ t('quota.mode_title') }}</label>
            <div class="mode-options">
              <label class="mode-option">
                <input type="radio" v-model="editForm.quota_mode" value="real" />
                <div class="mode-option__body">
                  <span class="mode-option__title">{{ t('quota.real') }}</span>
                  <span class="mode-option__desc">{{ t('quota.real_desc') }}</span>
                </div>
              </label>
              <label class="mode-option">
                <input type="radio" v-model="editForm.quota_mode" value="virtual" />
                <div class="mode-option__body">
                  <span class="mode-option__title">{{ t('quota.virtual') }}</span>
                  <span class="mode-option__desc">{{ t('quota.virtual_desc') }}</span>
                </div>
              </label>
              <label class="mode-option">
                <input type="radio" v-model="editForm.quota_mode" value="inherit" />
                <div class="mode-option__body">
                  <span class="mode-option__title">{{ t('quota.inherit') }}</span>
                  <span class="mode-option__desc">{{ t('quota.inherit_desc') }}</span>
                </div>
              </label>
            </div>
          </div>

          <!-- Inherit parent selector -->
          <div v-if="editForm.quota_mode === 'inherit'" class="edit-field">
            <label>{{ t('quota.inherit_parent') }}</label>
            <div v-if="editAvailableParents.length === 0" class="mode-hint">
              {{ t('quota.inherit_no_parents') }}
            </div>
            <select v-else v-model.number="editForm.inherit_from_id" class="edit-input">
              <option :value="null" disabled>{{ t('quota.inherit_select_hint') }}</option>
              <option v-for="mp in editAvailableParents" :key="mp.id" :value="mp.id">
                {{ mp.providerName }}
              </option>
            </select>
          </div>

          <!-- Virtual fields -->
          <div v-if="editForm.quota_mode === 'virtual'" class="edit-field">
            <label>{{ t('quota.edit_virtual_total') }}</label>
            <input v-model.number="editForm.virtual_total" type="number" min="1" class="edit-input" />
          </div>
          <div v-if="editForm.quota_mode === 'virtual'" class="edit-field">
            <label>{{ t('quota.edit_virtual_used') }}</label>
            <input v-model.number="editForm.virtual_used" type="number" min="0" class="edit-input" />
          </div>

          <div class="edit-actions">
            <button class="btn btn--primary" @click="handleEditSave(mount.id)">{{ t('quota.edit_save') }}</button>
            <button class="btn btn--secondary" @click="cancelEdit()">{{ t('quota.edit_cancel') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state: no provider -->
    <div v-else-if="!initialLoading && !selectedProvider" class="empty-state">
      <p>{{ t('quota.empty_no_provider') }}</p>
    </div>

    <!-- Empty state: no mounts -->
    <div v-else-if="!initialLoading && selectedProvider && providerMounts.length === 0 && !showCreateForm" class="empty-state">
      <p>{{ t('quota.no_mounts') }}</p>
    </div>

    <!-- Delete confirmation dialog -->
    <teleport to="body">
      <div v-if="deleteConfirmId !== null" class="dialog-overlay" @mousedown.self="cancelDelete">
        <div class="dialog">
          <h3 class="dialog__title">{{ t('quota.delete_confirm_title') }}</h3>
          <p class="dialog__message">{{ t('quota.delete_confirm_message') }}</p>
          <div class="dialog__actions">
            <button class="btn btn--danger" @click="handleDelete(deleteConfirmId)">{{ t('quota.confirm_delete') }}</button>
            <button class="btn btn--secondary" @click="cancelDelete()">{{ t('quota.cancel') }}</button>
          </div>
        </div>
      </div>
    </teleport>
  </section>
</template>

<style scoped>
.quota-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 30px;
  padding: 20px;
  background: var(--surface);
  border-radius: 12px;
  border: 1px solid var(--border);
}

.provider-dropdown {
  position: relative;
  min-width: 240px;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  background: var(--surface);
  border: 1.5px solid var(--border);
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  color: var(--text);
  transition: all 0.2s;
  text-align: left;
  -webkit-appearance: none;
  appearance: none;
}

.dropdown-trigger:hover:not(.dropdown-trigger--disabled) {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.04);
}

.dropdown-trigger--disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dropdown-trigger__name {
  font-weight: 600;
  flex: 1;
}

.dropdown-trigger__placeholder {
  color: var(--muted);
  flex: 1;
}

.dropdown-arrow {
  flex-shrink: 0;
  color: var(--muted);
  transition: transform 0.25s ease;
}

.dropdown-arrow--open {
  transform: rotate(180deg);
}

.dropdown-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.12), 0 4px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  z-index: 100;
  max-height: 280px;
  overflow-y: auto;
  opacity: 1;
}

.dropdown-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  text-align: left;
  color: var(--text);
  transition: all 0.15s;
}

.dropdown-item:hover {
  background: rgba(59, 130, 246, 0.06);
}

.dropdown-item--active {
  background: rgba(59, 130, 246, 0.1);
  font-weight: 600;
}

.dropdown-item--active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 4px;
  bottom: 4px;
  width: 3px;
  background: #3b82f6;
  border-radius: 0 3px 3px 0;
}

.dropdown-item__info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.dropdown-item__name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
}

.dropdown-item__desc {
  font-size: 11px;
  color: var(--muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Provider type tags */
.provider-type-tag {
  font-size: 11px;
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

/* Dropdown transition */
.dropdown-fade-enter-active {
  transition: all 0.2s ease-out;
}
.dropdown-fade-leave-active {
  transition: all 0.15s ease-in;
}
.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

.button-group {
  display: flex;
  gap: 12px;
  margin-left: auto;
}

/* ── Buttons ── */
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
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
}
.btn--secondary:hover:not(:disabled) {
  background: var(--surface-strong);
}

.btn--danger {
  background: #dc2626;
  color: white;
}
.btn--danger:hover:not(:disabled) {
  background: #b91c1c;
}

.btn--small {
  padding: 6px 14px;
  font-size: 13px;
}

/* ── Mount list ── */
.mount-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.mount-card {
  background: var(--surface);
  border-radius: 12px;
  padding: 24px;
  border: 1px solid var(--border);
}

.mount-card--stale {
  opacity: 0.7;
}

.mount-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.mount-card__header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mount-card__header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.mount-card__title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.mount-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 16px;
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

[data-theme="dark"] .mode-badge--real { background: rgba(6,95,70,0.3); color: #6ee7b7; }
[data-theme="dark"] .mode-badge--inherit { background: rgba(30,64,175,0.3); color: #93c5fd; }
[data-theme="dark"] .mode-badge--virtual { background: rgba(146,64,14,0.3); color: #fcd34d; }

.inherit-chain {
  font-size: 12px;
  color: var(--muted);
  font-family: 'SFMono-Regular', Consolas, monospace;
}

/* ── Quota stats ── */
.quota-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}

.quota-stats--empty {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}

.quota-empty-text {
  color: var(--muted);
  font-size: 14px;
}

.quota-stat {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.quota-stat__label {
  font-size: 14px;
  color: var(--muted);
}

.quota-stat__value {
  font-size: 26px;
  font-weight: 600;
  color: var(--text);
}

.quota-stat__value--available {
  color: #10b981;
}

.quota-progress {
  height: 8px;
  background: var(--border);
  border-radius: 4px;
  overflow: hidden;
  margin-top: 16px;
}

.quota-progress__bar {
  height: 100%;
  background: #3b82f6;
  border-radius: 4px;
  transition: width 0.3s;
}

.quota-card__time {
  font-size: 13px;
  color: var(--muted);
}

/* ── Create mount form ── */
.mode-section {
  margin-bottom: 20px;
  padding: 20px;
  background: var(--surface);
  border-radius: 12px;
  border: 1px solid var(--border);
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
  border: 1px solid var(--border);
  border-radius: 10px;
  cursor: pointer;
  background: var(--surface);
  transition: all 0.2s;
}
.mode-option:has(input:checked) {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.08);
}
.mode-option:hover {
  border-color: var(--muted);
  background: var(--surface-strong);
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
  color: var(--text);
}

.mode-option__desc {
  font-size: 12px;
  color: var(--muted);
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
  color: var(--text);
  white-space: nowrap;
}

.virtual-input__field {
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  width: 240px;
  background: var(--surface);
  color: var(--text);
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
  color: var(--text);
  white-space: nowrap;
}
.inherit-select__field {
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  width: 240px;
  background: var(--surface);
  color: var(--text);
  transition: border-color 0.2s;
}
.inherit-select__field:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.mode-hint {
  margin-top: 16px;
  padding: 10px 16px;
  background: rgba(245, 158, 11, 0.12);
  color: #92400e;
  border-radius: 8px;
  font-size: 13px;
}

[data-theme="dark"] .mode-hint {
  color: #fbbf24;
}

.create-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

/* ── Edit form ── */
.edit-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 10px;
}

.edit-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.edit-field label {
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
}

.edit-input {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  background: var(--surface);
  color: var(--text);
  transition: border-color 0.2s;
}
.edit-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.edit-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 4px;
}

/* ── Status toast ── */
.status-toast {
  padding: 10px 16px;
  border-radius: 8px;
  background: #3b82f6;
  color: white;
  font-size: 14px;
  text-align: center;
  margin-bottom: 12px;
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

/* ── Empty state ── */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--muted);
  background: var(--surface);
  border-radius: 12px;
  border: 1px dashed var(--border);
}

.empty-state p {
  margin: 0;
}

/* ── Loading state ── */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--muted);
  gap: 16px;
}

.loading-spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── Offline banner ── */
.offline-banner {
  padding: 12px 20px;
  margin-bottom: 20px;
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
}

/* ── Delete confirmation dialog ── */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: var(--surface);
  border-radius: 14px;
  padding: 28px;
  min-width: 360px;
  max-width: 480px;
  border: 1px solid var(--border);
}

.dialog__title {
  margin: 0 0 12px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
}

.dialog__message {
  margin: 0 0 24px;
  font-size: 14px;
  color: var(--muted);
  line-height: 1.5;
}

.dialog__actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

/* ════════════════════════════════════════════
   Mobile: responsive QuotaView
   ════════════════════════════════════════════ */
@media (max-width: 860px) {
  .quota-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    padding: 14px;
    margin-bottom: 20px;
  }

  .provider-dropdown {
    min-width: 0;
  }

  .button-group {
    margin-left: 0;
  }

  .button-group .btn {
    width: 100%;
  }

  .mode-options {
    flex-direction: column;
    gap: 8px;
  }

  .mode-option {
    padding: 12px;
  }

  .mount-card {
    padding: 16px;
    border-radius: 10px;
  }

  .mount-card__header {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    margin-bottom: 16px;
  }

  .mount-card__header-left {
    flex-wrap: wrap;
  }

  .mount-card__header-right {
    width: 100%;
    justify-content: flex-start;
    gap: 8px;
  }

  .mount-card__title {
    font-size: 16px;
  }

  .quota-stats {
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .quota-stat__value {
    font-size: 20px;
  }

  .mount-card__footer {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .mount-card__footer .btn {
    width: 100%;
  }

  .virtual-input,
  .inherit-select {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }

  .virtual-input__field,
  .inherit-select__field {
    width: 100%;
  }

  .create-actions {
    justify-content: stretch;
  }
  .create-actions .btn {
    width: 100%;
  }

  .edit-actions {
    flex-direction: column;
    gap: 8px;
  }
  .edit-actions .btn {
    width: 100%;
  }

  .dialog {
    min-width: 0;
    width: calc(100% - 32px);
    max-width: 100%;
    padding: 20px;
    margin: 0 16px;
  }
}
</style>
