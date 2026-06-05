<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import { getDrivers } from '@/api/storage'

const store = useConsoleStore()
const { t } = useI18n()

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'

const providerSummary = computed(() => ({
  total: store.providers.length,
  types: [...new Set(store.providers.map(p => p.provider_type))],
}))

const mountSummary = computed(() => {
  const ids = Object.keys(store.mountIdByProvider)
  return { count: ids.length, providers: ids }
})

// Default download directory
const downloadDirInput = ref(store.defaultDownloadDir)
const dirChanged = computed(() => downloadDirInput.value !== store.defaultDownloadDir)

function saveDownloadDir() {
  store.setDefaultDownloadDir(downloadDirInput.value.trim())
}

// OpenList status
const openListStatus = ref<'active' | 'error' | 'unknown'>('unknown')
const openListDetail = ref(t('settings.checking_status'))

async function checkOpenList() {
  try {
    const res = await getDrivers()
    if (res.code === 1000 || res.code === 0) {
      openListStatus.value = 'active'
      openListDetail.value = t('settings.connected_drivers', { count: res.data?.length || 0 })
    } else {
      openListStatus.value = 'error'
      openListDetail.value = t('settings.api_error')
    }
  } catch (e: any) {
    openListStatus.value = 'error'
    openListDetail.value = t('settings.openlist_disconnected')
  }
}

onMounted(() => {
  checkOpenList()
})
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('settings.title')"
      :description="t('settings.description')"
    />

    <div class="settings-grid">
      <article class="card">
        <h3 class="card__title">{{ t('settings.api') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.base_url') }}</span>
            <code class="field__value">{{ apiBaseUrl }}</code>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.proxy_target') }}</span>
            <code class="field__value">http://localhost:8080</code>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.timeout') }}</span>
            <span class="field__value">10s</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.openlist') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.openlist_url') }}</span>
            <code class="field__value">http://localhost:5244</code>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.openlist_status') }}</span>
            <span class="field__value">
              <span class="dot" :class="`dot--${openListStatus}`"></span>
              {{ openListDetail }}
            </span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.providers') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.total_registered') }}</span>
            <span class="field__value">{{ providerSummary.total }}</span>
          </div>
          <div class="field" v-if="providerSummary.types.length">
            <span class="field__label">{{ t('settings.types') }}</span>
            <span class="field__value">
              <span v-for="typeItem in providerSummary.types" :key="typeItem" class="tag">{{ typeItem }}</span>
            </span>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.mounts') }}</span>
            <span class="field__value">{{ mountSummary.count }} {{ t('settings.mounts_active') }}</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">{{ t('settings.aria2') }}</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">{{ t('settings.rpc_url') }}</span>
            <code class="field__value">http://127.0.0.1:6800/jsonrpc</code>
          </div>
          <div class="field">
            <span class="field__label">{{ t('settings.download_dir') }}</span>
            <div class="dir-input-row">
              <input
                v-model="downloadDirInput"
                class="dir-input"
                :placeholder="t('settings.dir_placeholder')"
                @keyup.enter="saveDownloadDir"
              />
              <button class="btn btn--sm" @click="saveDownloadDir" :disabled="!dirChanged">{{ t('settings.save') }}</button>
            </div>
            <span class="field__hint">{{ t('settings.dir_hint') }}</span>
          </div>
        </div>
      </article>
    </div>
  </section>
</template>

<style scoped>
.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
}

.card {
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  padding: 20px;
}

.card__title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: #111827;
}

.card__body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.field__label {
  font-size: 12px;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.field__value {
  font-size: 14px;
  color: #111827;
}

code.field__value {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  background: #f3f4f6;
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
}

.tag {
  display: inline-block;
  padding: 2px 8px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  margin-right: 4px;
}

.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

.dot--unknown {
  background: #9ca3af;
}

.dot--active {
  background: #10b981;
}

.dot--error {
  background: #ef4444;
}

.dir-input-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.dir-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}
.dir-input:focus {
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

.field__hint {
  font-size: 12px;
  color: #6b7280;
}
</style>
