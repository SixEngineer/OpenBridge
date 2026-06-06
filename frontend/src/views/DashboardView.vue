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
const { t, locale } = useI18n()

const openListStatus = ref<'active' | 'error' | 'disabled'>('disabled')
const quotaExpanded = ref(false)
const quotasByProvider = computed(() =>
  store.providers
    .filter(p => p.total_quota > 0)
    .map(p => {
      // Mock providers store values in GB but backend reports as MB
      const scale = p.provider_type === 'mock' ? 1024 : 1
      return {
        ...p,
        total_quota: p.total_quota * scale,
        used_quota: p.used_quota * scale,
        available_quota: p.available_quota * scale,
      }
    })
)
const openListDetail = ref(t('dashboard.checking_connection'))

const primaryProviderId = ref<number | null>(null)

const primaryProvider = computed(() => {
  if (primaryProviderId.value !== null) {
    return quotasByProvider.value.find(p => p.id === primaryProviderId.value) ?? null
  }
  return quotasByProvider.value[0] ?? null
})

function selectPrimaryProvider(id: number) {
  primaryProviderId.value = id
}

function formatBytes(mb: number): string {
  if (mb === 0) return '0 GB'
  const bytes = mb * 1024 * 1024
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
    title: primaryProvider.value
      ? `${t('dashboard.storage_used')} (${primaryProvider.value.name})`
      : t('dashboard.storage_used'),
    value: primaryProvider.value
      ? `${((primaryProvider.value.used_quota / primaryProvider.value.total_quota) * 100).toFixed(1)}%`
      : '—',
    detail: primaryProvider.value
      ? `${t('quota.used')} ${formatBytes(primaryProvider.value.used_quota)} / ${formatBytes(primaryProvider.value.total_quota)}`
      : t('dashboard.storage_used_detail_empty'),
    trend: primaryProvider.value
      ? `${t('quota.available')} ${formatBytes(primaryProvider.value.available_quota)}`
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
      openListDetail.value = t('dashboard.openlist_connected')
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
      <MetricCard
        v-for="(item, idx) in metricCards"
        :key="item.title"
        :item="item"
        @click="idx === 2 && (quotaExpanded = !quotaExpanded)"
      />
    </div>

    <!-- Expanded per-provider quota breakdown -->
    <transition name="slide">
      <div v-if="quotaExpanded && quotasByProvider.length > 0" class="quota-expand">
        <div class="quota-expand__header">
          <h4>{{ t('dashboard.storage_used') }}</h4>
        </div>
        <div class="quota-grid">
          <div
            v-for="p in quotasByProvider"
            :key="p.id"
            class="quota-provider-card"
            :class="{ 'quota-provider-card--active': primaryProvider?.id === p.id }"
            @click="selectPrimaryProvider(p.id); quotaExpanded = false"
          >
            <div class="quota-provider-card__top">
              <span class="quota-provider-card__name">{{ p.name }}</span>
              <span class="quota-provider-card__type">{{ p.provider_type }}</span>
            </div>
            <div class="quota-provider-card__stats">
              <div class="qstat">
                <span class="qstat__label">{{ t('quota.total') }}</span>
                <span class="qstat__value">{{ formatBytes(p.total_quota) }}</span>
              </div>
              <div class="qstat">
                <span class="qstat__label">{{ t('quota.used') }}</span>
                <span class="qstat__value">{{ formatBytes(p.used_quota) }}</span>
              </div>
              <div class="qstat">
                <span class="qstat__label">{{ t('quota.available') }}</span>
                <span class="qstat__value qstat__value--available">{{ formatBytes(p.available_quota) }}</span>
              </div>
            </div>
            <div class="quota-provider-card__bar">
              <div
                class="quota-provider-card__bar-fill"
                :style="{ width: `${(p.used_quota / p.total_quota) * 100}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </transition>

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
                {{ store.currentQuota ? t('dashboard.last_updated', { time: new Date(store.currentQuota.updated_at).toLocaleString(locale.value) }) : t('dashboard.no_quota_data') }}
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

/* ── Expanded quota breakdown ── */
.quota-expand {
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
}

.quota-expand__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.quota-expand__header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.quota-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
}

.quota-provider-card {
  padding: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #fafafa;
  cursor: pointer;
  transition: all 0.15s;
}
.quota-provider-card:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}
.quota-provider-card--active {
  border-color: #3b82f6;
  background: #eff6ff;
  box-shadow: 0 0 0 1px #3b82f6;
}

.quota-provider-card__top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.quota-provider-card__name {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.quota-provider-card__type {
  font-size: 11px;
  padding: 2px 8px;
  background: #dbeafe;
  color: #1e40af;
  border-radius: 8px;
  font-weight: 500;
}

.quota-provider-card__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 10px;
}

.qstat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.qstat__label {
  font-size: 11px;
  color: #9ca3af;
}

.qstat__value {
  font-size: 15px;
  font-weight: 600;
  color: #111827;
}

.qstat__value--available {
  color: #10b981;
}

.quota-provider-card__bar {
  height: 6px;
  background: #e5e7eb;
  border-radius: 3px;
  overflow: hidden;
}

.quota-provider-card__bar-fill {
  height: 100%;
  background: #3b82f6;
  border-radius: 3px;
  transition: width 0.3s;
}

.btn--sm {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  transition: all 0.2s;
  white-space: nowrap;
}
.btn--sm:hover { background: #f9fafb; }

.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
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
