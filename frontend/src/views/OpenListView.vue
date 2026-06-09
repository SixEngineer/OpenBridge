<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '@/components/common/PageHeader.vue'
import DownloadDialog from '@/components/download/DownloadDialog.vue'
import { copyFiles, getFileInfo, getFiles, moveFiles, removeFiles, renameFile } from '@/api/storage'
import { useConsoleStore } from '@/stores/console'

const { t, locale } = useI18n()
const store = useConsoleStore()

const currentPath = ref('/')
const pathInput = ref('/')
const files = ref<OpenListFileItem[]>([])
const filesLoading = ref(false)
const filesError = ref('')
const contentProvider = ref('')
let fetchSequence = 0
const FILE_BROWSER_PAGE_SIZE = 200
const FILE_BROWSER_MAX_PAGES = 500

interface OpenListFileItem {
  name: string
  size: number
  is_dir: boolean
  modified: string
  created?: string
  sign?: string
  thumb?: string
  type?: number
  hashinfo?: string
  hash_info?: unknown
}

interface ClipboardState {
  mode: 'copy' | 'move'
  srcDir: string
  names: string[]
}

const selectedNames = ref<Set<string>>(new Set())
const clipboard = ref<ClipboardState | null>(null)
const fileActionLoading = ref(false)
const fileActionError = ref('')
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const fileDetail = ref<any | null>(null)

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
    if (field === 'size') {
      const aSize = Number(a.size) || 0
      const bSize = Number(b.size) || 0
      return order === 'asc' ? aSize - bSize : bSize - aSize
    }
    if (field === 'modified') {
      const aTime = Date.parse(a.modified) || 0
      const bTime = Date.parse(b.modified) || 0
      return order === 'asc' ? aTime - bTime : bTime - aTime
    }
    const va = a.name.toLowerCase()
    const vb = b.name.toLowerCase()
    return order === 'asc' ? va.localeCompare(vb) : vb.localeCompare(va)
  })
  return list
})

const selectedFiles = computed(() =>
  files.value.filter(file => selectedNames.value.has(file.name))
)

const selectedCount = computed(() => selectedNames.value.size)

const isAllSelected = computed(() =>
  sortedFiles.value.length > 0 && selectedNames.value.size === sortedFiles.value.length
)

const canPaste = computed(() =>
  clipboard.value !== null && clipboard.value.names.length > 0
)

const clipboardLabel = computed(() => {
  if (!clipboard.value) return ''
  const action = clipboard.value.mode === 'copy' ? t('openlist.copy') : t('openlist.cut')
  return `${action}: ${clipboard.value.names.join(', ')}`
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
const downloadItemIsDir = ref(false)
const downloadItemName = ref('')

function openDownload(file: any) {
  const fullPath = (currentPath.value.replace(/\/+$/, '') + '/' + file.name).replace(/\/+/g, '/')
  downloadFilePath.value = fullPath
  downloadItemIsDir.value = Boolean(file.is_dir)
  downloadItemName.value = file.name
  downloadDialogVisible.value = true
}

function onDownloadSuccess(taskId: string) {
  // Dialog stays open to show success, user closes manually
}

function resetBrowser(path = '/') {
  currentPath.value = path
  pathInput.value = path
  files.value = []
  filesError.value = ''
  fileActionError.value = ''
  selectedNames.value = new Set()
  detailVisible.value = false
  fileDetail.value = null
  contentProvider.value = ''
}

async function fetchFiles(path = pathInput.value) {
  const sequence = ++fetchSequence
  filesLoading.value = true
  filesError.value = ''
  fileActionError.value = ''
  selectedNames.value = new Set()
  currentPath.value = path
  pathInput.value = path
  contentProvider.value = ''
  try {
    const collected: any[] = []
    let page = 1
    let total = 0
    let provider = ''

    while (page <= FILE_BROWSER_MAX_PAGES) {
      const res = await getFiles({ path, page, per_page: FILE_BROWSER_PAGE_SIZE }, { timeout: 0 })
      if (sequence !== fetchSequence) return
      if (res.code !== 1000) {
        throw new Error(res.msg || 'Failed to load files')
      }

      const data = res.data
      const content = Array.isArray(data?.content) ? data.content : []
      collected.push(...content)

      if (page === 1) {
        total = Number(data?.total) || 0
        provider = data?.provider || ''
        contentProvider.value = provider
      }

      files.value = [...collected]

      if (content.length === 0) break
      if (total > 0 && collected.length >= total) break
      if (content.length < FILE_BROWSER_PAGE_SIZE) break
      page += 1
    }

    if (sequence !== fetchSequence) return
    files.value = collected
    contentProvider.value = provider
  } catch (e: any) {
    if (sequence !== fetchSequence) return
    filesError.value = e?.message || 'Failed to load files'
    files.value = []
  } finally {
    if (sequence === fetchSequence) {
      filesLoading.value = false
    }
  }
}

function navigateTo(path: string) {
  void fetchFiles(path)
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

function joinPath(parent: string, name: string): string {
  const base = parent === '/' ? '' : parent.replace(/\/+$/, '')
  return `${base}/${name}`.replace(/\/+/g, '/')
}

function toggleSelect(name: string) {
  const next = new Set(selectedNames.value)
  if (next.has(name)) {
    next.delete(name)
  } else {
    next.add(name)
  }
  selectedNames.value = next
}

function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedNames.value = new Set()
  } else {
    selectedNames.value = new Set(sortedFiles.value.map(file => file.name))
  }
}

function clearSelection() {
  selectedNames.value = new Set()
}

function selectOnly(name: string) {
  selectedNames.value = new Set([name])
}

function selectedNamesList(): string[] {
  return [...selectedNames.value]
}

async function refreshCurrentPath() {
  await fetchFiles(currentPath.value)
}

async function handleDeleteSelected() {
  const names = selectedNamesList()
  if (names.length === 0) return
  if (!window.confirm(t('openlist.delete_confirm', { count: names.length }))) return

  fileActionLoading.value = true
  fileActionError.value = ''
  try {
    const res = await removeFiles({ dir: currentPath.value, names })
    if (res.code !== 1000) {
      throw new Error(res.msg || t('openlist.operation_failed'))
    }
    clearSelection()
    await refreshCurrentPath()
  } catch (e: any) {
    fileActionError.value = e?.message || t('common.request_error')
  } finally {
    fileActionLoading.value = false
  }
}

function handleCopySelected() {
  const names = selectedNamesList()
  if (names.length === 0) return
  clipboard.value = { mode: 'copy', srcDir: currentPath.value, names }
}

function handleCutSelected() {
  const names = selectedNamesList()
  if (names.length === 0) return
  clipboard.value = { mode: 'move', srcDir: currentPath.value, names }
}

async function handlePaste() {
  const pending = clipboard.value
  if (!pending || pending.names.length === 0) return

  fileActionLoading.value = true
  fileActionError.value = ''
  try {
    const payload = {
      src_dir: pending.srcDir,
      dst_dir: currentPath.value,
      names: pending.names,
    }
    const res = pending.mode === 'copy'
      ? await copyFiles(payload)
      : await moveFiles(payload)
    if (res.code !== 1000) {
      throw new Error(res.msg || t('openlist.operation_failed'))
    }
    if (pending.mode === 'move') {
      clipboard.value = null
    }
    clearSelection()
    await refreshCurrentPath()
  } catch (e: any) {
    fileActionError.value = e?.message || t('common.request_error')
  } finally {
    fileActionLoading.value = false
  }
}

async function handleRename(file?: OpenListFileItem) {
  const target = file ?? selectedFiles.value[0]
  if (!target) return
  const nextName = window.prompt(t('openlist.rename_prompt'), target.name)
  if (!nextName || nextName.trim() === target.name) return

  fileActionLoading.value = true
  fileActionError.value = ''
  try {
    const res = await renameFile({
      path: joinPath(currentPath.value, target.name),
      name: nextName.trim(),
    })
    if (res.code !== 1000) {
      throw new Error(res.msg || t('openlist.operation_failed'))
    }
    clearSelection()
    await refreshCurrentPath()
  } catch (e: any) {
    fileActionError.value = e?.message || t('common.request_error')
  } finally {
    fileActionLoading.value = false
  }
}

async function handleShowDetails(file?: OpenListFileItem) {
  const target = file ?? selectedFiles.value[0]
  if (!target) return

  detailVisible.value = true
  detailLoading.value = true
  detailError.value = ''
  fileDetail.value = null
  try {
    const res = await getFileInfo(joinPath(currentPath.value, target.name))
    if (res.code !== 1000) {
      throw new Error(res.msg || t('openlist.operation_failed'))
    }
    fileDetail.value = res.data
  } catch (e: any) {
    detailError.value = e?.message || t('common.request_error')
  } finally {
    detailLoading.value = false
  }
}

function fileIcon(file: OpenListFileItem): string {
  if (file.is_dir) return 'folder'
  const ext = file.name.split('.').pop()?.toLowerCase() || ''
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg', 'heic'].includes(ext)) return 'image'
  if (['mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm'].includes(ext)) return 'video'
  if (['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'].includes(ext)) return 'audio'
  if (['zip', 'rar', '7z', 'tar', 'gz', 'bz2', 'xz'].includes(ext)) return 'archive'
  if (['pdf'].includes(ext)) return 'pdf'
  if (['doc', 'docx', 'ppt', 'pptx', 'xls', 'xlsx'].includes(ext)) return 'office'
  if (['go', 'ts', 'js', 'vue', 'json', 'html', 'css', 'md', 'py', 'java', 'cpp', 'c', 'rs'].includes(ext)) return 'code'
  return 'file'
}

function fileIconText(file: OpenListFileItem): string {
  const map: Record<string, string> = {
    folder: 'DIR',
    image: 'IMG',
    video: 'VID',
    audio: 'AUD',
    archive: 'ZIP',
    pdf: 'PDF',
    office: 'DOC',
    code: 'DEV',
    file: 'FILE',
  }
  return map[fileIcon(file)] || 'FILE'
}

function formatSize(bytes: number): string {
  if (!bytes || bytes === 0) return '—'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
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

onMounted(() => {
  if (store.isLoggedIn) {
    void fetchFiles('/')
  }
})

watch(() => store.isLoggedIn, (loggedIn) => {
  if (!loggedIn) {
    fetchSequence += 1
    filesLoading.value = false
    resetBrowser('/')
  }
})

watch(() => store.openListSessionKey, (sessionKey, previousKey) => {
  if (!sessionKey || sessionKey === previousKey) return
  resetBrowser('/')
  void fetchFiles('/')
})
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

      <div class="file-toolbar">
        <div class="file-toolbar__left">
          <span class="selection-pill">{{ t('openlist.selected_count', { count: selectedCount }) }}</span>
          <span v-if="clipboard" class="clipboard-pill">{{ clipboardLabel }}</span>
        </div>
        <div class="file-toolbar__actions">
          <button class="btn--sm" :disabled="selectedCount === 0 || fileActionLoading" @click="handleCopySelected">{{ t('openlist.copy') }}</button>
          <button class="btn--sm" :disabled="selectedCount === 0 || fileActionLoading" @click="handleCutSelected">{{ t('openlist.cut') }}</button>
          <button class="btn--sm" :disabled="!canPaste || fileActionLoading" @click="handlePaste">{{ t('openlist.paste') }}</button>
          <button class="btn--sm" :disabled="selectedCount !== 1 || fileActionLoading" @click="handleRename()">{{ t('openlist.rename') }}</button>
          <button class="btn--sm" :disabled="selectedCount !== 1 || fileActionLoading" @click="handleShowDetails()">{{ t('openlist.details') }}</button>
          <button class="btn--sm btn--danger" :disabled="selectedCount === 0 || fileActionLoading" @click="handleDeleteSelected">{{ t('openlist.delete') }}</button>
        </div>
      </div>

      <div v-if="filesError" class="msg msg--error">{{ filesError }}</div>
      <div v-if="fileActionError" class="msg msg--error">{{ fileActionError }}</div>

      <div v-if="filesLoading" class="loading-hint">{{ t('openlist.loading') }}</div>

      <div v-else-if="files.length > 0" class="file-table">
        <div class="file-table__head">
          <span class="file-table__checkbox">
            <input type="checkbox" :checked="isAllSelected" @click.stop="toggleSelectAll" />
          </span>
          <span class="sortable" @click="toggleSort('name')">{{ t('openlist.name') }}<span class="sort-arrow">{{ sortIndicator('name') }}</span></span>
          <span class="sortable" @click="toggleSort('size')">{{ t('openlist.size') }}<span class="sort-arrow">{{ sortIndicator('size') }}</span></span>
          <span class="sortable" @click="toggleSort('modified')">{{ t('openlist.modified') }}<span class="sort-arrow">{{ sortIndicator('modified') }}</span></span>
          <span>{{ t('openlist.action') }}</span>
        </div>
        <div
          v-for="f in sortedFiles"
          :key="f.name"
          class="file-row"
          :class="{ 'file-row--dir': f.is_dir, 'file-row--selected': selectedNames.has(f.name) }"
          @click="enterDir(f)"
        >
          <span class="file-row__checkbox" @click.stop>
            <input
              type="checkbox"
              :checked="selectedNames.has(f.name)"
              @change="toggleSelect(f.name)"
            />
          </span>
          <span class="file-row__name">
            <span class="file-icon" :class="`file-icon--${fileIcon(f)}`">{{ fileIconText(f) }}</span>
            {{ f.name }}
          </span>
          <span class="file-row__size">{{ f.is_dir ? '' : formatSize(f.size) }}</span>
          <span class="file-row__time">{{ formatTime(f.modified) }}</span>
          <span class="file-row__action" @click.stop>
            <button class="btn-icon" :title="t('openlist.details')" @click="handleShowDetails(f)">i</button>
            <button class="btn-icon" :title="t('openlist.rename')" @click="handleRename(f)">R</button>
            <button
              class="btn-download"
              @click="openDownload(f)"
            >
              {{ f.is_dir ? t('openlist.download_folder') : t('openlist.download') }}
            </button>
          </span>
        </div>
      </div>
      <p v-else-if="!filesError && !filesLoading" class="empty-hint">
        {{ t('openlist.empty') }} <code>/</code>.
      </p>

      <transition name="slide">
        <div v-if="detailVisible" class="detail-panel">
          <div class="detail-panel__header">
            <h4>{{ t('openlist.details') }}</h4>
            <button class="btn--sm" @click="detailVisible = false">{{ t('common.collapse') }}</button>
          </div>
          <div v-if="detailLoading" class="loading-hint">{{ t('openlist.detail_loading') }}</div>
          <div v-else-if="detailError" class="msg msg--error">{{ detailError }}</div>
          <div v-else-if="fileDetail" class="detail-grid">
            <div><span>{{ t('openlist.name') }}</span><strong>{{ fileDetail.name }}</strong></div>
            <div><span>{{ t('openlist.type') }}</span><strong>{{ fileDetail.is_dir ? t('openlist.folder') : t('openlist.file') }}</strong></div>
            <div><span>{{ t('openlist.size') }}</span><strong>{{ fileDetail.is_dir ? '—' : formatSize(fileDetail.size) }}</strong></div>
            <div><span>{{ t('openlist.modified') }}</span><strong>{{ formatTime(fileDetail.modified) }}</strong></div>
            <div><span>{{ t('openlist.created') }}</span><strong>{{ formatTime(fileDetail.created) }}</strong></div>
            <div><span>{{ t('openlist.provider') }}</span><strong>{{ fileDetail.provider || contentProvider || '—' }}</strong></div>
            <div class="detail-grid__wide"><span>{{ t('openlist.path') }}</span><strong>{{ joinPath(currentPath, fileDetail.name || '') }}</strong></div>
          </div>
        </div>
      </transition>
    </section>

    <!-- Download dialog -->
    <DownloadDialog
      :visible="downloadDialogVisible"
      :file-path="downloadFilePath"
      :is-dir="downloadItemIsDir"
      :item-name="downloadItemName"
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

.btn--danger {
  background: #fee2e2;
  color: #b91c1c;
}
.btn--danger:hover:not(:disabled) {
  background: #fecaca;
}

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
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.panel-path {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
  color: var(--muted);
}

.loading-hint {
  color: var(--muted);
  font-size: 14px;
}

.empty-hint {
  color: var(--muted);
  font-size: 14px;
  margin: 0;
}
.empty-hint code {
  background: var(--surface);
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 10px 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 16px;
  font-size: 14px;
  overflow-x: auto;
  white-space: nowrap;
}

.breadcrumb__link {
  color: var(--text);
  cursor: pointer;
  font-weight: 500;
  padding: 2px 4px;
  border-radius: 4px;
  transition: background 0.15s;
}
.breadcrumb__link:hover {
  background: var(--border);
}

.breadcrumb__sep {
  color: var(--muted);
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

.file-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
}

.file-toolbar__left,
.file-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.selection-pill,
.clipboard-pill {
  display: inline-flex;
  align-items: center;
  max-width: 360px;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: var(--muted);
  background: var(--surface-strong);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.clipboard-pill {
  color: #1e40af;
  background: rgba(59, 130, 246, 0.12);
}

.file-table {
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
}

.file-table__head {
  display: grid;
  grid-template-columns: 36px minmax(180px, 1fr) 100px 180px 160px;
  gap: 12px;
  padding: 10px 16px;
  background: var(--surface);
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border);
}

.file-table__checkbox,
.file-row__checkbox {
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-table__checkbox input,
.file-row__checkbox input {
  accent-color: #3b82f6;
}

.file-row {
  display: grid;
  grid-template-columns: 36px minmax(180px, 1fr) 100px 180px 160px;
  gap: 12px;
  padding: 10px 16px;
  font-size: 14px;
  color: var(--text);
  border-bottom: 1px solid var(--border);
  cursor: pointer;
  transition: background 0.15s;
}
.file-row:hover { background: var(--surface); }
.file-row--dir { font-weight: 500; color: var(--text); }
.file-row--selected { background: rgba(59, 130, 246, 0.08); }
.file-row:last-child { border-bottom: none; }

.file-row__name {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-icon {
  display: inline-grid;
  place-items: center;
  width: 34px;
  height: 24px;
  border-radius: 7px;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.02em;
  color: white;
  flex-shrink: 0;
  background: #64748b;
}
.file-icon--folder { background: linear-gradient(135deg, #f59e0b, #d97706); }
.file-icon--image { background: linear-gradient(135deg, #ec4899, #8b5cf6); }
.file-icon--video { background: linear-gradient(135deg, #ef4444, #f97316); }
.file-icon--audio { background: linear-gradient(135deg, #06b6d4, #2563eb); }
.file-icon--archive { background: linear-gradient(135deg, #78716c, #44403c); }
.file-icon--pdf { background: linear-gradient(135deg, #dc2626, #991b1b); }
.file-icon--office { background: linear-gradient(135deg, #2563eb, #1e40af); }
.file-icon--code { background: linear-gradient(135deg, #0f766e, #16a34a); }
.file-icon--file { background: linear-gradient(135deg, #64748b, #475569); }
.file-row__size { color: var(--muted); font-size: 13px; }
.file-row__time { color: var(--muted); font-size: 13px; }

.file-row__action {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
}

.btn-icon {
  width: 26px;
  height: 26px;
  border: 1px solid var(--border);
  border-radius: 7px;
  background: var(--surface);
  color: var(--muted);
  font-size: 11px;
  font-weight: 800;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-icon:hover {
  color: #2563eb;
  border-color: rgba(37, 99, 235, 0.35);
  background: rgba(59, 130, 246, 0.08);
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
.sortable:hover { color: var(--text); }

.sort-arrow {
  font-size: 11px;
}

.detail-panel {
  margin-top: 16px;
  padding: 16px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--surface);
}

.detail-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.detail-panel__header h4 {
  margin: 0;
  color: var(--text);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.detail-grid > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.detail-grid span {
  font-size: 12px;
  color: var(--muted);
}

.detail-grid strong {
  color: var(--text);
  font-size: 14px;
  word-break: break-all;
}

.detail-grid__wide {
  grid-column: 1 / -1;
}

.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* ════════════════════════════════════════════
   Mobile: table → card layout
   ════════════════════════════════════════════ */
@media (max-width: 860px) {
  .file-table__head {
    display: none;
  }

  .file-toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .file-toolbar__actions {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .file-row {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 12px;
    padding: 14px 16px;
    position: relative;
    align-items: center;
  }

  .file-row--dir {
    font-weight: 600;
  }

  .file-row__name {
    width: 100%;
    padding-right: 96px;
    padding-left: 28px;
    font-size: 15px;
    white-space: normal;
    word-break: break-word;
  }

  .file-row__time {
    font-size: 12px;
    color: var(--muted);
    order: 1;
  }

  .file-row__time::after {
    content: "·";
    margin: 0 6px;
    color: var(--border);
  }

  .file-row__size {
    font-size: 12px;
    color: var(--muted);
    order: 2;
  }

  /* 文件夹不显示大小 */
  .file-row--dir .file-row__size {
    display: none;
  }

  .file-row--dir .file-row__time::after {
    display: none;
  }

  .file-row__action {
    position: absolute;
    top: 14px;
    right: 16px;
  }

  .file-row__checkbox {
    position: absolute;
    top: 17px;
    left: 14px;
  }

  .btn-download {
    padding: 6px 14px;
    font-size: 13px;
  }

  .btn-icon {
    display: none;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }

  .breadcrumb {
    font-size: 13px;
    padding: 8px 12px;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .panel__header {
    flex-direction: column;
    gap: 6px;
  }

  .panel-path {
    display: none;
  }
}
</style>
