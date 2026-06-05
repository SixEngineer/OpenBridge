<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConsoleStore } from '@/stores/console'
import { getDrivers } from '@/api/storage'
import type { MetricCardData } from '@/types/dashboard'

const store = useConsoleStore()

const openListStatus = ref<'active' | 'error' | 'disabled'>('disabled')
const openListDetail = ref('Checking connection...')

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const metricCards = computed<MetricCardData[]>(() => [
  {
    title: 'Active Providers',
    value: String(store.providers.length).padStart(2, '0'),
    detail: 'Adapters currently online',
    trend: store.providers.length > 0 ? `${store.providers.length} registered` : 'No providers yet',
  },
  {
    title: 'Active Mounts',
    value: String(Object.keys(store.mountIdByProvider).length).padStart(2, '0'),
    detail: 'Mount points created',
    trend: store.currentQuota
      ? `Used ${formatBytes(store.currentQuota.used)} / ${formatBytes(store.currentQuota.total)}`
      : 'No quota data',
  },
  {
    title: 'Storage Used',
    value: store.currentQuota
      ? `${((store.currentQuota.used / store.currentQuota.total) * 100).toFixed(1)}%`
      : '—',
    detail: store.currentQuota
      ? `${formatBytes(store.currentQuota.used)} of ${formatBytes(store.currentQuota.total)}`
      : 'No storage mounted',
    trend: store.currentQuota
      ? `${formatBytes(store.currentQuota.available)} available`
      : 'Create a mount first',
  },
  {
    title: 'OpenList Status',
    value: openListStatus.value === 'active' ? 'Online' : 'Offline',
    detail: openListDetail.value,
    trend: openListStatus.value === 'active' ? 'API responding' : 'Check OpenList connection',
  },
])

const healthyCount = computed(() => {
  let count = 0
  // 后端 API 始终算作健康（因为能加载页面就说明后端正常）
  count++
  // OpenList 连接状态
  if (openListStatus.value === 'active') count++
  // Quota 数据状态
  if (store.currentQuota) count++
  return count
})

async function checkOpenList() {
  try {
    const res = await getDrivers()
    if (res.code === 1000 || res.code === 0) {
      openListStatus.value = 'active'
      openListDetail.value = `Connected — ${res.data?.length || 0} driver(s) available`
    } else {
      openListStatus.value = 'error'
      openListDetail.value = (res.msg as string) || 'API error'
    }
  } catch (e: any) {
    openListStatus.value = 'error'
    openListDetail.value = 'Not connected (configure in OpenList desktop)'
  }
}

onMounted(() => {
  store.fetchProviders()
  checkOpenList()
})
</script>

<template>
  <section class="page">
    <PageHeader
      title="Dashboard"
      description="Service health, task flow, and quota pressure across the OpenBridge platform."
    />

    <div class="grid grid--metrics">
      <MetricCard v-for="item in metricCards" :key="item.title" :item="item" />
    </div>

    <div class="dashboard-panels">
      <section class="panel">
        <div class="panel__header">
          <h3>System Health</h3>
          <p>{{ healthyCount }}/3 services healthy</p>
        </div>
        <div class="status-list">
          <article class="status-row">
            <div>
              <p class="status-row__name">Backend API</p>
              <p class="status-row__detail">
                Connected — API responding normally
              </p>
            </div>
            <StatusBadge state="active" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">OpenList</p>
              <p class="status-row__detail">{{ openListDetail }}</p>
            </div>
            <StatusBadge :state="openListStatus" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">Quota Sync</p>
              <p class="status-row__detail">
                {{ store.currentQuota ? 'Last updated: ' + new Date(store.currentQuota.updated_at).toLocaleString() : 'No quota data' }}
              </p>
            </div>
            <StatusBadge :state="store.currentQuota ? 'active' : 'disabled'" />
          </article>
        </div>
      </section>

      <section class="panel">
        <div class="panel__header">
          <h3>Quick Actions</h3>
          <p>Common operations for managing your cloud storage</p>
        </div>
        <div class="action-grid">
          <router-link to="/providers" class="action-card">
            <div class="action-card__icon">📦</div>
            <h4>Manage Providers</h4>
            <p>Register and configure cloud storage providers</p>
          </router-link>
          <router-link to="/quota" class="action-card">
            <div class="action-card__icon">💾</div>
            <h4>View Quota</h4>
            <p>Monitor storage capacity and usage</p>
          </router-link>
          <router-link to="/openlist" class="action-card">
            <div class="action-card__icon">📁</div>
            <h4>Browse Files</h4>
            <p>Navigate through your mounted drives</p>
          </router-link>
          <router-link to="/tasks" class="action-card">
            <div class="action-card__icon">⬇️</div>
            <h4>Download Tasks</h4>
            <p>View and manage download operations</p>
          </router-link>
        </div>
      </section>
    </div>
  </section>
</template>


<style scoped>
.dashboard-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-top: 20px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.action-card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 20px;
  text-decoration: none;
  color: inherit;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.action-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
  transform: translateY(-2px);
}

.action-card__icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.action-card h4 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.action-card p {
  margin: 0;
  font-size: 13px;
  color: #6b7280;
  line-height: 1.4;
}

@media (max-width: 1024px) {
  .dashboard-panels {
    grid-template-columns: 1fr;
  }
  
  .action-grid {
    grid-template-columns: 1fr;
  }
}
</style>
