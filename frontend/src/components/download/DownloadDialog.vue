<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LocalPathInput from '@/components/common/LocalPathInput.vue'
import { buildFolderZipUrl, createTask, resolveDirectLink } from '@/api/task'
import { getFiles } from '@/api/storage'
import { useConsoleStore } from '@/stores/console'
import type { DirectLinkResult } from '@/types/download'

interface StorageEntry {
  name: string
  size: number
  is_dir: boolean
}

interface CollectedFolderFile {
  path: string
  name: string
  size: number
  relativeDir: string
}

interface FolderScanResult {
  provider: string
  fileCount: number
  totalSize: number
  files: CollectedFolderFile[]
}

const store = useConsoleStore()
const { t } = useI18n()

const props = defineProps<{
  visible: boolean
  filePath: string
  isDir?: boolean
  itemName?: string
}>()

const emit = defineEmits<{ close: []; success: [taskId: string] }>()

const resolving = ref(false)
const resolveError = ref('')
const linkResult = ref<DirectLinkResult | null>(null)

const scanningFolder = ref(false)
const folderScanResult = ref<FolderScanResult | null>(null)

const copied = ref(false)
const manifestCopied = ref(false)
const directDownloading = ref(false)
const zipDownloading = ref(false)
const preparingManifest = ref(false)
const showQrCode = ref(false)

const manifestError = ref('')
const manifestText = ref('')
const manifestPreparedCount = ref(0)
const manifestFailedCount = ref(0)

const creating = ref(false)
const createError = ref('')
const createdTaskId = ref('')
const createdTaskIds = ref<string[]>([])
const createdTaskCount = ref(0)
const failedTaskCount = ref(0)

const targetDir = ref(store.defaultDownloadDir)

const itemIsDir = computed(() => Boolean(props.isDir))
const itemName = computed(() => props.itemName || baseName(props.filePath))
const effectiveTargetDir = computed(() => targetDir.value.trim() || store.defaultDownloadDir.trim())
const canSubmitFolder = computed(() => !itemIsDir.value || effectiveTargetDir.value !== '')
const dialogTitle = computed(() => itemIsDir.value ? t('download_dialog.title_folder') : t('download_dialog.title_file'))
const isMobileLikeDevice = computed(() => {
  if (typeof window === 'undefined') return false
  const ua = window.navigator.userAgent || ''
  const coarsePointer = typeof window.matchMedia === 'function' && window.matchMedia('(pointer: coarse)').matches
  return /android|iphone|ipad|ipod|mobile|tablet/i.test(ua) || coarsePointer
})
const showDirectDownloadButton = computed(() => !itemIsDir.value && Boolean(linkResult.value?.direct_link))
const terminalDirectMode = computed(() => !itemIsDir.value && Boolean(linkResult.value?.direct_link) && !linkResult.value?.is_openlist_proxy)
const canShowQrAction = computed(() => terminalDirectMode.value)
const isBaiduFolder = computed(() => /baidu/i.test(folderScanResult.value?.provider || ''))
const folderZipUrl = computed(() => itemIsDir.value ? buildFolderZipUrl(props.filePath) : '')
const qrCodeUrl = computed(() => {
  if (!linkResult.value?.direct_link || !canShowQrAction.value) return ''
  const url = new URL('https://api.qrserver.com/v1/create-qr-code/')
  url.searchParams.set('size', '280x280')
  url.searchParams.set('data', linkResult.value.direct_link)
  return url.toString()
})
const deviceNote = computed(() => {
  if (itemIsDir.value && isBaiduFolder.value) {
    return t('download_dialog.computer_note_folder_baidu')
  }
  if (itemIsDir.value) {
    return t('download_dialog.computer_note_folder')
  }
  if (terminalDirectMode.value) {
    return t('download_dialog.terminal_direct_mode')
  }
  return isMobileLikeDevice.value
    ? t('download_dialog.mobile_file_note')
    : t('download_dialog.computer_note_file')
})

watch(() => props.visible, (visible) => {
  if (visible && props.filePath) {
    document.body.style.overflow = 'hidden'
    initializeDialog()
    return
  }
  document.body.style.overflow = ''
})

function resetState() {
  resolving.value = false
  resolveError.value = ''
  linkResult.value = null
  scanningFolder.value = false
  folderScanResult.value = null
  copied.value = false
  manifestCopied.value = false
  directDownloading.value = false
  zipDownloading.value = false
  preparingManifest.value = false
  showQrCode.value = false
  manifestError.value = ''
  manifestText.value = ''
  manifestPreparedCount.value = 0
  manifestFailedCount.value = 0
  creating.value = false
  createError.value = ''
  createdTaskId.value = ''
  createdTaskIds.value = []
  createdTaskCount.value = 0
  failedTaskCount.value = 0
}

async function initializeDialog() {
  resetState()
  if (itemIsDir.value) {
    await scanFolder()
    return
  }
  await resolveLink()
}

function baseName(filePath: string): string {
  const normalized = filePath.replace(/\/+$/, '')
  const idx = normalized.lastIndexOf('/')
  return idx >= 0 ? normalized.slice(idx + 1) : normalized
}

function joinOpenListPath(parent: string, name: string): string {
  return `${parent.replace(/\/+$/, '')}/${name}`.replace(/\/+/g, '/')
}

function joinTargetDir(baseDir: string, ...parts: string[]): string {
  const separator = baseDir.includes('\\') && !baseDir.includes('/') ? '\\' : '/'
  const normalizedBase = baseDir.replace(/[\\/]+$/, '')
  const normalizedParts = parts
    .map(part => part.replace(/^[\\/]+|[\\/]+$/g, ''))
    .filter(Boolean)
  return [normalizedBase, ...normalizedParts].join(separator)
}

async function fetchFolderPage(filePath: string, page: number, perPage: number) {
  const res = await getFiles({ path: filePath, page, per_page: perPage })
  if (res.code !== 1000) {
    throw new Error((res.msg as string) || t('download_dialog.scan_failed'))
  }
  return res.data
}

async function listFolderEntries(filePath: string): Promise<{ provider: string; entries: StorageEntry[] }> {
  const perPage = 200
  let page = 1
  let total = 0
  const entries: StorageEntry[] = []
  let provider = ''

  while (true) {
    const data = await fetchFolderPage(filePath, page, perPage)
    provider = data?.provider || provider
    const content = (data?.content || []) as StorageEntry[]
    entries.push(...content)
    total = Number(data?.total || entries.length)
    if (entries.length >= total || content.length === 0) break
    page += 1
  }

  return { provider, entries }
}

async function collectFolderFiles(filePath: string, relativeDir = ''): Promise<FolderScanResult> {
  const { provider, entries } = await listFolderEntries(filePath)
  const files: CollectedFolderFile[] = []
  let totalSize = 0

  for (const entry of entries) {
    const fullPath = joinOpenListPath(filePath, entry.name)
    if (entry.is_dir) {
      const nested = await collectFolderFiles(fullPath, joinOpenListPath(relativeDir || '/', entry.name).replace(/^\//, ''))
      files.push(...nested.files)
      totalSize += nested.totalSize
      continue
    }

    files.push({
      path: fullPath,
      name: entry.name,
      size: entry.size || 0,
      relativeDir,
    })
    totalSize += entry.size || 0
  }

  return {
    provider,
    fileCount: files.length,
    totalSize,
    files,
  }
}

async function scanFolder() {
  scanningFolder.value = true
  resolveError.value = ''
  try {
    folderScanResult.value = await collectFolderFiles(props.filePath)
    if (!folderScanResult.value.fileCount) {
      resolveError.value = t('download_dialog.folder_empty')
    }
  } catch (error: any) {
    resolveError.value = error?.message || t('download_dialog.scan_failed')
  } finally {
    scanningFolder.value = false
  }
}

async function copyText(text: string, target: typeof copied | typeof manifestCopied) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }

  target.value = true
  window.setTimeout(() => {
    target.value = false
  }, 1800)
}

async function copyDirectLink() {
  if (!linkResult.value?.direct_link) return
  await copyText(linkResult.value.direct_link, copied)
}

async function copyManifest() {
  if (!manifestText.value) return
  await copyText(manifestText.value, manifestCopied)
}

async function resolveLink() {
  resolving.value = true
  resolveError.value = ''
  linkResult.value = null
  try {
    const res = await resolveDirectLink(props.filePath)
    if (res.code === 1000) {
      linkResult.value = res.data
      return
    }
    resolveError.value = (res.msg as string) || t('download_dialog.resolve_failed')
  } catch (error: any) {
    resolveError.value = error?.message || t('common.request_error')
  } finally {
    resolving.value = false
  }
}

function downloadToCurrentDevice() {
  if (!linkResult.value?.direct_link) return
  directDownloading.value = true

  const anchor = document.createElement('a')
  anchor.href = linkResult.value.direct_link
  anchor.target = '_blank'
  anchor.rel = 'noopener noreferrer'
  anchor.download = linkResult.value.name || ''
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)

  window.setTimeout(() => {
    directDownloading.value = false
  }, 1200)
}

function downloadFolderZip() {
  if (!folderZipUrl.value) return
  zipDownloading.value = true

  const anchor = document.createElement('a')
  anchor.href = folderZipUrl.value
  anchor.target = '_blank'
  anchor.rel = 'noopener noreferrer'
  anchor.download = `${itemName.value || 'folder'}.zip`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)

  window.setTimeout(() => {
    zipDownloading.value = false
  }, 1200)
}

async function buildDirectManifest() {
  if (!folderScanResult.value?.files.length) return
  preparingManifest.value = true
  manifestError.value = ''
  manifestText.value = ''
  manifestPreparedCount.value = 0
  manifestFailedCount.value = 0

  const lines: string[] = [
    '# OpenBridge direct link manifest',
    `# source=${props.filePath}`,
    `# generated_at=${new Date().toISOString()}`,
    '',
  ]

  for (const file of folderScanResult.value.files) {
    try {
      const res = await resolveDirectLink(file.path)
      const relativePath = [file.relativeDir, file.name].filter(Boolean).join('/')
      lines.push(`[${relativePath}]`)
      lines.push(`path=${file.path}`)
      lines.push(`url=${res.data.direct_link}`)
      lines.push(`proxy=${String(res.data.is_openlist_proxy)}`)
      if (res.data.header?.trim()) {
        lines.push(`header=${JSON.stringify(res.data.header)}`)
      }
      lines.push('')
      manifestPreparedCount.value += 1
    } catch (error: any) {
      manifestFailedCount.value += 1
      lines.push(`# failed path=${file.path} reason=${error?.message || 'resolve failed'}`)
    }
  }

  manifestText.value = lines.join('\n')
  if (!manifestPreparedCount.value) {
    manifestError.value = t('download_dialog.manifest_failed')
  } else if (manifestFailedCount.value) {
    manifestError.value = t('download_dialog.manifest_partial', {
      success: manifestPreparedCount.value,
      failed: manifestFailedCount.value,
    })
  }

  preparingManifest.value = false
}

function downloadManifest() {
  if (!manifestText.value) return
  const blob = new Blob([manifestText.value], { type: 'text/plain;charset=utf-8' })
  const objectURL = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = objectURL
  anchor.download = `${itemName.value || 'manifest'}-direct-links.txt`
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(objectURL)
}

async function createSingleTask(filePath: string, dir?: string) {
  const res = await createTask({
    path: filePath,
    dir: dir || undefined,
  })
  if (res.code !== 1000) {
    throw new Error((res.msg as string) || t('download_dialog.create_failed'))
  }
  const taskId = (res.data as any).task_id || (res.data as any).TaskID
  if (taskId) {
    store.recordTaskId(taskId)
  }
  return taskId || ''
}

async function handleConfirm() {
  if (!props.filePath) return
  creating.value = true
  createError.value = ''

  try {
    if (itemIsDir.value) {
      if (!canSubmitFolder.value) {
        createError.value = t('download_dialog.folder_requires_dir')
        return
      }
      if (!folderScanResult.value?.fileCount) {
        createError.value = t('download_dialog.folder_empty')
        return
      }

      const baseDir = effectiveTargetDir.value
      const rootFolder = itemName.value
      const createdIds: string[] = []
      const failures: string[] = []

      for (const file of folderScanResult.value.files) {
        const downloadDir = joinTargetDir(baseDir, rootFolder, file.relativeDir)
        try {
          const taskId = await createSingleTask(file.path, downloadDir)
          if (taskId) createdIds.push(taskId)
        } catch (error: any) {
          failures.push(`${file.path}: ${error?.message || t('download_dialog.create_failed')}`)
        }
      }

      createdTaskIds.value = createdIds
      createdTaskCount.value = createdIds.length
      failedTaskCount.value = failures.length
      if (createdIds[0]) {
        emit('success', createdIds[0])
      }
      if (failures.length) {
        createError.value = `${t('download_dialog.partial_failed')} ${failures.slice(0, 3).join(' | ')}`
      }
      return
    }

    const taskId = await createSingleTask(props.filePath, targetDir.value.trim())
    createdTaskId.value = taskId
    if (taskId) {
      emit('success', taskId)
    }
  } catch (error: any) {
    createError.value = error?.message || t('common.request_error')
  } finally {
    creating.value = false
  }
}

function handleClose() {
  showQrCode.value = false
  emit('close')
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '—'
  const unit = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
  const index = Math.floor(Math.log(bytes) / Math.log(unit))
  return parseFloat((bytes / Math.pow(unit, index)).toFixed(1)) + ' ' + sizes[index]
}
</script>

<template>
  <Teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="overlay" @mousedown.self="handleClose">
        <div class="dialog">
          <div class="dialog__header">
            <h3>{{ dialogTitle }}</h3>
            <button class="dialog__close" @click="handleClose">&times;</button>
          </div>

          <div class="dialog__body">
            <div class="note note--info">{{ deviceNote }}</div>

            <div class="info-row">
              <span class="info-row__label">{{ itemIsDir ? t('download_dialog.folder_path') : t('download_dialog.file_path') }}</span>
              <code class="info-row__value">{{ filePath }}</code>
            </div>

            <div v-if="itemIsDir" class="info-row">
              <span class="info-row__label">{{ t('download_dialog.folder_name') }}</span>
              <span class="info-row__value">{{ itemName }}</span>
            </div>

            <div v-if="resolving || scanningFolder" class="state-hint">
              {{ itemIsDir ? t('download_dialog.scanning_folder') : t('download_dialog.resolving') }}
            </div>
            <div v-else-if="resolveError" class="msg msg--error">{{ resolveError }}</div>

            <template v-else-if="itemIsDir && folderScanResult">
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.provider') }}</span>
                <span class="info-row__value">{{ folderScanResult.provider || '—' }}</span>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.folder_files') }}</span>
                <span class="info-row__value">{{ folderScanResult.fileCount }}</span>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.folder_total_size') }}</span>
                <span class="info-row__value">{{ formatBytes(folderScanResult.totalSize) }}</span>
              </div>

              <div v-if="isBaiduFolder" class="action-card">
                <div class="action-card__title">{{ t('download_dialog.folder_manifest_title') }}</div>
                <p class="action-card__desc">{{ t('download_dialog.folder_manifest_hint') }}</p>
                <div class="action-row">
                  <button class="btn btn--ghost" :disabled="preparingManifest" @click="buildDirectManifest">
                    {{ preparingManifest ? t('download_dialog.preparing_manifest') : t('download_dialog.prepare_manifest') }}
                  </button>
                  <button class="btn btn--secondary" :disabled="!manifestText" @click="copyManifest">
                    {{ manifestCopied ? t('download_dialog.copied') : t('download_dialog.copy_manifest') }}
                  </button>
                  <button class="btn btn--secondary" :disabled="!manifestText" @click="downloadManifest">
                    {{ t('download_dialog.download_manifest') }}
                  </button>
                </div>
                <div v-if="manifestText" class="note note--soft">
                  {{ t('download_dialog.manifest_ready', { success: manifestPreparedCount, failed: manifestFailedCount }) }}
                </div>
                <div v-if="manifestError" class="msg msg--error">{{ manifestError }}</div>
                <div v-if="manifestText" class="info-row" style="margin-top: 12px;">
                  <span class="info-row__label">{{ t('download_dialog.manifest_preview') }}</span>
                  <textarea class="manifest-box" :value="manifestText" readonly />
                </div>
              </div>

              <div v-else class="action-card">
                <div class="action-card__title">{{ t('download_dialog.folder_zip_title') }}</div>
                <p class="action-card__desc">{{ t('download_dialog.folder_zip_hint') }}</p>
                <button class="btn btn--ghost" :disabled="zipDownloading" @click="downloadFolderZip">
                  {{ zipDownloading ? t('download_dialog.zipping_now') : t('download_dialog.download_zip_current_device') }}
                </button>
              </div>

              <div class="note note--soft">{{ t('download_dialog.folder_structure_hint', { name: itemName }) }}</div>
            </template>

            <template v-else-if="linkResult">
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.file_name') }}</span>
                <span class="info-row__value">{{ linkResult.name }}</span>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.file_size') }}</span>
                <span class="info-row__value">{{ formatBytes(linkResult.size) }}</span>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.provider') }}</span>
                <span class="info-row__value">{{ linkResult.provider }}</span>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.direct_link') }}</span>
                <div class="link-row">
                  <code class="info-row__value truncate">{{ linkResult.direct_link }}</code>
                  <button class="btn btn--copy" @click="copyDirectLink">
                    {{ copied ? t('download_dialog.copied') : t('download_dialog.copy_link') }}
                  </button>
                </div>
              </div>
              <div class="info-row">
                <span class="info-row__label">{{ t('download_dialog.proxy') }}</span>
                <span class="badge" :class="linkResult.is_openlist_proxy ? 'badge--yes' : 'badge--no'">
                  {{ linkResult.is_openlist_proxy ? t('download_dialog.proxy_yes') : t('download_dialog.proxy_no') }}
                </span>
              </div>
              <div v-if="terminalDirectMode" class="action-card">
                <div class="action-card__title">{{ t('download_dialog.direct_actions_title') }}</div>
                <p class="action-card__desc">{{ t('download_dialog.direct_actions_hint') }}</p>
                <div class="action-row">
                  <button class="btn btn--ghost" :disabled="directDownloading" @click="downloadToCurrentDevice">
                    {{ directDownloading ? t('download_dialog.downloading_now') : t('download_dialog.baidu_direct_download') }}
                  </button>
                  <button class="btn btn--secondary" @click="copyDirectLink">
                    {{ copied ? t('download_dialog.copied') : t('download_dialog.copy_link') }}
                  </button>
                  <button v-if="canShowQrAction" class="btn btn--secondary" @click="showQrCode = true">
                    {{ t('download_dialog.show_qr') }}
                  </button>
                </div>
              </div>
            </template>

            <div v-if="(itemIsDir && folderScanResult) || (!itemIsDir && linkResult)" class="form-field" style="margin-top: 16px;">
              <label class="info-row__label">{{ t('download_dialog.download_dir') }}</label>
              <LocalPathInput
                v-model="targetDir"
                :placeholder="t('download_dialog.dir_placeholder')"
                mode="directory"
                :title="t('download_dialog.download_dir')"
              />
              <p v-if="itemIsDir && !canSubmitFolder" class="field-hint field-hint--error">
                {{ t('download_dialog.folder_requires_dir') }}
              </p>
            </div>

            <div v-if="createError" class="msg msg--error" style="margin-top: 12px;">
              {{ createError }}
            </div>

            <div v-if="createdTaskId" class="success-row">
              <span class="success-icon">&#10003;</span>
              <span>{{ t('download_dialog.task_created') }} <code>{{ createdTaskId }}</code></span>
            </div>

            <div v-if="itemIsDir && createdTaskCount" class="success-row">
              <span class="success-icon">&#10003;</span>
              <span>{{ t('download_dialog.folder_created', { count: createdTaskCount, failed: failedTaskCount }) }}</span>
            </div>
          </div>

          <div class="dialog__footer">
            <button class="btn btn--secondary" @click="handleClose" :disabled="creating">
              {{ createdTaskId || createdTaskCount ? t('download_dialog.close') : t('download_dialog.cancel') }}
            </button>
            <button
              v-if="showDirectDownloadButton && !terminalDirectMode"
              class="btn btn--ghost"
              :disabled="creating || resolving || directDownloading"
              @click="downloadToCurrentDevice"
            >
              {{ directDownloading ? t('download_dialog.downloading_now') : t('download_dialog.download_current_device') }}
            </button>
            <button
              v-if="((itemIsDir && folderScanResult && !createdTaskCount) || (!itemIsDir && linkResult && !createdTaskId))"
              class="btn btn--primary"
              :disabled="creating || resolving || scanningFolder || !canSubmitFolder"
              @click="handleConfirm"
            >
              {{
                creating
                  ? (itemIsDir ? t('download_dialog.submitting_folder') : t('download_dialog.submitting'))
                  : (itemIsDir ? t('download_dialog.submit_folder') : t('download_dialog.submit'))
              }}
            </button>
            <button
              v-if="createdTaskId || createdTaskCount"
              class="btn btn--primary"
              @click="handleClose"
            >
              {{ t('download_dialog.done') }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <transition name="fade">
      <div v-if="visible && showQrCode" class="overlay overlay--qr" @mousedown.self="showQrCode = false">
        <div class="qr-dialog">
          <div class="dialog__header">
            <h3>{{ t('download_dialog.qr_code') }}</h3>
            <button class="dialog__close" @click="showQrCode = false">&times;</button>
          </div>
          <div class="qr-dialog__body">
            <img v-if="qrCodeUrl" class="qr-image" :src="qrCodeUrl" :alt="t('download_dialog.qr_code')" />
            <p class="note note--soft">{{ t('download_dialog.qr_privacy_note') }}</p>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.overlay--qr {
  z-index: 1002;
}

.dialog {
  background: var(--surface-strong);
  border-radius: 12px;
  width: 640px;
  max-width: 94vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.qr-dialog {
  background: var(--surface-strong);
  border-radius: 12px;
  width: 360px;
  max-width: 92vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.18);
}

.dialog__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.dialog__header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
}

.dialog__close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
}

.dialog__close:hover {
  color: var(--text);
  background: var(--surface);
}

.dialog__body {
  padding: 20px 24px;
}

.dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 0 24px 20px;
  flex-wrap: wrap;
}

.qr-dialog__body {
  padding: 20px 24px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.qr-image {
  width: 280px;
  height: 280px;
  max-width: 100%;
  border-radius: 12px;
  background: white;
  padding: 8px;
}

.info-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.info-row__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-row__value {
  font-size: 14px;
  color: var(--text);
}

code.info-row__value {
  background: var(--surface);
  padding: 6px 10px;
  border-radius: 6px;
  font-size: 13px;
  word-break: break-all;
}

.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.link-row code {
  flex: 1;
  min-width: 0;
}

.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  width: fit-content;
}

.badge--yes {
  background: #fef3c7;
  color: #92400e;
}

.badge--no {
  background: #d1fae5;
  color: #065f46;
}

[data-theme="dark"] .badge--yes {
  background: rgba(146, 64, 14, 0.3);
  color: #fcd34d;
}

[data-theme="dark"] .badge--no {
  background: rgba(6, 95, 70, 0.3);
  color: #6ee7b7;
}

.state-hint {
  text-align: center;
  padding: 24px;
  color: var(--muted);
  font-size: 14px;
}

.msg {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
}

.msg--error {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.success-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(6, 95, 70, 0.1);
  color: #065f46;
  border-radius: 8px;
  font-size: 14px;
  margin-top: 12px;
}

[data-theme="dark"] .success-row {
  background: rgba(6, 95, 70, 0.25);
  color: #6ee7b7;
}

.success-icon {
  font-size: 18px;
  font-weight: 700;
}

.success-row code {
  font-weight: 600;
  background: rgba(0, 0, 0, 0.06);
  padding: 2px 8px;
  border-radius: 4px;
}

[data-theme="dark"] .success-row code {
  background: rgba(255, 255, 255, 0.1);
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
  background: var(--surface);
  color: var(--text);
  border: 1px solid var(--border);
}

.btn--secondary:hover:not(:disabled) {
  background: var(--surface-strong);
}

.btn--ghost {
  background: rgba(59, 130, 246, 0.08);
  color: #2563eb;
  border: 1px solid rgba(59, 130, 246, 0.22);
}

.btn--ghost:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.14);
}

.btn--copy {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  white-space: nowrap;
  transition: all 0.2s;
  flex-shrink: 0;
}

.btn--copy:hover {
  background: var(--surface-strong);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-hint {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.field-hint--error {
  color: #dc2626;
}

.note {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.5;
  margin-bottom: 14px;
}

.note--info {
  background: rgba(59, 130, 246, 0.08);
  color: #1d4ed8;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.note--soft {
  background: var(--surface);
  color: var(--muted);
}

.action-card {
  margin: 16px 0;
  padding: 16px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: linear-gradient(180deg, rgba(59, 130, 246, 0.06), rgba(59, 130, 246, 0.02));
}

.action-card__title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 6px;
}

.action-card__desc {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--muted);
}

.action-row {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.manifest-box {
  width: 100%;
  min-height: 180px;
  resize: vertical;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  padding: 12px;
  font: inherit;
  line-height: 1.5;
  box-sizing: border-box;
}

@media (max-width: 860px) {
  .overlay {
    align-items: flex-end;
  }

  .dialog {
    width: 100%;
    max-width: 100%;
    max-height: 92vh;
    border-radius: 16px 16px 0 0;
    overflow-y: auto;
    margin-top: auto;
  }

  .dialog__header {
    padding: 16px 20px 0;
    position: sticky;
    top: 0;
    background: var(--surface-strong);
    z-index: 1;
  }

  .dialog__body {
    padding: 16px 20px;
  }

  .dialog__footer {
    padding: 0 20px 20px;
    position: sticky;
    bottom: 0;
    background: var(--surface-strong);
  }

  .link-row {
    flex-direction: column;
    align-items: stretch;
  }

  .link-row code {
    word-break: break-all;
  }

  .btn--copy {
    align-self: flex-start;
  }

  .info-row {
    margin-bottom: 10px;
  }

  .action-row {
    flex-direction: column;
  }

  .action-row .btn {
    width: 100%;
  }
}
</style>
