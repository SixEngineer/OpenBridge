<script setup lang="ts">
import { ref, computed } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'

const store = useConsoleStore()

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
</script>

<template>
  <section class="page">
    <PageHeader
      title="Settings"
      description="System configuration and connection overview"
    />

    <div class="settings-grid">
      <article class="card">
        <h3 class="card__title">API Connection</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">Base URL</span>
            <code class="field__value">{{ apiBaseUrl }}</code>
          </div>
          <div class="field">
            <span class="field__label">Proxy Target</span>
            <code class="field__value">http://localhost:8080</code>
          </div>
          <div class="field">
            <span class="field__label">Timeout</span>
            <span class="field__value">10s</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">OpenList</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">Base URL</span>
            <code class="field__value">http://localhost:5244</code>
          </div>
          <div class="field">
            <span class="field__label">Status</span>
            <span class="field__value">
              <span class="dot dot--unknown"></span>
              Not connected (configure in OpenList desktop)
            </span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">Providers</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">Total Registered</span>
            <span class="field__value">{{ providerSummary.total }}</span>
          </div>
          <div class="field" v-if="providerSummary.types.length">
            <span class="field__label">Types</span>
            <span class="field__value">
              <span v-for="t in providerSummary.types" :key="t" class="tag">{{ t }}</span>
            </span>
          </div>
          <div class="field">
            <span class="field__label">Mounts</span>
            <span class="field__value">{{ mountSummary.count }} active</span>
          </div>
        </div>
      </article>

      <article class="card">
        <h3 class="card__title">aria2</h3>
        <div class="card__body">
          <div class="field">
            <span class="field__label">RPC URL</span>
            <code class="field__value">http://127.0.0.1:6800/jsonrpc</code>
          </div>
          <div class="field">
            <span class="field__label">Default Download Directory</span>
            <div class="dir-input-row">
              <input
                v-model="downloadDirInput"
                class="dir-input"
                placeholder="e.g. D:\Downloads"
                @keyup.enter="saveDownloadDir"
              />
              <button class="btn btn--sm" @click="saveDownloadDir" :disabled="!dirChanged">Save</button>
            </div>
            <span class="field__hint">Leave empty to use aria2's configured default. Changes apply to new downloads.</span>
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
