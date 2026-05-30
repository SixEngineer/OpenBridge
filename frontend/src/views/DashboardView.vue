<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConsoleStore } from '@/stores/console'
import { getDrivers } from '@/api/storage'
import type { MetricCardData } from '@/types/dashboard'

const store = useConsoleStore()

const openListStatus = ref<'active' | 'error'>('disabled')
const openListDetail = ref('Not tested')

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
  if (store.providers.length > 0) count++
  if (openListStatus.value === 'active') count++
  if (store.currentQuota) count++
  return count
})

async function checkOpenList() {
  try {
    const res = await getDrivers()
    if (res.code === 1000) {
      openListStatus.value = 'active'
      openListDetail.value = `Connected — ${res.data?.length || 0} driver(s) available`
    } else {
      openListStatus.value = 'error'
      openListDetail.value = res.msg as string || 'API error'
    }
  } catch (e: any) {
    openListStatus.value = 'error'
    openListDetail.value = e?.message || 'Connection failed'
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
                {{ store.providers.length > 0 ? 'Responding — data loaded successfully' : 'No data yet' }}
              </p>
            </div>
            <StatusBadge :state="store.providers.length > 0 ? 'active' : 'disabled'" />
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
          <h3>Recent Alerts</h3>
          <p>Signals worth surfacing during demos and later backend integration.</p>
        </div>
        <div class="alert-list">
          <article
            v-for="item in store.alerts"
            :key="item.title"
            class="alert-card"
            :class="`alert-card--${item.level}`"
          >
            <p class="alert-card__title">{{ item.title }}</p>
            <p class="alert-card__detail">{{ item.detail }}</p>
          </article>
        </div>
      </section>
    </div>

    <section class="panel">
      <div class="panel__header">
        <h3>Recent Tasks</h3>
        <p>Download orchestration at a glance.</p>
      </div>
      <div class="task-digest-list">
        <article v-for="task in store.tasks" :key="task.id" class="task-digest">
          <div>
            <p class="task-digest__name">{{ task.name }}</p>
            <p class="task-digest__meta">{{ task.id }} · {{ task.provider }}</p>
          </div>
          <div class="task-digest__right">
            <span class="task-digest__status">{{ task.status }}</span>
            <div class="progress">
              <div class="progress__bar" :style="{ width: `${task.progress}%` }"></div>
            </div>
            <span class="task-digest__progress">{{ task.progress }}%</span>
          </div>
        </article>
      </div>
    </section>
  </section>
</template>
