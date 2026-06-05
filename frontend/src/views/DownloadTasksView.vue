<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, onBeforeUnmount } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import { getTaskDetail } from '@/api/task'
import type { DownloadTask } from '@/types/download'

const { t } = useI18n()

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

// Active statuses (need continuous refresh)
const activeStatuses = ['pending', 'resolving', 'resolved', 'submitted', 'downloading']

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

// Manual refresh
async function handleRefresh() {
  await fetchAllTasks()
  if (hasActiveTasks.value) startAutoRefresh()
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
function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function statusLabel(s: string): string {
  const map: Record<string, string> = {
    pending: 'Pending',
    resolving: 'Resolving',
    resolved: 'Resolved',
    submitted: 'Submitted',
    downloading: 'Downloading',
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',
  }
  return map[s] || s
}

function formatTime(t: string | null | undefined): string {
  if (!t) return '—'
  return new Date(t).toLocaleString()
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
          v-for="st in ['all', ...activeStatuses, 'completed', 'failed', 'cancelled']"
          :key="st"
          class="tab-btn"
          :class="{ 'tab-btn--active': statusFilter === st }"
          @click="statusFilter = st"
        >
          {{ statusLabel(st) }}
          <span class="tab-count">({{ statusCounts[st] || 0 }})</span>
        </button>
      </div>
      <button class="btn btn--sm" :disabled="listLoading" @click="handleRefresh">
        {{ listLoading ? t('tasks.refreshing') : t('tasks.refresh') }}
      </button>
    </div>

    <div v-if="listLoading && tasks.length === 0" class="loading-hint">{{ t('tasks.loading') }}</div>
    <div v-if="listError" class="msg msg--error">{{ listError }}</div>

    <div v-if="filteredTasks.length > 0" class="task-table">
      <div class="task-table__head">
        <span>{{ t('tasks.file_name') }}</span>
        <span>{{ t('tasks.size_col') }}</span>
        <span>{{ t('tasks.status_col') }}</span>
        <span>{{ t('tasks.progress_col') }}</span>
        <span>{{ t('tasks.created_col') }}</span>
      </div>
      <div
        v-for="t in filteredTasks"
        :key="t.TaskID"
        class="task-row"
        :class="{ 'task-row--selected': selectedTaskId === t.TaskID }"
        @click="selectTask(t.TaskID)"
      >
        <span class="task-row__name" :title="t.FileName">{{ t.FileName || t.SourcePath }}</span>
        <span class="task-row__size">{{ formatBytes(t.FileSize) }}</span>
        <span>
          <span class="badge" :class="`badge--${t.Status}`">{{ statusLabel(t.Status) }}</span>
        </span>
        <span>
          <div class="progress-mini">
            <div
              class="progress-mini__fill"
              :style="{ width: `${t.Progress}%` }"
            ></div>
            <span class="progress-mini__text">{{ t.Progress.toFixed(0) }}%</span>
          </div>
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
      <button class="btn-stop-refresh" @click="stopAutoRefresh">Stop</button>
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
            <span class="detail-field__label">{{ t('tasks.progress') }}</span>
            <div class="progress-bar">
              <div
                class="progress-bar__fill"
                :style="{ width: `${selectedTask.Progress}%` }"
              ></div>
              <span class="progress-bar__text">{{ selectedTask.Progress.toFixed(1) }}%</span>
            </div>
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
            <span class="truncate">{{ selectedTask.DirectLink }}</span>
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

/* ── Task table ── */
.task-table {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 16px;
}

.task-table__head {
  display: grid;
  grid-template-columns: 1fr 90px 110px 120px 160px;
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

.task-row {
  display: grid;
  grid-template-columns: 1fr 90px 110px 120px 160px;
  gap: 12px;
  padding: 12px 16px;
  font-size: 14px;
  align-items: center;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: background 0.15s;
}
.task-row:hover { background: #f9fafb; }
.task-row--selected { background: #eff6ff; }
.task-row:last-child { border-bottom: none; }

.task-row__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #111827;
  font-weight: 500;
}

.task-row__size {
  color: #6b7280;
  font-size: 13px;
}

.task-row__time {
  color: #6b7280;
  font-size: 13px;
}

/* ── Progress bar (mini) ── */
.progress-mini {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 18px;
  background: #e5e7eb;
  border-radius: 9px;
  overflow: hidden;
  position: relative;
}

.progress-mini__fill {
  height: 100%;
  background: #3b82f6;
  border-radius: 9px;
  transition: width 0.3s;
}

.progress-mini__text {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 10px;
  font-weight: 600;
  color: #111827;
  text-shadow: 0 0 4px white;
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

.text--error { color: #dc2626; }

/* ── Progress bar (large) ── */
.progress-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  height: 20px;
  background: #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  position: relative;
}
.progress-bar__fill {
  height: 100%;
  background: #3b82f6;
  border-radius: 10px;
  transition: width 0.3s;
}
.progress-bar__text {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 11px;
  font-weight: 600;
  color: #111827;
  text-shadow: 0 0 4px white;
}

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

/* ── Panel ── */
.panel {
  background: white;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}
</style>
