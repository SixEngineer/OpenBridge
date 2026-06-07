<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConsoleStore } from '@/stores/console'
import type { MetricCardData } from '@/types/dashboard'
import { getProviderList } from '@/api/provider'
import { getAria2Status } from '@/api/task'

const store = useConsoleStore()
const { t, locale } = useI18n()

const aria2Status = ref<'active' | 'error' | 'disabled'>('disabled')
const quotaExpanded = ref(false)
const mountsExpanded = ref(false)
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
const aria2Detail = ref(t('dashboard.checking_connection'))
const backendStatus = ref<'active' | 'error' | 'disabled'>('active')
const backendDetail = ref(t('dashboard.backend_api_connected'))

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

function usedPercent(used: number, total: number): number {
  if (total <= 0) return 0
  const pct = Math.round((used / total) * 100)
  if (pct === 0 && used > 0) return 1
  return Math.min(pct, 100)
}

function formatBytes(mb: number): string {
  if (mb === 0) return '0 GB'
  const bytes = mb * 1024 * 1024
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
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
    value: String(store.allMounts.length).padStart(2, '0'),
    detail: t('dashboard.active_mounts_detail'),
    trend: '',
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
    title: t('dashboard.aria2_status_title'),
    value: aria2Status.value === 'active' ? t('dashboard.aria2_online') : t('dashboard.aria2_offline'),
    detail: aria2Detail.value,
    trend: aria2Status.value === 'active' ? t('dashboard.aria2_trend_ok') : t('dashboard.aria2_trend_error'),
  },
])

const healthyCount = computed(() => {
  let count = 0
  // 后端 API 健康检查
  if (backendStatus.value === 'active') count++
  // aria2 RPC 连接状态
  if (aria2Status.value === 'active') count++
  // Quota 数据状态（后端离线时不计入健康）
  if (backendStatus.value === 'active' && store.currentQuota) count++
  return count
})

async function checkAria2() {
  try {
    const data = await getAria2Status()
    if (data.code === 1000 && data.data) {
      aria2Status.value = 'active'
      aria2Detail.value = t('dashboard.aria2_connected', { version: data.data.version || '' })
    } else {
      aria2Status.value = 'error'
      aria2Detail.value = t('dashboard.aria2_not_connected')
    }
  } catch (e: any) {
    aria2Status.value = 'error'
    aria2Detail.value = t('dashboard.aria2_not_connected')
  }
}

async function checkBackend() {
  try {
    const res = await getProviderList()
    if (res.code === 1000 || res.code === 0) {
      backendStatus.value = 'active'
      backendDetail.value = t('dashboard.backend_api_connected')
    } else {
      backendStatus.value = 'error'
      backendDetail.value = t('dashboard.backend_api_error')
    }
  } catch (e: any) {
    backendStatus.value = 'error'
    backendDetail.value = t('dashboard.backend_api_disconnected')
  }
}

onMounted(() => {
  store.fetchProviders()
  store.fetchAllMounts()
  checkBackend()
  checkAria2()
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
        @click="idx === 1 && backendStatus === 'active' ? (mountsExpanded = !mountsExpanded, quotaExpanded = false) : idx === 2 && (quotaExpanded = !quotaExpanded, mountsExpanded = false)"
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
              <span class="provider-type-tag" :class="`provider-type-tag--${p.provider_type}`">{{ p.provider_type }}</span>
            </div>
            <div class="quota-provider-card__body">
              <div class="quota-provider-card__chart-wrap">
                <svg viewBox="0 0 36 36" class="donut">
                  <path
                    class="donut__bg"
                    d="M18 2.0845 a 15.9155 15.9155 0 0 0 0 31.831 a 15.9155 15.9155 0 0 0 0 -31.831"
                  />
                  <path
                    class="donut__fill"
                    stroke="#10b981"
                    :stroke-dasharray="`${usedPercent(p.used_quota, p.total_quota)} ${100 - usedPercent(p.used_quota, p.total_quota)}`"
                    d="M18 2.0845 a 15.9155 15.9155 0 0 0 0 31.831 a 15.9155 15.9155 0 0 0 0 -31.831"
                  />
                  <text x="18" y="17.5" class="donut__text">{{ usedPercent(p.used_quota, p.total_quota) }}%</text>
                </svg>
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
            </div>
          </div>
        </div>
      </div>
    </transition>

    <!-- Expanded mount list -->
    <transition name="slide">
      <div v-if="mountsExpanded && store.mounts.length > 0" class="quota-expand">
        <div class="quota-expand__header">
          <h4>{{ t('dashboard.active_mounts') }}</h4>
        </div>
        <div class="quota-grid">
          <div
            v-for="m in store.mounts"
            :key="m.id"
            class="quota-provider-card"
          >
            <div class="quota-provider-card__top">
              <span class="quota-provider-card__name">{{ m.name }}</span>
              <span class="mount-mode-badge" :class="`mount-mode-badge--${m.mode}`">{{ m.mode }}</span>
            </div>
            <div class="quota-provider-card__stats" style="grid-template-columns: 1fr;">
              <div class="qstat">
                <span class="qstat__label">{{ t('dashboard.active_mounts_detail') }}</span>
                <span class="qstat__value" style="font-size:13px;font-weight:400;">{{ m.providerName }}</span>
              </div>
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
              <p class="status-row__detail">{{ backendDetail }}</p>
            </div>
            <StatusBadge :state="backendStatus" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">{{ t('dashboard.aria2_section') }}</p>
              <p class="status-row__detail">{{ aria2Detail }}</p>
            </div>
            <StatusBadge :state="aria2Status" />
          </article>
          <article class="status-row">
            <div>
              <p class="status-row__name">{{ t('dashboard.quota_sync') }}</p>
              <p class="status-row__detail">
                {{ backendStatus !== 'active'
                  ? t('dashboard.backend_api_disconnected')
                  : store.currentQuota
                    ? t('dashboard.last_updated', { time: new Date(store.currentQuota.updated_at).toLocaleString(locale) })
                    : t('dashboard.no_quota_data') }}
              </p>
            </div>
            <StatusBadge :state="backendStatus !== 'active' ? 'error' : (store.currentQuota ? 'active' : 'disabled')" />
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
  background: var(--surface);
  border: 1px solid var(--border);
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
  color: var(--text);
}

.action-card p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 1.4;
}

/* ── Expanded quota breakdown ── */
.quota-expand {
  margin-bottom: 20px;
  padding: 20px;
  background: var(--surface);
  border-radius: 12px;
  border: 1px solid var(--border);
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
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--surface);
  cursor: pointer;
  transition: all 0.15s;
}
.quota-provider-card:hover {
  border-color: var(--accent);
  background: rgba(59, 130, 246, 0.06);
}
.quota-provider-card--active {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.08);
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
  color: var(--text);
}

/* Provider type color tags */
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

/* Mount mode badges (active mounts expand section) */
.mount-mode-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 8px;
  font-weight: 600;
}
.mount-mode-badge--real {
  background: #d1fae5;
  color: #065f46;
}
.mount-mode-badge--inherit {
  background: #dbeafe;
  color: #1e40af;
}
.mount-mode-badge--virtual {
  background: #fef3c7;
  color: #92400e;
}
[data-theme="dark"] .mount-mode-badge--real {
  background: rgba(6,95,70,0.3);
  color: #6ee7b7;
}
[data-theme="dark"] .mount-mode-badge--inherit {
  background: rgba(30,64,175,0.3);
  color: #93c5fd;
}
[data-theme="dark"] .mount-mode-badge--virtual {
  background: rgba(146,64,14,0.3);
  color: #fcd34d;
}

.quota-provider-card__body {
  display: flex;
  gap: 16px;
  align-items: center;
}

.quota-provider-card__chart-wrap {
  flex-shrink: 0;
  width: 72px;
  height: 72px;
}

.donut {
  width: 100%;
  height: 100%;
}

.donut__bg {
  fill: none;
  stroke: var(--border);
  stroke-width: 4;
}

.donut__fill {
  fill: none;
  stroke-width: 4;
  stroke-linecap: round;
  transition: stroke-dasharray 0.4s ease;
}

.donut__text {
  fill: var(--text);
  font-size: 7.5px;
  text-anchor: middle;
  dominant-baseline: central;
  font-weight: 700;
}

.quota-provider-card__stats {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.qstat {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
}

.qstat__label {
  font-size: 12px;
  color: var(--muted);
  flex-shrink: 0;
}

.qstat__value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  text-align: right;
}

.qstat__value--available {
  color: #10b981;
}

.btn--sm {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  transition: all 0.2s;
  white-space: nowrap;
}
.btn--sm:hover { background: var(--surface-strong); }

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

@media (max-width: 860px) {
  .grid--metrics {
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .action-card {
    padding: 16px;
  }

  .action-card__icon {
    font-size: 26px;
  }

  .qstat__value {
    font-size: 13px;
  }

  .quota-expand {
    padding: 14px;
  }

  .quota-grid {
    grid-template-columns: 1fr;
  }

  .quota-provider-card__body {
    flex-direction: column;
    align-items: stretch;
  }

  .quota-provider-card__chart-wrap {
    align-self: center;
  }
}
</style>
