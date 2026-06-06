<script setup lang="ts">
import { ref, computed } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'

const store = useConsoleStore()
const { t } = useI18n()

// ── Default download directory ──
const downloadDirInput = ref(store.defaultDownloadDir)
const dirChanged = computed(() => downloadDirInput.value !== store.defaultDownloadDir)

function saveDownloadDir() {
  store.setDefaultDownloadDir(downloadDirInput.value.trim())
}

// ── aria2 RPC URL ──
const ARIA2_URL_KEY = 'openbridge_aria2_rpc_url'
const aria2UrlInput = ref(localStorage.getItem(ARIA2_URL_KEY) || 'http://127.0.0.1:6800/jsonrpc')
const aria2UrlChanged = computed(() => aria2UrlInput.value !== (localStorage.getItem(ARIA2_URL_KEY) || 'http://127.0.0.1:6800/jsonrpc'))

function saveAria2Url() {
  localStorage.setItem(ARIA2_URL_KEY, aria2UrlInput.value.trim())
}

// ── OpenList URL ──
const OL_URL_KEY = 'openbridge_ol_url'
const olUrlInput = ref(localStorage.getItem(OL_URL_KEY) || 'http://localhost:5244')
const olUrlChanged = computed(() => olUrlInput.value !== (localStorage.getItem(OL_URL_KEY) || 'http://localhost:5244'))

function saveOlUrl() {
  localStorage.setItem(OL_URL_KEY, olUrlInput.value.trim())
}
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('settings.title')"
      :description="t('settings.description')"
    />

    <div class="settings-grid">
      <article class="card">
        <h3 class="card__title">{{ t('settings.aria2') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.rpc_url') }}</span>
            <div class="input-row">
              <input
                v-model="aria2UrlInput"
                class="config-input"
                placeholder="http://127.0.0.1:6800/jsonrpc"
                @keyup.enter="saveAria2Url"
              />
              <button class="btn btn--sm" @click="saveAria2Url" :disabled="!aria2UrlChanged">{{ t('settings.save') }}</button>
            </div>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.download_dir') }}</span>
            <div class="input-row">
              <input
                v-model="downloadDirInput"
                class="config-input"
                :placeholder="t('settings.dir_placeholder')"
                @keyup.enter="saveDownloadDir"
              />
              <button class="btn btn--sm" @click="saveDownloadDir" :disabled="!dirChanged">{{ t('settings.save') }}</button>
            </div>
            <span class="field__hint">{{ t('settings.dir_hint') }}</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.openlist') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.openlist_url') }}</span>
            <div class="input-row">
              <input
                v-model="olUrlInput"
                class="config-input"
                placeholder="http://localhost:5244"
                @keyup.enter="saveOlUrl"
              />
              <button class="btn btn--sm" @click="saveOlUrl" :disabled="!olUrlChanged">{{ t('settings.save') }}</button>
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

.card__title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
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

.version-info {
  text-align: center;
  margin: 32px 0 0;
  font-size: 13px;
  color: var(--muted);
}
</style>
