<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '@/components/common/PageHeader.vue'
import DownloadDialog from '@/components/download/DownloadDialog.vue'
import { getFiles } from '@/api/storage'

const { t, locale } = useI18n()

const currentPath = ref('/')
const pathInput = ref('/')
const files = ref<any[]>([])
const filesLoading = ref(false)
const filesError = ref('')
const contentProvider = ref('')

// Sort state
const sortField = ref<'name' | 'size' | 'modified'>('name')
const sortOrder = ref<'asc' | 'desc'>('desc')

const sortedFiles = computed(() => {
  const list = [...files.value]
  const field = sortField.value
  const order = sortOrder.value
  list.sort((a, b) => {
    // 按名称或修改时间排序时，文件夹排前面；按大小排序时，根据升降序决定文件夹位置
    if (field !== 'size' && a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1
    if (field === 'size' && a.is_dir !== b.is_dir) {
      return order === 'asc' ? (a.is_dir ? -1 : 1) : (a.is_dir ? 1 : -1)
    }
    let va = a[field]
    let vb = b[field]
    if (field === 'size') {
      va = va || 0
      vb = vb || 0
      return order === 'asc' ? va - vb : vb - va
    }
    // String fields
    va = (va || '').toString().toLowerCase()
    vb = (vb || '').toString().toLowerCase()
    return order === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va)
  })
  return list
})

function toggleSort(field: 'name' | 'size' | 'modified') {
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

// Download dialog
const downloadDialogVisible = ref(false)
const downloadFilePath = ref('')

function openDownload(file: any) {
  const fullPath = (currentPath.value.replace(/\/+$/, '') + '/' + file.name).replace(/\/+/g, '/')
  downloadFilePath.value = fullPath
  downloadDialogVisible.value = true
}

function onDownloadSuccess(taskId: string) {
  // Dialog stays open to show success, user closes manually
}

async function fetchFiles() {
  filesLoading.value = true
  filesError.value = ''
  currentPath.value = pathInput.value
  contentProvider.value = ''
  try {
    const res = await getFiles({ path: pathInput.value, page: 1, per_page: 50 })
    if (res.code === 1000) {
      files.value = res.data?.content || []
      contentProvider.value = res.data?.provider || ''
    }
  } catch (e: any) {
    filesError.value = e?.message || 'Failed to load files'
    files.value = []
  } finally {
    filesLoading.value = false
  }
}

function navigateTo(path: string) {
  pathInput.value = path
  fetchFiles()
}

const breadcrumbSegments = computed(() => {
  const parts = currentPath.value.replace(/\/+$/, '').split('/').filter(Boolean)
  const segments: { name: string; path: string; isLast: boolean }[] = []
  // Root segment
  segments.push({ name: t('openlist.root_name'), path: '/', isLast: parts.length === 0 })
  // Build up each sub-path
  let accumulated = ''
  parts.forEach((part, i) => {
    accumulated += '/' + part
    segments.push({ name: part, path: accumulated, isLast: i === parts.length - 1 })
  })
  return segments
})

function enterDir(item: any) {
  if (item.is_dir) {
    let newPath = currentPath.value.replace(/\/+$/, '') + '/' + item.name
    navigateTo(newPath)
  }
}

function formatSize(bytes: number): string {
  if (!bytes || bytes === 0) return '—'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatTime(t: string | number | Date): string {
  if (!t) return '—'
  try {
    return new Date(t).toLocaleString(locale.value)
  } catch {
    return String(t)
  }
}

onMounted(fetchFiles)
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('openlist.title')"
      :description="t('openlist.description')"
    />

    <section class="panel">
      <div class="panel__header">
        <h3>{{ t('openlist.file_browser') }}</h3>
        <span v-if="contentProvider && contentProvider !== 'unknown'" class="provider-tag">{{ contentProvider }}</span>
        <span v-else class="panel-path">{{ currentPath }}</span>
      </div>

      <nav class="breadcrumb">
        <template v-for="(seg, i) in breadcrumbSegments" :key="i">
          <span class="breadcrumb__link" @click="navigateTo(seg.path)">{{ seg.name }}</span>
          <span v-if="!seg.isLast" class="breadcrumb__sep">/</span>
        </template>
      </nav>

      <div v-if="filesError" class="msg msg--error">{{ filesError }}</div>

      <div v-if="filesLoading" class="loading-hint">{{ t('openlist.loading') }}</div>

      <div v-else-if="files.length > 0" class="file-table">
        <div class="file-table__head">
          <span class="sortable" @click="toggleSort('name')">{{ t('openlist.name') }}<span class="sort-arrow">{{ sortIndicator('name') }}</span></span>
          <span class="sortable" @click="toggleSort('size')">{{ t('openlist.size') }}<span class="sort-arrow">{{ sortIndicator('size') }}</span></span>
          <span class="sortable" @click="toggleSort('modified')">{{ t('openlist.modified') }}<span class="sort-arrow">{{ sortIndicator('modified') }}</span></span>
          <span>{{ t('openlist.action') }}</span>
        </div>
        <div
          v-for="f in sortedFiles"
          :key="f.name"
          class="file-row"
          :class="{ 'file-row--dir': f.is_dir }"
          @click="enterDir(f)"
        >
          <span class="file-row__name">
            <span class="file-icon">{{ f.is_dir ? '📁' : '📄' }}</span>
            {{ f.name }}
          </span>
          <span class="file-row__size">{{ f.is_dir ? '—' : formatSize(f.size) }}</span>
          <span class="file-row__time">{{ formatTime(f.modified) }}</span>
          <span class="file-row__action" @click.stop>
            <button
              v-if="!f.is_dir"
              class="btn-download"
              @click="openDownload(f)"
            >
              {{ t('openlist.download') }}
            </button>
          </span>
        </div>
      </div>
      <p v-else-if="!filesError && !filesLoading" class="empty-hint">
        {{ t('openlist.empty') }} <code>/</code>.
      </p>
    </section>

    <!-- Download dialog -->
    <DownloadDialog
      :visible="downloadDialogVisible"
      :file-path="downloadFilePath"
      @close="downloadDialogVisible = false"
      @success="onDownloadSuccess"
    />
  </section>
</template>

<style scoped>
.btn--sm {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
  background: #f3f4f6;
  color: #374151;
}
.btn--sm:hover:not(:disabled) { background: #e5e7eb; }
.btn--sm:disabled { opacity: 0.6; cursor: not-allowed; }

.btn--primary {
  background: #3b82f6;
  color: white;
}
.btn--primary:hover:not(:disabled) { background: #2563eb; }

.msg {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  margin-bottom: 12px;
}
.msg--error {
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #fecaca;
}

.panel-path {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  color: #6b7280;
}

.loading-hint {
  color: #6b7280;
  font-size: 14px;
}

.empty-hint {
  color: #9ca3af;
  font-size: 14px;
  margin: 0;
}
.empty-hint code {
  background: #f3f4f6;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 10px 16px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
  overflow-x: auto;
  white-space: nowrap;
}

.breadcrumb__link {
  color: #374151;
  cursor: pointer;
  font-weight: 500;
  padding: 2px 4px;
  border-radius: 4px;
  transition: background 0.15s;
}
.breadcrumb__link:hover {
  background: #e5e7eb;
}

.breadcrumb__sep {
  color: #9ca3af;
  margin: 0 2px;
}

.provider-tag {
  padding: 3px 10px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.file-table {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}

.file-table__head {
  display: grid;
  grid-template-columns: 1fr 100px 180px 80px;
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

.file-row {
  display: grid;
  grid-template-columns: 1fr 100px 180px 80px;
  gap: 12px;
  padding: 10px 16px;
  font-size: 14px;
  color: #374151;
  border-bottom: 1px solid #f3f4f6;
  cursor: pointer;
  transition: background 0.15s;
}
.file-row:hover { background: #f9fafb; }
.file-row--dir { font-weight: 500; color: #111827; }
.file-row:last-child { border-bottom: none; }

.file-row__name {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-icon { font-size: 14px; flex-shrink: 0; }
.file-row__size { color: #6b7280; font-size: 13px; }
.file-row__time { color: #6b7280; font-size: 13px; }

.file-row__action {
  display: flex;
  align-items: center;
}

.btn-download {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  background: #3b82f6;
  color: white;
  transition: background 0.2s;
}
.btn-download:hover { background: #2563eb; }

.sortable {
  cursor: pointer;
  user-select: none;
}
.sortable:hover { color: #374151; }

.sort-arrow {
  font-size: 11px;
}
</style>
