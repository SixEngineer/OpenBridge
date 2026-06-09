<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConsoleStore } from '@/stores/console'
import type { MetricCardData } from '@/types/dashboard'
import { getProviderList } from '@/api/provider'
import { getAria2Status } from '@/api/task'
import { getSystemMetrics, type SystemMetrics } from '@/api/system'

const store = useConsoleStore()
const { t, locale } = useI18n()

const aria2Status = ref<'active' | 'error' | 'disabled'>('disabled')
const quotaExpanded = ref(false)
const mountsExpanded = ref(false)
const quotasByProvider = computed(() =>
  store.providers.reduce<Array<{
    id: number
    name: string
    provider_type: string
    net_disk: string
    account_id: string
    status: string
    total_quota: number
    used_quota: number
    available_quota: number
    quota_mode: 'real' | 'virtual'
    last_quota_sync_at?: string
    last_error?: string
    created_at: string
    updated_at: string
  }>>((list, p) => {
      const effectiveQuota = store.getEffectiveProviderQuota(p)
      if (!effectiveQuota) return list

      if (effectiveQuota.mode === 'virtual') {
        list.push({
          ...p,
          total_quota: effectiveQuota.total,
          used_quota: effectiveQuota.used,
          available_quota: effectiveQuota.available,
          quota_mode: effectiveQuota.mode,
        })
        return list
      }

      const scale = p.provider_type === 'mock' ? 1024 : 1
      list.push({
        ...p,
        total_quota: effectiveQuota.total * scale,
        used_quota: effectiveQuota.used * scale,
        available_quota: effectiveQuota.available * scale,
        quota_mode: effectiveQuota.mode,
      })
      return list
    }, []).filter(p => p.total_quota > 0)
)
const aria2Detail = ref(t('dashboard.checking_connection'))
const backendStatus = ref<'active' | 'error' | 'disabled'>('active')
const backendDetail = ref(t('dashboard.backend_api_connected'))
const systemMetrics = ref<SystemMetrics | null>(null)
const systemMetricsError = ref('')
let metricsTimer: ReturnType<typeof setInterval> | null = null
const METRICS_REFRESH_INTERVAL_MS = 1000
let metricsRefreshInFlight = false

const primaryProviderId = ref<number | null>(null)

const primaryProvider = computed(() => {
  if (primaryProviderId.value !== null) {
    return quotasByProvider.value.find(p => p.id === primaryProviderId.value) ?? null
  }
  return quotasByProvider.value[0] ?? null
})

const aggregateQuota = computed(() => {
  return quotasByProvider.value.reduce((summary, provider) => {
    summary.total += provider.total_quota
    summary.used += provider.used_quota
    summary.available += provider.available_quota
    return summary
  }, {
    total: 0,
    used: 0,
    available: 0,
  })
})

const aggregateQuotaPercent = computed(() => {
  if (aggregateQuota.value.total <= 0) return 0
  return Math.min((aggregateQuota.value.used / aggregateQuota.value.total) * 100, 100)
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

function formatSystemBytes(bytes: number): string {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatTransferRate(bytesPerSecond: number): string {
  return `${formatSystemBytes(bytesPerSecond)}/s`
}

const normalizedSystemMetrics = computed(() => {
  if (!systemMetrics.value) return null
  const raw = systemMetrics.value
  return {
    cpuUsage: Number(raw.cpu_usage ?? 0),
    processCPUUsage: Number(raw.process_cpu_usage ?? 0),
    memoryUsage: Number(raw.memory_usage ?? 0),
    memoryUsedBytes: Number(raw.memory_used_bytes ?? 0),
    memoryTotalBytes: Number(raw.memory_total_bytes ?? 0),
    processMemoryBytes: Number(raw.process_memory_bytes ?? 0),
    diskUsage: Number(raw.disk_usage ?? 0),
    diskUsedBytes: Number(raw.disk_used_bytes ?? 0),
    diskTotalBytes: Number(raw.disk_total_bytes ?? 0),
    appDiskUsageBytes: Number(raw.app_disk_usage_bytes ?? 0),
    networkReceiveBPS: Number(raw.network_receive_bytes_per_sec ?? 0),
    networkTransmitBPS: Number(raw.network_transmit_bytes_per_sec ?? 0),
    diskPath: raw.disk_path || 'C:',
    hostname: raw.hostname || '',
    sampledAt: raw.sampled_at || '',
  }
})

const hostMetricCards = computed(() => {
  if (!normalizedSystemMetrics.value) return []
  const metrics = normalizedSystemMetrics.value
  const sampledTime = metrics.sampledAt
    ? new Date(metrics.sampledAt).toLocaleTimeString(locale.value)
    : '—'
  return [
    {
      key: 'cpu',
      title: t('dashboard.cpu_usage'),
      percent: metrics.cpuUsage,
      appPercent: metrics.processCPUUsage,
      ringClass: 'host-donut__fill--cpu',
      appRingClass: 'host-donut__inner--cpu',
      detail: t('dashboard.cpu_usage_detail'),
      totalValue: `${metrics.cpuUsage.toFixed(1)}%`,
      appValue: `${metrics.processCPUUsage.toFixed(1)}%`,
      footnote: t('dashboard.sampled_at', { time: sampledTime }),
    },
    {
      key: 'memory',
      title: t('dashboard.memory_usage'),
      percent: metrics.memoryUsage,
      appPercent: metrics.memoryTotalBytes > 0
        ? Math.min((metrics.processMemoryBytes / metrics.memoryTotalBytes) * 100, 100)
        : 0,
      ringClass: 'host-donut__fill--memory',
      appRingClass: 'host-donut__inner--memory',
      detail: `${formatSystemBytes(metrics.memoryUsedBytes)} / ${formatSystemBytes(metrics.memoryTotalBytes)}`,
      totalValue: `${formatSystemBytes(metrics.memoryUsedBytes)} / ${formatSystemBytes(metrics.memoryTotalBytes)}`,
      appValue: formatSystemBytes(metrics.processMemoryBytes),
      footnote: t('dashboard.memory_usage_detail'),
    },
    {
      key: 'disk',
      title: t('dashboard.disk_usage'),
      percent: metrics.diskUsage,
      appPercent: metrics.diskTotalBytes > 0
        ? Math.min((metrics.appDiskUsageBytes / metrics.diskTotalBytes) * 100, 100)
        : 0,
      ringClass: 'host-donut__fill--disk',
      appRingClass: 'host-donut__inner--disk',
      detail: `${formatSystemBytes(metrics.diskUsedBytes)} / ${formatSystemBytes(metrics.diskTotalBytes)}`,
      totalValue: `${formatSystemBytes(metrics.diskUsedBytes)} / ${formatSystemBytes(metrics.diskTotalBytes)}`,
      appValue: formatSystemBytes(metrics.appDiskUsageBytes),
      footnote: t('dashboard.disk_usage_detail', { path: metrics.diskPath }),
    },
  ]
})

const networkSummary = computed(() => {
  if (!normalizedSystemMetrics.value) return null
  return {
    download: formatTransferRate(normalizedSystemMetrics.value.networkReceiveBPS),
    upload: formatTransferRate(normalizedSystemMetrics.value.networkTransmitBPS),
  }
})

function ringDashArray(percent: number) {
  const normalized = Math.max(0, Math.min(percent, 100))
  return `${normalized} ${100 - normalized}`
}

const quickActions = computed(() => {
  const actions = [
    {
      to: '/providers',
      icon: 'registry',
      title: t('dashboard.manage_providers'),
      desc: t('dashboard.manage_providers_desc'),
    },
    {
      to: '/openlist',
      icon: 'files',
      title: t('dashboard.browse_files'),
      desc: t('dashboard.browse_files_desc'),
    },
    {
      to: '/tasks',
      icon: 'download',
      title: t('dashboard.download_tasks'),
      desc: t('dashboard.download_tasks_desc'),
    },
  ]

  if (store.isAdmin) {
    actions.splice(1, 0, {
      to: '/quota',
      icon: 'quota',
      title: t('dashboard.view_quota'),
      desc: t('dashboard.view_quota_desc'),
    })
  }

  return actions
})

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

async function fetchHostMetrics() {
  if (metricsRefreshInFlight) return
  metricsRefreshInFlight = true
  try {
    const res = await getSystemMetrics()
    systemMetrics.value = res.data
    systemMetricsError.value = ''
  } catch (error: any) {
    systemMetricsError.value = error?.message || t('common.request_error')
  } finally {
    metricsRefreshInFlight = false
  }
}

function startMetricsPolling() {
  stopMetricsPolling()
  metricsTimer = setInterval(() => {
    void fetchHostMetrics()
  }, METRICS_REFRESH_INTERVAL_MS)
}

function stopMetricsPolling() {
  if (metricsTimer) {
    clearInterval(metricsTimer)
    metricsTimer = null
  }
}

onMounted(() => {
  void Promise.allSettled([
    store.fetchProviders(),
    store.fetchAllMounts(),
    checkBackend(),
    checkAria2(),
    fetchHostMetrics(),
  ])
  startMetricsPolling()
})

onUnmounted(() => {
  stopMetricsPolling()
})
</script>

<template>
  <section class="page page--dashboard">
    <div class="page-ornament page-ornament--halo" aria-hidden="true"></div>
    <div class="page-ornament page-ornament--mesh" aria-hidden="true"></div>
    <PageHeader
      :title="t('dashboard.title')"
      :description="t('dashboard.description')"
    />

    <div class="grid grid--metrics">
      <MetricCard
        v-for="(item, idx) in metricCards"
        :key="item.title"
        :item="item"
        @click="idx === 1 && backendStatus === 'active' ? (mountsExpanded = !mountsExpanded, quotaExpanded = false) : undefined"
      />
    </div>

    <section class="storage-hero panel">
      <div class="panel__header storage-hero__header">
        <div>
          <h3>{{ t('dashboard.storage_overview') }}</h3>
          <p>{{ t('dashboard.storage_overview_desc') }}</p>
        </div>
        <button
          v-if="quotasByProvider.length > 0"
          class="btn--sm"
          @click="quotaExpanded = !quotaExpanded; mountsExpanded = false"
        >
          {{ quotaExpanded ? t('common.collapse') : t('dashboard.storage_by_provider') }}
        </button>
      </div>
      <div class="storage-hero__body">
        <div class="storage-hero__ring-wrap">
          <svg viewBox="0 0 42 42" class="storage-hero__ring">
            <path
              class="storage-hero__ring-bg"
              d="M21 2.5 a 18.5 18.5 0 0 0 0 37 a 18.5 18.5 0 0 0 0 -37"
            />
            <path
              class="storage-hero__ring-fill"
              :stroke-dasharray="ringDashArray(aggregateQuotaPercent)"
              d="M21 2.5 a 18.5 18.5 0 0 0 0 37 a 18.5 18.5 0 0 0 0 -37"
            />
            <text x="21" y="18.2" class="storage-hero__ring-percent">{{ aggregateQuotaPercent.toFixed(1) }}%</text>
            <text x="21" y="24.2" class="storage-hero__ring-label">{{ t('dashboard.storage_total') }}</text>
          </svg>
        </div>
        <div class="storage-hero__stats">
          <div class="storage-hero__primary">
            <p class="storage-hero__eyebrow">{{ t('dashboard.storage_total_space') }}</p>
            <h4>
              {{ aggregateQuota.total > 0 ? formatBytes(aggregateQuota.total) : '—' }}
            </h4>
            <p class="storage-hero__detail">
              {{ aggregateQuota.total > 0
                ? t('dashboard.storage_used_detail_available', {
                  used: formatBytes(aggregateQuota.used),
                  total: formatBytes(aggregateQuota.total),
                })
                : t('dashboard.storage_used_detail_empty') }}
            </p>
          </div>
          <div class="storage-hero__stat-grid">
            <div class="storage-chip">
              <span class="storage-chip__label">{{ t('quota.used') }}</span>
              <strong>{{ formatBytes(aggregateQuota.used) }}</strong>
            </div>
            <div class="storage-chip">
              <span class="storage-chip__label">{{ t('quota.available') }}</span>
              <strong>{{ formatBytes(aggregateQuota.available) }}</strong>
            </div>
            <div class="storage-chip">
              <span class="storage-chip__label">{{ t('dashboard.storage_total_space') }}</span>
              <strong>{{ formatBytes(aggregateQuota.total) }}</strong>
            </div>
            <div class="storage-chip">
              <span class="storage-chip__label">{{ t('dashboard.active_providers') }}</span>
              <strong>{{ quotasByProvider.length }}</strong>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Expanded per-provider quota breakdown -->
    <transition name="slide">
      <div v-if="quotaExpanded && quotasByProvider.length > 0" class="quota-expand">
        <div class="quota-expand__header">
          <h4>{{ t('dashboard.storage_by_provider') }}</h4>
          <p>{{ t('dashboard.storage_by_provider_desc') }}</p>
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

    <section class="panel host-panel">
      <div class="panel__header">
        <div>
          <h3>{{ t('dashboard.host_metrics') }}</h3>
          <p>
            {{ normalizedSystemMetrics?.hostname || t('dashboard.host_metrics_desc') }}
          </p>
        </div>
        <button class="btn--sm" @click="fetchHostMetrics">{{ t('common.refresh') }}</button>
      </div>
      <div v-if="networkSummary" class="network-strip">
        <div class="network-pill network-pill--download">
          <span class="network-pill__label">{{ t('dashboard.network_download') }}</span>
          <strong>{{ networkSummary.download }}</strong>
        </div>
        <div class="network-pill network-pill--upload">
          <span class="network-pill__label">{{ t('dashboard.network_upload') }}</span>
          <strong>{{ networkSummary.upload }}</strong>
        </div>
      </div>
      <div v-if="hostMetricCards.length > 0" class="host-metrics-grid">
        <article v-for="item in hostMetricCards" :key="item.key" class="host-card">
          <div class="host-card__ring">
            <svg viewBox="0 0 36 36" class="host-donut">
              <path
                class="host-donut__bg"
                d="M18 2.0845 a 15.9155 15.9155 0 0 0 0 31.831 a 15.9155 15.9155 0 0 0 0 -31.831"
              />
              <path
                class="host-donut__fill"
                :class="item.ringClass"
                :stroke-dasharray="ringDashArray(item.percent)"
                d="M18 2.0845 a 15.9155 15.9155 0 0 0 0 31.831 a 15.9155 15.9155 0 0 0 0 -31.831"
              />
              <path
                class="host-donut__inner-bg"
                d="M18 5.0845 a 12.9155 12.9155 0 0 0 0 25.831 a 12.9155 12.9155 0 0 0 0 -25.831"
              />
              <path
                class="host-donut__inner"
                :class="item.appRingClass"
                :stroke-dasharray="ringDashArray(item.appPercent)"
                d="M18 5.0845 a 12.9155 12.9155 0 0 0 0 25.831 a 12.9155 12.9155 0 0 0 0 -25.831"
              />
              <text x="18" y="16.5" class="host-donut__text">{{ item.percent.toFixed(1) }}%</text>
              <text x="18" y="22.1" class="host-donut__subtext">OB {{ item.appPercent.toFixed(1) }}%</text>
            </svg>
          </div>
          <div class="host-card__body">
            <p class="host-card__title">{{ item.title }}</p>
            <div class="host-card__legend">
              <span class="host-legend host-legend--total">{{ t('dashboard.host_total_usage') }} · {{ item.totalValue }}</span>
              <span class="host-legend host-legend--app">{{ t('dashboard.openbridge_usage') }} · {{ item.appValue }}</span>
            </div>
            <p class="host-card__detail">{{ item.detail }}</p>
            <p class="host-card__footnote">{{ item.footnote }}</p>
          </div>
        </article>
      </div>
      <p v-else class="host-panel__empty">
        {{ systemMetricsError || t('dashboard.host_metrics_empty') }}
      </p>
    </section>

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
          <router-link v-for="action in quickActions" :key="action.to" :to="action.to" class="action-card">
            <div class="action-card__icon" :class="`action-card__icon--${action.icon}`">
              <span>{{ action.title.slice(0, 1) }}</span>
            </div>
            <h4>{{ action.title }}</h4>
            <p>{{ action.desc }}</p>
          </router-link>
        </div>
      </section>
    </div>
  </section>
</template>


<style scoped>
.page--dashboard {
  position: relative;
  overflow: hidden;
}

.page-ornament {
  position: absolute;
  pointer-events: none;
  z-index: 0;
}

.page-ornament--halo {
  top: -120px;
  right: -60px;
  width: 360px;
  height: 360px;
  border-radius: 50%;
  background:
    radial-gradient(circle at 30% 30%, rgba(15, 118, 110, 0.22), transparent 48%),
    radial-gradient(circle at 65% 55%, rgba(245, 158, 11, 0.18), transparent 42%),
    radial-gradient(circle at 50% 50%, rgba(14, 165, 233, 0.08), transparent 68%);
  filter: blur(10px);
}

.page-ornament--mesh {
  inset: 110px auto auto -120px;
  width: 320px;
  height: 320px;
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.05) 1px, transparent 1px);
  background-size: 24px 24px;
  mask-image: radial-gradient(circle at center, rgba(0, 0, 0, 0.95), transparent 72%);
  opacity: 0.65;
  transform: rotate(-12deg);
}

.page--dashboard > :not(.page-ornament) {
  position: relative;
  z-index: 1;
}

.dashboard-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-top: 8px;
}

.host-panel {
  margin-top: 8px;
}

.page--dashboard :deep(.grid--metrics) {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.storage-hero {
  overflow: hidden;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.97), rgba(234, 245, 247, 0.86)),
    var(--surface);
}

.storage-hero__header {
  margin-bottom: 12px;
}

.storage-hero__body {
  display: grid;
  grid-template-columns: 240px minmax(0, 1fr);
  gap: 16px;
  align-items: center;
}

.storage-hero__ring-wrap {
  display: flex;
  justify-content: center;
  align-items: center;
}

.storage-hero__ring {
  width: 180px;
  height: 180px;
}

.storage-hero__ring-bg {
  fill: none;
  stroke: rgba(148, 163, 184, 0.18);
  stroke-width: 3.2;
}

.storage-hero__ring-fill {
  fill: none;
  stroke: url(#storageRingGradient);
  stroke: #0f766e;
  stroke-width: 3.2;
  stroke-linecap: round;
  filter: drop-shadow(0 10px 16px rgba(15, 118, 110, 0.15));
}

.storage-hero__ring-percent {
  fill: var(--text);
  font-size: 6.1px;
  font-weight: 800;
  text-anchor: middle;
}

.storage-hero__ring-label {
  fill: var(--muted);
  font-size: 2.8px;
  font-weight: 700;
  text-anchor: middle;
  letter-spacing: 0.08em;
}

.storage-hero__stats {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.storage-hero__eyebrow {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.storage-hero__primary h4 {
  margin: 0;
  font-size: clamp(28px, 4vw, 40px);
  line-height: 1;
  color: var(--text);
}

.storage-hero__detail {
  margin: 10px 0 0;
  font-size: 14px;
  color: var(--muted);
}

.storage-hero__stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.storage-chip {
  padding: 14px;
  border-radius: 14px;
  border: 1px solid rgba(37, 99, 235, 0.1);
  background: rgba(255, 255, 255, 0.68);
}

.storage-chip__label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--muted);
}

.storage-chip strong {
  font-size: 16px;
  color: var(--text);
}

.host-metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.network-strip {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.network-pill {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 14px;
  border-radius: 14px;
  border: 1px solid var(--border);
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.9), rgba(246, 250, 251, 0.72)),
    var(--surface);
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.05);
}

.network-pill--download {
  border-color: rgba(37, 99, 235, 0.16);
}

.network-pill--upload {
  border-color: rgba(15, 118, 110, 0.16);
}

.network-pill__label {
  font-size: 12px;
  color: var(--muted);
}

.host-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border-radius: 14px;
  border: 1px solid var(--border);
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.86), rgba(255, 255, 255, 0.72)),
    var(--surface);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(10px);
  animation: card-rise 0.45s ease both;
}

.host-card__ring {
  width: 82px;
  height: 82px;
  flex-shrink: 0;
}

.host-donut {
  width: 100%;
  height: 100%;
}

.host-donut__bg {
  fill: none;
  stroke: var(--border);
  stroke-width: 4;
}

.host-donut__fill {
  fill: none;
  stroke-width: 4;
  stroke-linecap: round;
  transition: stroke-dasharray 0.55s ease;
}

.host-donut__fill--cpu {
  stroke: #0f766e;
}

.host-donut__fill--memory {
  stroke: #2563eb;
}

.host-donut__fill--disk {
  stroke: #c2410c;
}

.host-donut__inner-bg {
  fill: none;
  stroke: rgba(148, 163, 184, 0.22);
  stroke-width: 2.8;
}

.host-donut__inner {
  fill: none;
  stroke-width: 2.8;
  stroke-linecap: round;
  transition: stroke-dasharray 0.55s ease;
}

.host-donut__inner--cpu {
  stroke: #14b8a6;
}

.host-donut__inner--memory {
  stroke: #60a5fa;
}

.host-donut__inner--disk {
  stroke: #fb923c;
}

.host-donut__text {
  color: var(--text);
  fill: var(--text);
  font-size: 5.9px;
  font-weight: 700;
  text-anchor: middle;
  dominant-baseline: central;
}

.host-donut__subtext {
  fill: var(--muted);
  font-size: 3.4px;
  font-weight: 700;
  text-anchor: middle;
  letter-spacing: 0.04em;
}

.host-card__body {
  min-width: 0;
}

.host-card__title {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
}

.host-card__legend {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 8px;
}

.host-legend {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--muted);
}

.host-legend::before {
  content: '';
  width: 10px;
  height: 10px;
  border-radius: 999px;
  flex-shrink: 0;
}

.host-legend--total::before {
  background: linear-gradient(135deg, #0f766e, #2563eb);
}

.host-legend--app::before {
  background: linear-gradient(135deg, #14b8a6, #fb923c);
}

.host-card__detail,
.host-card__footnote,
.host-panel__empty {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
}

.host-card__footnote {
  margin-top: 6px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.action-card {
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(250, 250, 249, 0.74)),
    var(--surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 20px;
  text-decoration: none;
  color: inherit;
  transition: transform 0.28s ease, box-shadow 0.28s ease, border-color 0.28s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  animation: card-rise 0.45s ease both;
}

.action-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.1);
  transform: translateY(-2px);
}

.action-card__icon {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  font-size: 24px;
  font-weight: 800;
  color: white;
  margin-bottom: 12px;
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
}

.action-card__icon--quota {
  background: linear-gradient(135deg, #10b981, #047857);
}

.action-card__icon--files {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.action-card__icon--download {
  background: linear-gradient(135deg, #ef4444, #b91c1c);
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
  align-items: flex-start;
  justify-content: space-between;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}
.quota-expand__header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}
.quota-expand__header p {
  margin: 0;
  font-size: 13px;
  color: var(--muted);
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
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(255, 255, 255, 0.72)),
    var(--surface);
  cursor: pointer;
  transition: transform 0.22s ease, border-color 0.22s ease, box-shadow 0.22s ease;
  animation: card-rise 0.45s ease both;
}

[data-theme="dark"] .storage-hero {
  background:
    linear-gradient(135deg, rgba(22, 39, 59, 0.98), rgba(17, 31, 47, 0.96)),
    var(--surface);
  border-color: rgba(91, 192, 190, 0.2);
}

[data-theme="dark"] .storage-chip {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(91, 192, 190, 0.12);
}

[data-theme="dark"] .network-pill,
[data-theme="dark"] .host-card,
[data-theme="dark"] .quota-provider-card,
[data-theme="dark"] .action-card,
[data-theme="dark"] .quota-expand {
  background:
    linear-gradient(180deg, rgba(23, 37, 56, 0.96), rgba(18, 29, 44, 0.94)),
    var(--surface);
  border-color: rgba(255, 255, 255, 0.09);
}
.quota-provider-card:hover {
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(37, 99, 235, 0.08);
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
  transition: opacity 0.3s ease, transform 0.3s ease;
}
.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.985);
}

@keyframes card-rise {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1024px) {
  .page--dashboard :deep(.grid--metrics) {
    grid-template-columns: 1fr;
  }

  .dashboard-panels {
    grid-template-columns: 1fr;
  }

  .action-grid {
    grid-template-columns: 1fr;
  }

  .host-metrics-grid {
    grid-template-columns: 1fr;
  }

  .storage-hero__body {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .page-ornament--halo {
    width: 240px;
    height: 240px;
    top: -70px;
    right: -80px;
  }

  .page-ornament--mesh {
    width: 220px;
    height: 220px;
    left: -90px;
    top: 130px;
  }

  .dashboard-panels .panel:last-child {
    display: none;
  }

  .page--dashboard :deep(.grid--metrics) {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .storage-hero__ring {
    width: 170px;
    height: 170px;
  }

  .storage-hero__stat-grid {
    grid-template-columns: 1fr;
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

  .host-card__legend {
    gap: 6px;
  }

  .quota-expand {
    padding: 14px;
  }

  .network-strip {
    flex-direction: column;
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
