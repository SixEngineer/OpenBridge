<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { resolveDirectLink, createTask } from '@/api/task'
import { useConsoleStore } from '@/stores/console'
import type { DirectLinkResult } from '@/types/download'

const store = useConsoleStore()
const { t } = useI18n()

const props = defineProps<{ visible: boolean; filePath: string }>()
const emit = defineEmits<{ close: []; success: [taskId: string] }>()

// Resolve state
const resolving = ref(false)
const resolveError = ref('')
const linkResult = ref<DirectLinkResult | null>(null)

// Copy state
const copied = ref(false)

// Create state
const creating = ref(false)
const createError = ref('')
const createdTaskId = ref('')

// Download directory (pre-filled from settings)
const targetDir = ref(store.defaultDownloadDir)

// Auto-resolve when dialog opens
watch(() => props.visible, (v) => {
  if (v && props.filePath) {
    resolveLink()
    copied.value = false
  }
})

async function copyDirectLink() {
  if (!linkResult.value?.direct_link) return
  try {
    await navigator.clipboard.writeText(linkResult.value.direct_link)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // fallback for insecure context
    const ta = document.createElement('textarea')
    ta.value = linkResult.value.direct_link
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }
}

async function resolveLink() {
  resolving.value = true
  resolveError.value = ''
  linkResult.value = null
  createdTaskId.value = ''
  createError.value = ''
  try {
    const res = await resolveDirectLink(props.filePath)
    if (res.code === 1000) {
      linkResult.value = res.data
    } else {
      resolveError.value = (res.msg as string) || t('download_dialog.resolve_failed')
    }
  } catch (e: any) {
    resolveError.value = e?.message || t('common.request_error')
  } finally {
    resolving.value = false
  }
}

async function handleConfirm() {
  if (!props.filePath) return
  creating.value = true
  createError.value = ''
  try {
    const res = await createTask({
      path: props.filePath,
      dir: targetDir.value.trim() || undefined,
    })
    if (res.code === 1000) {
      createdTaskId.value = (res.data as any).task_id || (res.data as any).TaskID
      store.recordTaskId(createdTaskId.value)
      emit('success', createdTaskId.value)
    } else {
      createError.value = (res.msg as string) || t('download_dialog.create_failed')
    }
  } catch (e: any) {
    createError.value = e?.message || t('common.request_error')
  } finally {
    creating.value = false
  }
}

function handleClose() {
  emit('close')
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '—'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

<template>
  <Teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="overlay" @click.self="handleClose">
        <div class="dialog">
          <div class="dialog__header">
            <h3>{{ t('download_dialog.title') }}</h3>
            <button class="dialog__close" @click="handleClose">&times;</button>
          </div>

          <div class="dialog__body">
            <!-- File path -->
            <div class="info-row">
              <span class="info-row__label">{{ t('download_dialog.file_path') }}</span>
              <code class="info-row__value">{{ filePath }}</code>
            </div>

            <!-- Loading state -->
            <div v-if="resolving" class="state-hint">{{ t('download_dialog.resolving') }}</div>
            <div v-else-if="resolveError" class="msg msg--error">{{ resolveError }}</div>

            <!-- Resolve result -->
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

              <!-- Download directory input -->
              <div class="form-field" style="margin-top: 16px;">
                <label class="info-row__label">{{ t('download_dialog.download_dir') }}</label>
                <input
                  v-model="targetDir"
                  class="path-input"
                  :placeholder="t('download_dialog.dir_placeholder')"
                />
              </div>

              <div v-if="createError" class="msg msg--error" style="margin-top: 12px;">
                {{ createError }}
              </div>

              <div v-if="createdTaskId" class="success-row">
                <span class="success-icon">&#10003;</span>
                <span>{{ t('download_dialog.task_created') }} <code>{{ createdTaskId }}</code></span>
              </div>
            </template>
          </div>

          <div class="dialog__footer">
            <button class="btn btn--secondary" @click="handleClose" :disabled="creating">
              {{ createdTaskId ? t('download_dialog.close') : t('download_dialog.cancel') }}
            </button>
            <button
              v-if="linkResult && !createdTaskId"
              class="btn btn--primary"
              :disabled="creating || resolving"
              @click="handleConfirm"
            >
              {{ creating ? t('download_dialog.submitting') : t('download_dialog.submit') }}
            </button>
            <button
              v-if="createdTaskId"
              class="btn btn--primary"
              @click="handleClose"
            >
              {{ t('download_dialog.done') }}
            </button>
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

.dialog {
  background: white;
  border-radius: 12px;
  width: 520px;
  max-width: 90vw;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
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
}

.dialog__close {
  background: none;
  border: none;
  font-size: 24px;
  color: #9ca3af;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
}
.dialog__close:hover { color: #374151; background: #f3f4f6; }

.dialog__body {
  padding: 20px 24px;
}

.dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 0 24px 20px;
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
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-row__value {
  font-size: 14px;
  color: #111827;
}

code.info-row__value {
  background: #f3f4f6;
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

.path-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s;
}
.path-input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px rgba(59,130,246,0.15); }

.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  width: fit-content;
}
.badge--yes { background: #fef3c7; color: #92400e; }
.badge--no  { background: #d1fae5; color: #065f46; }

.state-hint {
  text-align: center;
  padding: 24px;
  color: #6b7280;
  font-size: 14px;
}

.msg {
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
}
.msg--error { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }

.success-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #d1fae5;
  color: #065f46;
  border-radius: 8px;
  font-size: 14px;
  margin-top: 12px;
}
.success-icon { font-size: 18px; font-weight: 700; }
.success-row code {
  font-weight: 600;
  background: rgba(0,0,0,0.06);
  padding: 2px 8px;
  border-radius: 4px;
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
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn--primary { background: #3b82f6; color: white; }
.btn--primary:hover:not(:disabled) { background: #2563eb; }
.btn--secondary { background: white; color: #374151; border: 1px solid #d1d5db; }
.btn--secondary:hover:not(:disabled) { background: #f9fafb; }
.btn--copy {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  white-space: nowrap;
  transition: all 0.2s;
  flex-shrink: 0;
}
.btn--copy:hover { background: #f3f4f6; border-color: #9ca3af; }
.btn--copy:active { background: #e5e7eb; }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.form-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
</style>
