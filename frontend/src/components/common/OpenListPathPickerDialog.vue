<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getFiles } from '@/api/storage'

interface StorageEntry {
  name: string
  size: number
  is_dir: boolean
  modified: string
}

const { t, locale } = useI18n()

const props = withDefaults(defineProps<{
  visible: boolean
  initialPath?: string
}>(), {
  initialPath: '/',
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', path: string): void
}>()

const currentPath = ref('/')
const files = ref<StorageEntry[]>([])
const loading = ref(false)
const error = ref('')
const provider = ref('')

watch(() => props.visible, async (visible) => {
  if (!visible) return
  currentPath.value = normalizePath(props.initialPath || '/')
  await fetchFiles()
}, { immediate: true })

const breadcrumbSegments = computed(() => {
  const parts = currentPath.value.replace(/\/+$/, '').split('/').filter(Boolean)
  const segments: Array<{ name: string; path: string }> = [{ name: t('tasks.path_root'), path: '/' }]
  let accumulated = ''
  for (const part of parts) {
    accumulated += '/' + part
    segments.push({ name: part, path: accumulated })
  }
  return segments
})

function normalizePath(path: string) {
  if (!path.trim()) return '/'
  return ('/' + path.trim()).replace(/\/+/g, '/').replace(/\/+$/, '') || '/'
}

function joinPath(base: string, name: string) {
  return normalizePath(`${base}/${name}`)
}

async function fetchFiles() {
  loading.value = true
  error.value = ''
  provider.value = ''
  try {
    const res = await getFiles({ path: currentPath.value, page: 1, per_page: 200 }, { timeout: 0 })
    files.value = (res.data?.content || []) as StorageEntry[]
    provider.value = res.data?.provider || ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('common.request_error')
    files.value = []
  } finally {
    loading.value = false
  }
}

function openDirectory(entry: StorageEntry) {
  currentPath.value = joinPath(currentPath.value, entry.name)
  void fetchFiles()
}

function chooseCurrentFolder() {
  emit('select', currentPath.value)
}

function chooseFile(entry: StorageEntry) {
  emit('select', joinPath(currentPath.value, entry.name))
}

function navigateTo(path: string) {
  currentPath.value = path
  void fetchFiles()
}

function formatTime(value: string) {
  if (!value) return '—'
  return new Date(value).toLocaleString(locale.value)
}
</script>

<template>
  <div v-if="visible" class="dialog-overlay" @mousedown.self="emit('close')">
    <div class="dialog">
      <div class="dialog__header">
        <div>
          <h3>{{ t('tasks.source_browser_title') }}</h3>
          <p>{{ provider || currentPath }}</p>
        </div>
        <button class="dialog__close" @click="emit('close')">&times;</button>
      </div>

      <div class="dialog__body">
        <div class="picker-toolbar">
          <nav class="breadcrumb">
            <template v-for="(seg, index) in breadcrumbSegments" :key="seg.path">
              <button class="breadcrumb__link" @click="navigateTo(seg.path)">{{ seg.name }}</button>
              <span v-if="index < breadcrumbSegments.length - 1" class="breadcrumb__sep">/</span>
            </template>
          </nav>
          <button class="btn btn--primary" @click="chooseCurrentFolder">{{ t('tasks.select_this_folder') }}</button>
        </div>

        <div v-if="loading" class="state-hint">{{ t('tasks.source_loading') }}</div>
        <div v-else-if="error" class="msg msg--error">{{ error }}</div>
        <div v-else-if="files.length === 0" class="state-hint">{{ t('tasks.source_empty') }}</div>
        <div v-else class="file-list">
          <article v-for="entry in files" :key="entry.name" class="file-row">
            <button class="file-row__main" @click="entry.is_dir ? openDirectory(entry) : chooseFile(entry)">
              <span class="file-row__icon">{{ entry.is_dir ? 'DIR' : 'FILE' }}</span>
              <span class="file-row__name">{{ entry.name }}</span>
              <span class="file-row__time">{{ formatTime(entry.modified) }}</span>
            </button>
            <button
              v-if="!entry.is_dir"
              class="btn btn--secondary"
              @click="chooseFile(entry)"
            >
              {{ t('tasks.select_file') }}
            </button>
            <button
              v-else
              class="btn btn--secondary"
              @click="openDirectory(entry)"
            >
              {{ t('tasks.open_folder') }}
            </button>
          </article>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.dialog {
  width: min(920px, 94vw);
  max-height: 88vh;
  display: flex;
  flex-direction: column;
  background: var(--surface-strong);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.18);
}

.dialog__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
  border-bottom: 1px solid var(--border);
}

.dialog__header h3 {
  margin: 0 0 4px;
}

.dialog__header p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
}

.dialog__close {
  border: none;
  background: none;
  color: var(--muted);
  font-size: 24px;
  cursor: pointer;
}

.dialog__body {
  padding: 20px 22px;
  overflow-y: auto;
}

.picker-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.breadcrumb__link {
  border: none;
  background: var(--surface);
  color: var(--text);
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
}

.breadcrumb__sep {
  color: var(--muted);
}

.file-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.file-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: center;
  border: 1px solid var(--border);
  background: var(--surface);
  border-radius: 12px;
  padding: 10px;
}

.file-row__main {
  display: grid;
  grid-template-columns: 56px 1fr auto;
  gap: 12px;
  align-items: center;
  border: none;
  background: transparent;
  color: var(--text);
  cursor: pointer;
  text-align: left;
}

.file-row__icon {
  font-size: 11px;
  font-weight: 700;
  color: #3b82f6;
  letter-spacing: 0.06em;
}

.file-row__name {
  font-size: 14px;
  word-break: break-word;
}

.file-row__time,
.state-hint {
  font-size: 13px;
  color: var(--muted);
}

.msg {
  padding: 10px 16px;
  border-radius: 8px;
}

.msg--error {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.btn {
  padding: 9px 14px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
}

.btn--primary {
  background: #3b82f6;
  color: white;
}

.btn--secondary {
  background: var(--surface-strong);
  border: 1px solid var(--border);
  color: var(--text);
}

@media (max-width: 860px) {
  .dialog {
    width: 100%;
    max-height: 94vh;
    border-radius: 18px 18px 0 0;
    align-self: flex-end;
  }

  .picker-toolbar,
  .file-row,
  .file-row__main {
    grid-template-columns: 1fr;
    display: flex;
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
