<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, onBeforeUnmount } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import { getTaskDetail } from '@/api/task'
import type { DownloadTask } from '@/types/download'

const { t, locale } = useI18n()

const store = useConsoleStore()

// ── Create task ──
const sourcePath = ref('')
const targetDir = ref(store.defaultDownloadDir)
const creating = ref(false)
const createError = ref('')

async function handleCreate() {
  if (!sourcePath.value.trim()) return
  creating.value = true
  createError.value = ''
  try {
    const res = await store.addTask({
      path: sourcePath.value.trim(),
      dir: targetDir.value.trim() || undefined,
    })
    if (res.code !== 1000) {
      createError.value = (res.msg as string) || t('tasks.create_failed')
    } else {
      sourcePath.value = ''
      targetDir.value = ''
    }
  } catch (e: any) {
    createError.value = e?.message || t('common.request_error')
  } finally {
    creating.value = false
  }
}

// ── Task list ──
const tasks = ref<DownloadTask[]>([])
const listLoading = ref(false)
const listError = ref('')
const selectedTaskId = ref<string | null>(null)

const selectedTask = computed(() =>
  tasks.value.find(t => t.TaskID === selectedTaskId.value) ?? null
)

const statusFilter = ref<string>('all')

const filteredTasks = computed(() => {
  if (statusFilter.value === 'all') return tasks.value
  return tasks.value.filter(t => t.Status === statusFilter.value)
})

// Sort state
const sortField = ref<'FileName' | 'FileSize' | 'Status' | 'CreatedAt'>('CreatedAt')
const sortOrder = ref<'asc' | 'desc'>('desc')

const sortedTasks = computed(() => {
  const list = [...filteredTasks.value]
  const field = sortField.value
  const order = sortOrder.value
  list.sort((a, b) => {
    let va: any = (a as any)[field]
    let vb: any = (b as any)[field]
    if (field === 'FileSize') {
      va = va || 0
      vb = vb || 0
      return order === 'asc' ? va - vb : vb - va
    }
    if (field === 'CreatedAt') {
      va = va ? new Date(va).getTime() : 0
      vb = vb ? new Date(vb).getTime() : 0
      return order === 'asc' ? va - vb : vb - va
    }
    va = (va || '').toString().toLowerCase()
    vb = (vb || '').toString().toLowerCase()
    return order === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va)
  })
  return list
})

function toggleSort(field: 'FileName' | 'FileSize' | 'Status' | 'CreatedAt') {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortOrder.value = 'asc'
  }
}

function sortIndicator(field: string): string {
  if (sortField.value !== field) return ''
  return sortOrder.value === 'asc' ? ' ▲' : ' ▼'
}

// Active statuses (need continuous refresh)
const activeStatuses = ['submitted', 'downloading']

const hasActiveTasks = computed(() =>
  tasks.value.some(t => activeStatuses.includes(t.Status))
)

// Count per status
const statusCounts = computed(() => {
  const counts: Record<string, number> = { all: tasks.value.length }
  for (const t of tasks.value) {
    counts[t.Status] = (counts[t.Status] || 0) + 1
  }
  return counts
})

async function fetchAllTasks() {
  const ids = store.downloadTaskIds
  if (ids.length === 0) {
    tasks.value = []
    return
  }
  listLoading.value = true
  listError.value = ''
  const results: DownloadTask[] = []
  for (const id of ids) {
    try {
      const res = await getTaskDetail(id)
      if (res.code === 1000) {
        results.push(res.data)
      }
    } catch {
      // Skip individual failures
    }
  }
  tasks.value = results
  listLoading.value = false
}

function selectTask(taskId: string) {
  selectedTaskId.value = selectedTaskId.value === taskId ? null : taskId
}

// ── Auto refresh ──
let refreshTimer: ReturnType<typeof setInterval> | null = null
const refreshCount = ref(0)
const MAX_REFRESHES = 12 // stop after ~5 minutes of no progress
const autoRefreshActive = ref(true)

function startAutoRefresh() {
  stopAutoRefresh()
  refreshCount.value = 0
  autoRefreshActive.value = true
  refreshTimer = setInterval(() => {
    refreshCount.value++
    if (refreshCount.value > MAX_REFRESHES) {
      stopAutoRefresh()
      return
    }
    if (hasActiveTasks.value) {
      fetchAllTasks()
    } else {
      stopAutoRefresh()
    }
  }, 25000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  autoRefreshActive.value = false
}

function toggleAutoRefresh() {
  if (refreshTimer) {
    stopAutoRefresh()
  } else {
    startAutoRefresh()
  }
}

// Pause refresh when page is hidden
function onVisibilityChange() {
  if (document.hidden && refreshTimer) {
    stopAutoRefresh()
  } else if (!document.hidden && hasActiveTasks.value) {
    startAutoRefresh()
  }
}

function handleDelete(taskId: string) {
  tasks.value = tasks.value.filter(t => t.TaskID !== taskId)
  store.removeTaskId(taskId)
  const next = new Set(selectedTaskIds.value)
  next.delete(taskId)
  selectedTaskIds.value = next
  if (selectedTaskId.value === taskId) selectedTaskId.value = null
}

// ── Multi-select ──
const selectedTaskIds = ref<Set<string>>(new Set())

const isAllSelected = computed(() => {
  return sortedTasks.value.length > 0 && selectedTaskIds.value.size === sortedTasks.value.length
})

const hasSelection = computed(() => selectedTaskIds.value.size > 0)

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedTaskIds.value = new Set()
  } else {
    selectedTaskIds.value = new Set(sortedTasks.value.map(t => t.TaskID))
  }
}

function toggleSelect(taskId: string) {
  const next = new Set(selectedTaskIds.value)
  if (next.has(taskId)) {
    next.delete(taskId)
  } else {
    next.add(taskId)
  }
  selectedTaskIds.value = next
}

function clearSelected() {
  for (const id of selectedTaskIds.value) {
    tasks.value = tasks.value.filter(t => t.TaskID !== id)
    store.removeTaskId(id)
    if (selectedTaskId.value === id) selectedTaskId.value = null
  }
  selectedTaskIds.value = new Set()
}

// Clear selection when filter changes
watch(statusFilter, () => {
  selectedTaskIds.value = new Set()
})

// Clear completed tasks
function clearCompleted() {
  const completed = tasks.value.filter(t => t.Status === 'completed')
  for (const t of completed) {
    store.removeTaskId(t.TaskID)
  }
  tasks.value = tasks.value.filter(t => t.Status !== 'completed')
  selectedTaskIds.value = new Set()
  if (selectedTaskId.value && !tasks.value.find(t => t.TaskID === selectedTaskId.value)) {
    selectedTaskId.value = null
  }
}

onMounted(async () => {
  await fetchAllTasks()
  if (hasActiveTasks.value) startAutoRefresh()
  document.addEventListener('visibilitychange', onVisibilityChange)
})

onUnmounted(() => {
  stopAutoRefresh()
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
})

// ── Utilities ──
const copiedLinkId = ref<string | null>(null)

function copyDirectLink(link: string, taskId: string) {
  navigator.clipboard.writeText(link)
  copiedLinkId.value = taskId
  setTimeout(() => { copiedLinkId.value = null }, 1500)
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function statusLabel(s: string): string {
  const key = s === 'all' ? 'tasks.status_all' : `tasks.status_${s}`
  const translated = t(key)
  return translated !== key ? translated : s
}

function formatTime(t: string | null | undefined): string {
  if (!t) return '—'
  return new Date(t).toLocaleString(locale.value)
}
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('tasks.title')"
      :description="t('tasks.description')"
    />

    <section class="panel create-panel">
      <div class="create-form">
        <div class="create-form__field">
          <input
            v-model="sourcePath"
            :placeholder="t('tasks.source_path_placeholder')"
            @keyup.enter="handleCreate"
          />
        </div>
        <div class="create-form__field">
          <input
            v-model="targetDir"
            :placeholder="t('tasks.target_dir_placeholder')"
            @keyup.enter="handleCreate"
          />
        </div>
        <button
          class="btn btn--primary"
          :disabled="creating || !sourcePath.trim()"
          @click="handleCreate"
        >
          {{ creating ? t('tasks.creating') : t('tasks.create_task') }}
        </button>
      </div>
      <div v-if="createError" class="msg msg--error" style="margin-top: 12px;">
        {{ createError }}
      </div>
    </section>

    <div class="toolbar">
      <div class="toolbar__tabs">
        <button
          v-for="st in ['all', ...activeStatuses, 'completed']"
          :key="st"
          class="tab-btn"
          :class="{ 'tab-btn--active': statusFilter === st }"
          @click="statusFilter = st"
        >
          {{ statusLabel(st) }}
          <span class="tab-count">({{ statusCounts[st] || 0 }})</span>
        </button>
      </div>
<button
        v-if="statusCounts.completed > 0 || hasSelection"
        class="btn btn--sm btn--danger"
        @click="hasSelection ? clearSelected() : clearCompleted()"
      >{{ hasSelection ? `${t('tasks.clear_selected')} (${selectedTaskIds.size})` : t('tasks.clear_completed') }}</button>
    </div>

    <div v-if="listLoading && tasks.length === 0" class="loading-hint">{{ t('tasks.loading') }}</div>
    <div v-if="listError" class="msg msg--error">{{ listError }}</div>

    <div v-if="sortedTasks.length > 0" class="task-table">
      <div class="task-table__head">
        <span class="task-table__checkbox">
          <input type="checkbox" :checked="isAllSelected" @click.stop="toggleSelectAll" />
        </span>
        <span class="sortable" @click="toggleSort('FileName')">{{ t('tasks.file_name') }}<span class="sort-arrow">{{ sortIndicator('FileName') }}</span></span>
        <span class="sortable" @click="toggleSort('FileSize')">{{ t('tasks.size_col') }}<span class="sort-arrow">{{ sortIndicator('FileSize') }}</span></span>
        <span class="sortable" @click="toggleSort('Status')">{{ t('tasks.status_col') }}<span class="sort-arrow">{{ sortIndicator('Status') }}</span></span>
        <span class="sortable" @click="toggleSort('CreatedAt')">{{ t('tasks.created_col') }}<span class="sort-arrow">{{ sortIndicator('CreatedAt') }}</span></span>
      </div>
      <div
        v-for="t in sortedTasks"
        :key="t.TaskID"
        class="task-row"
        :class="{ 'task-row--selected': selectedTaskId === t.TaskID }"
        @click="selectTask(t.TaskID)"
      >
        <span class="task-row__checkbox">
          <input
            type="checkbox"
            :checked="selectedTaskIds.has(t.TaskID)"
            @click.stop="toggleSelect(t.TaskID)"
          />
        </span>
        <span class="task-row__name" :title="t.FileName">
          <span class="task-row__name-text">{{ t.FileName || t.SourcePath }}</span>
          <span class="task-row__delete" @click.stop="handleDelete(t.TaskID)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
          </span>
        </span>
        <span class="task-row__size">{{ formatBytes(t.FileSize) }}</span>
        <span>
          <span class="badge" :class="`badge--${t.Status}`">{{ statusLabel(t.Status) }}</span>
        </span>
        <span class="task-row__time">{{ formatTime(t.CreatedAt) }}</span>
      </div>
    </div>

    <div v-if="!listLoading && filteredTasks.length === 0 && store.downloadTaskIds.length === 0" class="empty-state">
      <p>{{ t('tasks.empty_no_tasks') }}</p>
    </div>
    <div v-else-if="!listLoading && filteredTasks.length === 0" class="empty-state">
      <p>{{ t('tasks.empty_no_match') }}</p>
    </div>

    <div v-if="autoRefreshActive" class="live-hint">
      {{ t('tasks.auto_refresh') }} ({{ Math.max(MAX_REFRESHES - refreshCount, 0) }} left)
      <button class="btn-stop-refresh" @click="stopAutoRefresh">{{ t('tasks.stop_refresh') }}</button>
    </div>

    <transition name="slide">
      <div v-if="selectedTask" class="panel detail-panel">
        <div class="detail-panel__header">
          <h3>{{ t('tasks.task_details') }}</h3>
          <button class="btn btn--sm" @click="selectedTaskId = null">{{ t('tasks.close') }}</button>
        </div>
        <div class="detail-grid">
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.task_id') }}</span>
            <span class="mono">{{ selectedTask.TaskID }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.source_path') }}</span>
            <span>{{ selectedTask.SourcePath }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.file_name_detail') }}</span>
            <span>{{ selectedTask.FileName || '—' }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.file_size') }}</span>
            <span>{{ formatBytes(selectedTask.FileSize) }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.aria2_gid') }}</span>
            <span v-if="selectedTask.Aria2GID" class="mono">{{ selectedTask.Aria2GID }}</span>
            <span v-else>—</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.status') }}</span>
            <span class="badge" :class="`badge--${selectedTask.Status}`">{{ statusLabel(selectedTask.Status) }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.retry_count') }}</span>
            <span>{{ selectedTask.RetryCount }}</span>
          </div>
          <div class="detail-field" v-if="selectedTask.ErrorMessage">
            <span class="detail-field__label">{{ t('tasks.error') }}</span>
            <span class="text--error">{{ selectedTask.ErrorMessage }}</span>
          </div>
          <div class="detail-field" v-if="selectedTask.DirectLink">
            <span class="detail-field__label">{{ t('tasks.direct_link') }}</span>
            <div class="direct-link-row">
              <button class="btn btn--copy" @click.stop="copyDirectLink(selectedTask.DirectLink, selectedTask.TaskID)">
                {{ copiedLinkId === selectedTask.TaskID ? '已复制!' : t('tasks.copy_link') }}
              </button>
              <code class="direct-link-value">{{ selectedTask.DirectLink }}</code>
            </div>
          </div>
          <div class="detail-field" v-if="selectedTask.StartedAt">
            <span class="detail-field__label">{{ t('tasks.started') }}</span>
            <span>{{ formatTime(selectedTask.StartedAt) }}</span>
          </div>
          <div class="detail-field" v-if="selectedTask.FinishedAt">
            <span class="detail-field__label">{{ t('tasks.finished') }}</span>
            <span>{{ formatTime(selectedTask.FinishedAt) }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.created') }}</span>
            <span>{{ formatTime(selectedTask.CreatedAt) }}</span>
          </div>
          <div class="detail-field">
            <span class="detail-field__label">{{ t('tasks.updated') }}</span>
            <span>{{ formatTime(selectedTask.UpdatedAt) }}</span>
          </div>
        </div>
      </div>
    </transition>
  </section>
</template>

<style scoped>
/* ── Create task (compact) ── */
.create-panel {
  margin-bottom: 16px;
}

.create-form {
  display: flex;
  gap: 10px;
  align-items: center;
}

.create-form__field {
  flex: 1;
}

.create-form__field input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s;
}
.create-form__field input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

/* ── Toolbar ── */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  gap: 12px;
}

.toolbar__tabs {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.tab-btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid #e5e7eb;
  background: white;
  color: #6b7280;
  transition: all 0.2s;
}
.tab-btn:hover { background: #f9fafb; color: #374151; }
.tab-btn--active {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.tab-count {
  font-size: 11px;
  opacity: 0.75;
}

.btn--sm {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  transition: all 0.2s;
  white-space: nowrap;
}
.btn--sm:hover:not(:disabled) { background: #f9fafb; }
.btn--sm:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--danger {
  color: #dc2626;
  border-color: #fca5a5;
}
.btn--danger:hover:not(:disabled) {
  background: #fef2f2;
  border-color: #dc2626;
}

/* ── Task table ── */
.task-table {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 16px;
}

.task-table__head {
  display: grid;
  grid-template-columns: 36px 1fr 90px 110px 160px;
  gap: 12px;
  padding: 10px 16px;
  background: #f9fafb;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e5e7eb;
}

.task-table__checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
}
.task-table__checkbox input[type="checkbox"] {
  accent-color: #bfdbfe;
}

.task-row {
  display: grid;
  grid-template-columns: 36px 1fr 90px 110px 160px;
  gap: 12px;
  padding: 12px 16px;
  font-size: 14px;
  align-items: center;
  border-bottom: 1px solid #f3f4f6;
  transition: background 0.15s;
}

.task-row__checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
}
.task-row__checkbox input[type="checkbox"] {
  opacity: 0;
  transition: opacity 0.15s;
  accent-color: #bfdbfe;
}
.task-row__checkbox input[type="checkbox"]:checked {
  opacity: 1;
}
.task-row:hover .task-row__checkbox input[type="checkbox"] {
  opacity: 1;
}
.task-row:hover { background: #f9fafb; }
.task-row--selected { background: #eff6ff; }
.task-row:last-child { border-bottom: none; }

.task-row__name {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  color: #111827;
  font-weight: 500;
}
.task-row__name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}
.task-row__delete {
  flex-shrink: 0;
  display: none;
  cursor: pointer;
  color: #9ca3af;
  transition: color 0.15s;
  line-height: 1;
}
.task-row:hover .task-row__delete {
  display: inline-flex;
}
.task-row__delete:hover {
  color: #dc2626;
}

.task-row__size {
  color: #6b7280;
  font-size: 13px;
}

.task-row__time {
  color: #6b7280;
  font-size: 13px;
}

/* ── Status badges ── */
.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}
.badge--pending,
.badge--submitted { background: #fef3c7; color: #92400e; }
.badge--resolving { background: #dbeafe; color: #1e40af; }
.badge--resolved,
.badge--downloading { background: #d1fae5; color: #065f46; }
.badge--completed { background: #10b981; color: white; }
.badge--failed { background: #fef2f2; color: #dc2626; }
.badge--cancelled { background: #f3f4f6; color: #6b7280; }

/* ── Hints ── */
.loading-hint {
  color: #6b7280;
  font-size: 14px;
  padding: 24px;
  text-align: center;
}

.live-hint {
  text-align: center;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.btn-stop-refresh {
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #6b7280;
  transition: all 0.2s;
}
.btn-stop-refresh:hover {
  background: #fee2e2;
  border-color: #ef4444;
  color: #dc2626;
}

.msg {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 12px;
}
.msg--error { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
  background: #f9fafb;
  border-radius: 12px;
  border: 1px dashed #d1d5db;
}

/* ── Detail panel ── */
.detail-panel {
  margin-bottom: 16px;
}

.detail-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.detail-panel__header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.detail-field__label {
  font-size: 12px;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-field span:last-child {
  font-size: 14px;
  color: #111827;
}

.mono {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 13px;
}

.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.direct-link-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.direct-link-value {
  flex: 1;
  overflow-x: auto;
  white-space: nowrap;
  padding: 6px 10px;
  background: #f3f4f6;
  border-radius: 6px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  color: #374151;
  scrollbar-width: thin;
}

.btn--copy {
  padding: 4px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  white-space: nowrap;
  transition: all 0.2s;
  flex-shrink: 0;
}
.btn--copy:hover {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.text--error { color: #dc2626; }

/* ── Slide animation ── */
.slide-enter-active, .slide-leave-active {
  transition: all 0.25s ease;
}
.slide-enter-from, .slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
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
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--primary { background: #3b82f6; color: white; }
.btn--primary:hover:not(:disabled) { background: #2563eb; }

/* ── Sortable column headers ── */
.sortable {
  cursor: pointer;
  user-select: none;
}
.sortable:hover { color: #374151; }

.sort-arrow {
  font-size: 11px;
}

/* ── Panel ── */
.panel {
  background: white;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}
</style>
