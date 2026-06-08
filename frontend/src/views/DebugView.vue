<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'
import { getProviderList } from '@/api/provider'
import { userReset } from '@/api/user'

const store = useConsoleStore()
const { t } = useI18n()

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// Ping backend
const pingResult = ref<{ ok: boolean; data?: any; error?: string } | null>(null)
const pinging = ref(false)

async function handlePing() {
  pinging.value = true
  pingResult.value = null
  try {
    const res = await getProviderList()
    pingResult.value = { ok: true, data: res }
  } catch (e: any) {
    pingResult.value = { ok: false, error: e?.message || t('common.request_error') }
  } finally {
    pinging.value = false
  }
}

// Store state visibility
const showStore = ref(false)

// Reset data
const resetting = ref(false)
const confirmReset = ref(false)

async function handleReset() {
  if (!confirmReset.value) {
    confirmReset.value = true
    return
  }
  resetting.value = true
  try {
    await userReset('all')
    store.fetchProviders()
    store.fetchAllMounts()
    confirmReset.value = false
    alert(t('settings.reset_success'))
  } catch (e: any) {
    alert(e.message || t('settings.reset_failed'))
  } finally {
    resetting.value = false
  }
}
</script>

<template>
  <section class="page">
    <PageHeader
      :title="t('debug.title')"
      :description="t('debug.description')"
    />

    <div class="debug-grid">
      <!-- Health Check -->
      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('debug.backend_connectivity') }}</h3>
          <p>{{ t('debug.backend_connectivity_desc') }}</p>
        </div>
        <div class="debug-ping">
          <button
            class="btn btn--primary"
            :disabled="pinging"
            @click="handlePing"
          >
            {{ pinging ? t('debug.pinging') : t('debug.ping_backend') }}
          </button>

          <div v-if="pingResult" class="ping-result" :class="{ 'ping-result--ok': pingResult.ok, 'ping-result--err': !pingResult.ok }">
            <p class="ping-result__status">{{ pingResult.ok ? t('debug.connected') : t('debug.failed') }}</p>
            <p class="ping-result__detail">
              {{ pingResult.ok ? t('debug.provider_list_count', { count: pingResult.data?.data?.length || 0 }) : pingResult.error }}
            </p>
          </div>
        </div>
      </section>

      <!-- Provider Debug -->
      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('debug.provider_registry') }}</h3>
          <p>{{ t('debug.registered_count', { count: store.providers.length }) }}</p>
        </div>
        <div class="provider-debug" v-if="store.providers.length > 0">
          <div v-for="p in store.providers" :key="p.id" class="provider-row">
            <div class="provider-row__info">
              <span class="provider-row__name">{{ p.name }}</span>
              <code class="provider-row__id">ID: {{ p.id }}</code>
            </div>
            <span class="provider-type-tag" :class="`provider-type-tag--${p.provider_type}`">{{ p.provider_type }}</span>
          </div>
        </div>
        <p v-else class="empty-hint">{{ t('debug.no_providers') }}</p>
      </section>

      <!-- Store State -->
      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('debug.store_state') }}</h3>
          <button class="btn btn--sm" @click="showStore = !showStore">
            {{ showStore ? t('debug.hide') : t('debug.show') }}
          </button>
        </div>
        <pre v-if="showStore" class="store-dump">{{ JSON.stringify({
          providers: store.providers,
          mountIdByProvider: store.mountIdByProvider,
          currentQuota: store.currentQuota,
          quotaLoading: store.quotaLoading,
        }, null, 2) }}</pre>
      </section>

      <!-- API Connection Info -->
      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('debug.api_info') }}</h3>
          <p>{{ t('debug.api_info_desc') }}</p>
        </div>
        <div class="info-grid">
          <div class="info-field">
            <span class="info-field__label">{{ t('settings.base_url') }}</span>
            <code class="info-field__value">{{ apiBaseUrl }}</code>
          </div>
          <div class="info-field">
            <span class="info-field__label">{{ t('settings.proxy_target') }}</span>
            <code class="info-field__value">http://localhost:8080</code>
          </div>
        </div>
      </section>

      <!-- Reset Data -->
      <section v-if="store.isAdmin" class="panel panel--danger">
        <div class="panel__header">
          <h3>{{ t('settings.reset') }}</h3>
          <p>{{ t('settings.reset_desc') }}</p>
        </div>
        <button
          class="btn btn--danger"
          :disabled="resetting"
          @click="handleReset"
        >
          {{ confirmReset ? t('settings.reset_confirm') : t('settings.reset_button') }}
        </button>
      </section>
    </div>
  </section>
</template>

<style scoped>
.debug-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.btn--sm {
  padding: 6px 14px;
  font-size: 13px;
  background: var(--surface-strong);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}
.btn--sm:hover {
  background: var(--surface);
  border-color: var(--muted);
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

.debug-ping {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ping-result {
  padding: 14px;
  border-radius: 8px;
}

.ping-result--ok {
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.2);
}

.ping-result--err {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.ping-result__status {
  margin: 0 0 4px;
  font-weight: 600;
  font-size: 14px;
}

.ping-result--ok .ping-result__status {
  color: #16a34a;
}

.ping-result--err .ping-result__status {
  color: #dc2626;
}

.ping-result__detail {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
}

.provider-debug {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.provider-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: var(--surface);
  border-radius: 8px;
  border: 1px solid var(--border);
}

.provider-row__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.provider-row__name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
}

.provider-row__id {
  font-size: 12px;
  color: var(--muted);
  font-family: 'SFMono-Regular', Consolas, monospace;
}

/* Provider type color tags (matching Dashboard) */
.provider-type-tag {
  font-size: 11px;
  padding: 2px 10px;
  border-radius: 20px;
  font-weight: 600;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}

.provider-type-tag--mock {
  background: rgba(156, 163, 175, 0.15);
  color: #6b7280;
}
.provider-type-tag--baidu {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
}
.provider-type-tag--quark {
  background: rgba(251, 146, 60, 0.12);
  color: #f97316;
}
.provider-type-tag--local {
  background: rgba(16, 185, 129, 0.12);
  color: #10b981;
}

[data-theme="dark"] .provider-type-tag--mock {
  background: rgba(156, 163, 175, 0.2);
  color: #9ca3af;
}
[data-theme="dark"] .provider-type-tag--baidu {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}
[data-theme="dark"] .provider-type-tag--quark {
  background: rgba(251, 146, 60, 0.2);
  color: #fb923c;
}
[data-theme="dark"] .provider-type-tag--local {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

.empty-hint {
  color: var(--muted);
  font-size: 14px;
  margin: 0;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-field {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.info-field__label {
  font-size: 12px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-field__value {
  font-size: 14px;
  color: var(--text);
}

code.info-field__value {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;
  font-size: 13px;
  background: var(--surface);
  padding: 2px 8px;
  border-radius: 4px;
  width: fit-content;
  border: 1px solid var(--border);
}

.panel--danger {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.06);
}
.panel--danger .panel__header h3 {
  color: #dc2626;
}
.panel--danger .panel__header p {
  color: #dc2626;
}

.btn--danger {
  padding: 8px 20px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  background: #ef4444;
  color: white;
  transition: background 0.2s;
}
.btn--danger:hover:not(:disabled) { background: #dc2626; }
.btn--danger:disabled { opacity: 0.6; cursor: not-allowed; }

.store-dump {
  background: #1f2937;
  color: #e5e7eb;
  padding: 16px;
  border-radius: 8px;
  font-size: 12px;
  font-family: 'SFMono-Regular', Consolas, monospace;
  overflow-x: auto;
  max-height: 400px;
  overflow-y: auto;
  margin: 0;
}

@media (max-width: 860px) {
  .debug-grid {
    gap: 14px;
  }

  .panel {
    padding: 14px;
  }

  .panel__header {
    flex-direction: column;
    gap: 6px;
  }

  .provider-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }

  .provider-row__info {
    width: 100%;
  }
}
</style>
