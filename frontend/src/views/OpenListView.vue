<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import DownloadDialog from '@/components/download/DownloadDialog.vue'
import { getFiles } from '@/api/storage'

const currentPath = ref('/')
const pathInput = ref('/')
const files = ref<any[]>([])
const filesLoading = ref(false)
const filesError = ref('')
const contentProvider = ref('')

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

function enterDir(item: any) {
  if (item.is_dir) {
    let newPath = currentPath.value.replace(/\/+$/, '') + '/' + item.name
    pathInput.value = newPath
    fetchFiles()
  }
}

function goUp() {
  const parts = currentPath.value.replace(/\/+$/, '').split('/')
  parts.pop()
  pathInput.value = parts.join('/') || '/'
  fetchFiles()
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
    return new Date(t).toLocaleString()
  } catch {
    return String(t)
  }
}

onMounted(fetchFiles)
</script>

<template>
  <section class="page">
    <PageHeader
      title="OpenList"
      description="Browse files from mounted cloud drives"
    />

    <section class="panel">
      <div class="panel__header">
        <h3>File Browser</h3>
        <span v-if="contentProvider" class="provider-tag">{{ contentProvider }}</span>
        <span v-else class="breadcrumb">{{ currentPath }}</span>
      </div>

      <div class="browser-controls">
        <div class="path-row">
          <button class="btn btn--sm" @click="goUp" :disabled="currentPath === '/'">Up</button>
          <input
            v-model="pathInput"
            class="path-input"
            placeholder="/"
            @keyup.enter="fetchFiles"
          />
          <button class="btn btn--primary btn--sm" :disabled="filesLoading" @click="fetchFiles">
            {{ filesLoading ? 'Loading...' : 'Go' }}
          </button>
          <button class="btn btn--sm" @click="pathInput = '/'; fetchFiles()">Root</button>
        </div>
      </div>

      <div v-if="filesError" class="msg msg--error">{{ filesError }}</div>

      <div v-if="filesLoading" class="loading-hint">Loading files...</div>

      <div v-else-if="files.length > 0" class="file-table">
        <div class="file-table__head">
          <span>Name</span>
          <span>Size</span>
          <span>Modified</span>
          <span>Action</span>
        </div>
        <div
          v-for="f in files"
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
              Download
            </button>
          </span>
        </div>
      </div>
      <p v-else-if="!filesError && !filesLoading" class="empty-hint">
        No files found at this path. If you have mounted drives in OpenList, they should appear at <code>/</code>.
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

.breadcrumb {
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

.browser-controls {
  margin-bottom: 16px;
}

.path-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.path-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  outline: none;
}

.path-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
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
</style>
