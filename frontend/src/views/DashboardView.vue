<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConsoleStore } from '@/stores/console'
import { getDrivers } from '@/api/storage'
import type { MetricCardData } from '@/types/dashboard'

const store = useConsoleStore()
const { t } = useI18n()

const openListStatus = ref<'active' | 'error' | 'disabled'>('disabled')
const openListDetail = ref(t('dashboard.checking_connection'))

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const metricCards = computed<MetricCardData[]>(() => [
  {
    title: t('dashboard.active_providers'),
    value: String(store.providers.length).padStart(2, '0'),
    detail: t('dashboard.active_providers_detail'),
    trend: store.providers.length > 0
      ? t('dashboard.active_providers_trend_count', { count: store.providers.length })
      : t('dashboard.active_providers_trend_empty'),
  },
  {
    title: t('dashboard.active_mounts'),
    value: String(Object.keys(store.mountIdByProvider).length).padStart(2, '0'),
    detail: t('dashboard.active_mounts_detail'),
    trend: store.currentQuota
      ? t('dashboard.active_mounts_trend_used', {
          used: formatBytes(store.currentQuota.used),
          total: formatBytes(store.currentQuota.total),
        })
      : t('dashboard.active_mounts_trend_empty'),
  },
  {
    title: t('dashboard.storage_used'),
    value: store.currentQuota
      ? `${((store.currentQuota.used / store.currentQuota.total) * 100).toFixed(1)}%`
      : '—',
    detail: store.currentQuota
      ? t('dashboard.storage_used_detail_available', {
          used: formatBytes(store.currentQuota.used),
          total: formatBytes(store.currentQuota.total),
        })
      : t('dashboard.storage_used_detail_empty'),
    trend: store.currentQuota
      ? t('dashboard.storage_used_trend_available', { available: formatBytes(store.currentQuota.available) })
      : t('dashboard.storage_used_trend_empty'),
  },
  {
    title: t('dashboard.openlist_status_title'),
    value: openListStatus.value === 'active' ? t('dashboard.openlist_online') : t('dashboard.openlist_offline'),
    detail: openListDetail.value,
    trend: openListStatus.value === 'active' ? t('dashboard.openlist_trend_ok') : t('dashboard.openlist_trend_error'),
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
      openListDetail.value = t('dashboard.connected_drivers', { count: res.data?.length || 0 })
    } else {
      openListStatus.value = 'error'
      openListDetail.value = (res.msg as string) || t('dashboard.api_error')
    }
  } catch (e: any) {
    openListStatus.value = 'error'
    openListDetail.value = t('dashboard.not_connected')
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
      :title="t('dashboard.title')"
      :description="t('dashboard.description')"
    />

    <div class="grid grid--metrics">
      <MetricCard v-for="item in metricCards" :key="item.title" :item="item" />
    </div>

    <div class="dashboard-panels">
      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('dashboard.system_health') }}</h3>
          <p>{{ t('dashboard.healthy_services', { count: healthyCount }) }}</p>
        </div>
        <div class="status-list">
          <article class="status-row">
            <div>
              <p class="status-row__name">{{ t('dashboard.backend_api') }}</p>
              <p class="status-row__detail">
                {{ t('dashboard.backend_api_connected') }}
              </p>
            </div>
            <StatusBadge state="active" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">{{ t('dashboard.openlist_section') }}</p>
              <p class="status-row__detail">{{ openListDetail }}</p>
            </div>
            <StatusBadge :state="openListStatus" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">{{ t('dashboard.quota_sync') }}</p>
              <p class="status-row__detail">
                {{ store.currentQuota ? t('dashboard.last_updated', { time: new Date(store.currentQuota.updated_at).toLocaleString() }) : t('dashboard.no_quota_data') }}
              </p>
            </div>
            <StatusBadge :state="store.currentQuota ? 'active' : 'disabled'" />
          </article>
        </div>
      </section>

      <section class="panel">
        <div class="panel__header">
          <h3>{{ t('dashboard.quick_actions') }}</h3>
          <p>{{ t('dashboard.quick_actions_desc') }}</p>
        </div>
        <div class="action-grid">
          <router-link to="/providers" class="action-card">
            <div class="action-card__icon">📦</div>
            <h4>{{ t('dashboard.manage_providers') }}</h4>
            <p>{{ t('dashboard.manage_providers_desc') }}</p>
          </router-link>
          <router-link to="/quota" class="action-card">
            <div class="action-card__icon">💾</div>
            <h4>{{ t('dashboard.view_quota') }}</h4>
            <p>{{ t('dashboard.view_quota_desc') }}</p>
          </router-link>
          <router-link to="/openlist" class="action-card">
            <div class="action-card__icon">📁</div>
            <h4>{{ t('dashboard.browse_files') }}</h4>
            <p>{{ t('dashboard.browse_files_desc') }}</p>
          </router-link>
          <router-link to="/tasks" class="action-card">
            <div class="action-card__icon">⬇️</div>
            <h4>{{ t('dashboard.download_tasks') }}</h4>
            <p>{{ t('dashboard.download_tasks_desc') }}</p>
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
