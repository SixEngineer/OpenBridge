<script setup lang="ts">
import { ref } from 'vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConsoleStore } from '@/stores/console'
import { getProviderList } from '@/api/provider'

const store = useConsoleStore()

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
    pingResult.value = { ok: false, error: e?.message || 'Request failed' }
  } finally {
    pinging.value = false
  }
}

// Store state visibility
const showStore = ref(false)
</script>

<template>
  <section class="page">
    <PageHeader
      title="Debug"
      description="Diagnostic tools for development and troubleshooting"
    />

    <div class="debug-grid">
      <!-- Health Check -->
      <section class="panel">
        <div class="panel__header">
          <h3>Backend Connectivity</h3>
          <p>Ping the backend API to verify the connection</p>
        </div>
        <div class="debug-ping">
          <button
            class="btn btn--primary"
            :disabled="pinging"
            @click="handlePing"
          >
            {{ pinging ? 'Pinging...' : 'Ping Backend' }}
          </button>

          <div v-if="pingResult" class="ping-result" :class="{ 'ping-result--ok': pingResult.ok, 'ping-result--err': !pingResult.ok }">
            <p class="ping-result__status">{{ pingResult.ok ? 'Connected' : 'Failed' }}</p>
            <p class="ping-result__detail">
              {{ pingResult.ok ? `Provider list returned ${pingResult.data?.data?.length || 0} items` : pingResult.error }}
            </p>
          </div>
        </div>
      </section>

      <!-- Provider Debug -->
      <section class="panel">
        <div class="panel__header">
          <h3>Provider Registry</h3>
          <p>{{ store.providers.length }} registered</p>
        </div>
        <div class="provider-debug" v-if="store.providers.length > 0">
          <div v-for="p in store.providers" :key="p.id" class="provider-row">
            <div class="provider-row__info">
              <span class="provider-row__name">{{ p.name }}</span>
              <code class="provider-row__id">ID: {{ p.id }}</code>
            </div>
            <span class="tag">{{ p.provider_type }}</span>
          </div>
        </div>
        <p v-else class="empty-hint">No providers registered</p>
      </section>

      <!-- Store State -->
      <section class="panel">
        <div class="panel__header">
          <h3>Store State</h3>
          <button class="btn btn--sm" @click="showStore = !showStore">
            {{ showStore ? 'Hide' : 'Show' }}
          </button>
        </div>
        <pre v-if="showStore" class="store-dump">{{ JSON.stringify({
          providers: store.providers,
          mountIdByProvider: store.mountIdByProvider,
          currentQuota: store.currentQuota,
          quotaLoading: store.quotaLoading,
        }, null, 2) }}</pre>
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
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.ping-result--err {
  background: #fef2f2;
  border: 1px solid #fecaca;
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
  color: #6b7280;
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
  background: #f9fafb;
  border-radius: 8px;
}

.provider-row__info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.provider-row__name {
  font-size: 14px;
  font-weight: 500;
  color: #111827;
}

.provider-row__id {
  font-size: 12px;
  color: #6b7280;
  font-family: 'SFMono-Regular', Consolas, monospace;
}

.tag {
  padding: 3px 10px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.empty-hint {
  color: #9ca3af;
  font-size: 14px;
  margin: 0;
}

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
</style>
