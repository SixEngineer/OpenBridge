<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '@/components/common/PageHeader.vue'
import LocalPathInput from '@/components/common/LocalPathInput.vue'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import { getUserInfo, userReset } from '@/api/user'
import type { UserInfo } from '@/api/user'
import { getSettings, updateAria2Settings, updateOpenListSettings, updateRcloneSettings } from '@/api/settings'

const store = useConsoleStore()
const router = useRouter()
const { t } = useI18n()

// ── Default download directory ──
const downloadDirInput = ref(store.defaultDownloadDir)
const dirChanged = computed(() => downloadDirInput.value !== store.defaultDownloadDir)

function saveDownloadDir() {
  store.setDefaultDownloadDir(downloadDirInput.value.trim())
}

// ── aria2 RPC URL ──
const aria2UrlInput = ref('http://127.0.0.1:6800/jsonrpc')
const savedAria2Url = ref('')
const settingsLoading = ref(false)
const settingsError = ref(false)
const savingAria2 = ref(false)
const aria2UrlChanged = computed(() => aria2UrlInput.value.trim() !== savedAria2Url.value)
const rclonePathInput = ref('')
const savedRclonePath = ref('')
const savingRclone = ref(false)
const rclonePathChanged = computed(() => rclonePathInput.value.trim() !== savedRclonePath.value)

// ── OpenList URL ──
const olUrlInput = ref('http://127.0.0.1:5244')
const savedOlUrl = ref('')
const savingOpenList = ref(false)
const olUrlChanged = computed(() => olUrlInput.value.trim() !== savedOlUrl.value)

// ── Local session timeout ──
const sessionTimeoutInput = ref(String(store.sessionTimeoutMinutes))
const sessionTimeoutChanged = computed(() => Number(sessionTimeoutInput.value) !== store.sessionTimeoutMinutes)

function saveSessionTimeout() {
  const minutes = Number(sessionTimeoutInput.value)
  store.setSessionTimeout(minutes)
  sessionTimeoutInput.value = String(store.sessionTimeoutMinutes)
}

// ── User Info ──
const userInfo = ref<UserInfo | null>(null)
const userInfoLoading = ref(false)
const userInfoError = ref(false)
const resettingCurrent = ref(false)
const resettingAll = ref(false)

async function fetchUserInfo() {
  userInfoLoading.value = true
  userInfoError.value = false
  try {
    const res = await getUserInfo()
    if (res.code === 1000) {
      userInfo.value = res.data
    } else {
      userInfoError.value = true
    }
  } catch {
    userInfoError.value = true
  } finally {
    userInfoLoading.value = false
  }
}

async function fetchSettings() {
  settingsLoading.value = true
  settingsError.value = false
  try {
    const res = await getSettings()
    savedAria2Url.value = res.data.aria2_rpc_url || 'http://127.0.0.1:6800/jsonrpc'
    savedOlUrl.value = res.data.openlist_base_url || 'http://127.0.0.1:5244'
    savedRclonePath.value = res.data.rclone_path || ''
    aria2UrlInput.value = savedAria2Url.value
    olUrlInput.value = savedOlUrl.value
    rclonePathInput.value = savedRclonePath.value
  } catch {
    settingsError.value = true
  } finally {
    settingsLoading.value = false
  }
}

async function saveRclonePath() {
  savingRclone.value = true
  try {
    const res = await updateRcloneSettings({ path: rclonePathInput.value.trim() })
    savedRclonePath.value = res.data.rclone_path
    rclonePathInput.value = res.data.rclone_path
    settingsError.value = false
  } catch (error) {
    const message = error instanceof Error ? error.message : t('common.request_error')
    alert(message)
  } finally {
    savingRclone.value = false
  }
}

async function saveAria2Url() {
  savingAria2.value = true
  try {
    const res = await updateAria2Settings({ rpc_url: aria2UrlInput.value.trim() })
    savedAria2Url.value = res.data.aria2_rpc_url
    aria2UrlInput.value = res.data.aria2_rpc_url
    settingsError.value = false
  } catch (error) {
    const message = error instanceof Error ? error.message : t('common.request_error')
    alert(message)
  } finally {
    savingAria2.value = false
  }
}

async function saveOlUrl() {
  savingOpenList.value = true
  try {
    const res = await updateOpenListSettings({ base_url: olUrlInput.value.trim() })
    savedOlUrl.value = res.data.openlist_base_url
    olUrlInput.value = res.data.openlist_base_url
    settingsError.value = false
    store.logout('source_changed')
    router.replace('/login')
  } catch (error) {
    const message = error instanceof Error ? error.message : t('common.request_error')
    alert(message)
  } finally {
    savingOpenList.value = false
  }
}

async function handleReset(scope: 'current' | 'all') {
  const confirmKey = scope === 'current' ? 'settings.reset_current_confirm' : 'settings.reset_all_confirm'
  const successKey = scope === 'current' ? 'settings.reset_current_success' : 'settings.reset_all_success'
  const pending = scope === 'current' ? resettingCurrent : resettingAll

  if (!window.confirm(t(confirmKey))) {
    return
  }

  pending.value = true
  try {
    await userReset(scope)
    await Promise.all([
      store.fetchProviders(),
      store.fetchAllMounts(),
    ])
    alert(t(successKey))
  } catch (error) {
    const message = error instanceof Error ? error.message : t('settings.reset_failed')
    alert(message)
  } finally {
    pending.value = false
  }
}

function roleLabel(role: number): string {
  // OpenList 角色: 0=GENERAL, 1=GUEST, 2=ADMIN
  switch (role) {
    case 0: return t('settings.user_role_user')
    case 1: return t('settings.user_role_guest')
    case 2: return t('settings.user_role_admin')
    default: return t('settings.user_role_unknown', { role })
  }
}

onMounted(() => {
  sessionTimeoutInput.value = String(store.sessionTimeoutMinutes)
  fetchUserInfo()
  fetchSettings()
})
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('settings.title')"
      :description="t('settings.description')"
    />

    <div class="settings-grid">
      <!-- User Info card -->
      <article class="card">
        <h3 class="card__title">
          {{ t('settings.user_info') }}
          <button
            v-if="!userInfoLoading"
            class="card__refresh"
            @click="fetchUserInfo"
            :title="t('common.refresh')"
          >&#x21bb;</button>
        </h3>
        <div class="card__body">
          <!-- Loading -->
          <div v-if="userInfoLoading" class="user-info__loading">
            <span class="loading-dot"></span>
          </div>

          <!-- Error -->
          <div v-else-if="userInfoError" class="user-info__error">
            <p>{{ t('settings.load_error') }}</p>
            <button class="btn btn--sm" @click="fetchUserInfo">{{ t('common.retry') }}</button>
          </div>

          <!-- User info -->
          <div v-else-if="userInfo" class="user-info">
            <div class="user-info__row">
              <span class="field__label">{{ t('settings.username') }}</span>
              <span class="user-info__value">{{ userInfo.username }}</span>
            </div>
            <div class="user-info__row">
              <span class="field__label">{{ t('settings.user_role') }}</span>
              <span class="user-info__value">
                <span class="role-badge" :class="{ 'role-badge--admin': userInfo.role === 0 }">
                  {{ roleLabel(userInfo.role) }}
                </span>
              </span>
            </div>
            <div class="user-info__row">
              <span class="field__label">{{ t('settings.base_path') }}</span>
              <span class="user-info__value user-info__value--mono">{{ userInfo.base_path || '—' }}</span>
            </div>
            <div class="user-info__row">
              <span class="field__label">{{ t('settings.sso_id') }}</span>
              <span class="user-info__value user-info__value--mono">{{ userInfo.sso_id || '—' }}</span>
            </div>
            <div class="user-info__row">
              <span class="field__label">{{ t('settings.otp_enabled') }}</span>
              <span class="user-info__value" :class="userInfo.otp ? 'user-info__status--ok' : 'user-info__status--off'">
                {{ userInfo.otp ? t('settings.otp_enabled') : t('settings.otp_disabled') }}
              </span>
            </div>
            <div class="user-info__row">
              <span class="field__label">{{ t('common.status') }}</span>
              <span class="user-info__value" :class="userInfo.disabled ? 'user-info__status--err' : 'user-info__status--ok'">
                {{ userInfo.disabled ? t('settings.user_disabled') : t('settings.user_active') }}
              </span>
            </div>
          </div>

          <!-- No data -->
          <p v-else class="user-info__empty">{{ t('common.no_data') }}</p>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.aria2') }}</h3>
        <div class="card__body">
          <p v-if="settingsError" class="field__error">{{ t('settings.api_error') }}</p>
          <div class="field">
            <span class="field__label">{{ t('settings.rpc_url') }}</span>
            <div class="input-row">
              <input
                v-model="aria2UrlInput"
                class="config-input"
                placeholder="http://127.0.0.1:6800/jsonrpc"
                @keyup.enter="saveAria2Url"
              />
              <button class="btn btn--sm" @click="saveAria2Url" :disabled="settingsLoading || savingAria2 || !aria2UrlChanged">{{ t('settings.save') }}</button>
            </div>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.download_dir') }}</span>
            <div class="input-row">
              <LocalPathInput
                v-model="downloadDirInput"
                :placeholder="t('settings.dir_placeholder')"
                mode="directory"
                :title="t('settings.download_dir')"
              />
              <button class="btn btn--sm" @click="saveDownloadDir" :disabled="!dirChanged">{{ t('settings.save') }}</button>
            </div>
            <span class="field__hint">{{ t('settings.dir_hint') }}</span>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.session_timeout') }}</span>
            <div class="input-row">
              <input
                v-model="sessionTimeoutInput"
                class="config-input"
                inputmode="numeric"
                placeholder="120"
                @keyup.enter="saveSessionTimeout"
              />
              <button class="btn btn--sm" @click="saveSessionTimeout" :disabled="!sessionTimeoutChanged">{{ t('settings.save') }}</button>
            </div>
            <span class="field__hint">{{ t('settings.session_timeout_hint') }}</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('sidebar.items.rclone') }}</h3>
        <div class="card__body">
          <p v-if="settingsError" class="field__error">{{ t('settings.api_error') }}</p>
          <div class="field">
            <span class="field__label">{{ t('settings.rclone_path') }}</span>
            <div class="input-row">
              <LocalPathInput
                v-model="rclonePathInput"
                placeholder="C:\\Program Files\\rclone\\rclone.exe"
                mode="file"
                :title="t('settings.rclone_path')"
                filter="Executable (*.exe)|*.exe|All files (*.*)|*.*"
              />
              <button class="btn btn--sm" @click="saveRclonePath" :disabled="settingsLoading || savingRclone || !rclonePathChanged">{{ t('settings.save') }}</button>
            </div>
            <span class="field__hint">{{ t('settings.rclone_path_hint') }}</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.openlist') }}</h3>
        <div class="card__body">
          <p v-if="settingsError" class="field__error">{{ t('settings.api_error') }}</p>
          <div class="field">
            <span class="field__label">{{ t('settings.openlist_url') }}</span>
            <div class="input-row">
              <input
                v-model="olUrlInput"
                class="config-input"
                placeholder="http://127.0.0.1:5244"
                @keyup.enter="saveOlUrl"
              />
              <button class="btn btn--sm" @click="saveOlUrl" :disabled="settingsLoading || savingOpenList || !olUrlChanged">{{ t('settings.save') }}</button>
            </div>
          </div>
        </div>
      </article>

      <article v-if="store.isAdmin" class="card card--danger">
        <h3 class="card__title">{{ t('settings.reset') }}</h3>
        <div class="card__body">
          <p class="field__hint">{{ t('settings.reset_desc') }}</p>
          <div class="field">
            <span class="field__label">{{ t('settings.reset_current') }}</span>
            <div class="input-row">
              <span class="field__hint">{{ t('settings.reset_current_desc') }}</span>
              <button
                class="btn btn--sm btn--danger"
                @click="handleReset('current')"
                :disabled="resettingCurrent || resettingAll"
              >
                {{ resettingCurrent ? t('settings.resetting') : t('settings.reset_current_button') }}
              </button>
            </div>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.reset_all') }}</span>
            <div class="input-row">
              <span class="field__hint">{{ t('settings.reset_all_desc') }}</span>
              <button
                class="btn btn--sm btn--danger"
                @click="handleReset('all')"
                :disabled="resettingCurrent || resettingAll"
              >
                {{ resettingAll ? t('settings.resetting') : t('settings.reset_all_button') }}
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>

    <p class="version-info">OpenBridge v0.1.0</p>
  </section>
</template>

<style scoped>
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.card {
  background: var(--surface);
  border-radius: 12px;
  border: 1px solid var(--border);
  padding: 20px;
}

.card--danger {
  border-color: rgba(239, 68, 68, 0.28);
}

.card__title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 8px;
}

.card__refresh {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: var(--muted);
  padding: 0 4px;
  transition: color 0.2s;
}
.card__refresh:hover {
  color: #3b82f6;
}

.card__body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.field__label {
  font-size: 12px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.field__hint {
  font-size: 12px;
  color: var(--muted);
}

.field__error {
  margin: 0;
  font-size: 13px;
  color: #dc2626;
}

.input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.config-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  font-size: 14px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  outline: none;
  background: var(--surface);
  color: var(--text);
  transition: border-color 0.2s;
}
.config-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59,130,246,0.15);
}

.btn--sm {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  background: #3b82f6;
  color: white;
  transition: background 0.2s;
  white-space: nowrap;
}
.btn--sm:hover:not(:disabled) { background: #2563eb; }
.btn--sm:disabled { opacity: 0.6; cursor: not-allowed; }

.btn--danger {
  background: #dc2626;
}
.btn--danger:hover:not(:disabled) {
  background: #b91c1c;
}

.version-info {
  text-align: center;
  margin: 32px 0 0;
  font-size: 13px;
  color: var(--muted);
}

/* ── User info card ── */
.user-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-info__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.user-info__value {
  font-size: 14px;
  color: var(--text);
  text-align: right;
}

.user-info__value--mono {
  font-family: 'SFMono-Regular', Consolas, monospace;
  font-size: 13px;
}

.user-info__status--ok {
  color: #10b981;
}

.user-info__status--off {
  color: var(--muted);
}

.user-info__status--err {
  color: #dc2626;
}

.role-badge {
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  background: var(--surface-strong);
  color: var(--text);
}

.role-badge--admin {
  background: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
}

.user-info__loading {
  display: flex;
  justify-content: center;
  padding: 20px;
}

.loading-dot {
  width: 24px;
  height: 24px;
  border: 3px solid var(--border);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.user-info__error {
  text-align: center;
  padding: 16px;
  color: #dc2626;
  font-size: 14px;
}
.user-info__error p {
  margin: 0 0 12px;
}

.user-info__empty {
  text-align: center;
  padding: 20px;
  color: var(--muted);
  font-size: 14px;
  margin: 0;
}

@media (max-width: 860px) {
  .settings-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .card {
    padding: 16px;
  }

  .input-row {
    flex-direction: column;
    align-items: stretch;
  }

  .config-input {
    width: 100%;
  }

  .user-info__row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .user-info__value {
    text-align: left;
  }
}
</style>
