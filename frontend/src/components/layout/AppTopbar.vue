<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDrivers } from '@/api/storage'

const openListStatus = ref<'checking' | 'connected' | 'disconnected'>('checking')
const driverCount = ref(0)

async function checkOpenList() {
  try {
    const res = await getDrivers()
    if (res.code === 1000 || res.code === 0) {
      openListStatus.value = 'connected'
      driverCount.value = res.data?.length || 0
    } else {
      openListStatus.value = 'disconnected'
    }
  } catch (e) {
    openListStatus.value = 'disconnected'
  }
}

onMounted(() => {
  checkOpenList()
  // 每30秒检查一次连接状态
  setInterval(checkOpenList, 30000)
})
</script>

<template>
  <header class="topbar">
    <div>
      <p class="topbar__eyebrow">OpenBridge Console</p>
      <h1 class="topbar__title">Unified control surface for providers, quota, and tasks.</h1>
    </div>

    <div class="topbar__meta">
      <div class="status-indicator" :class="`status-indicator--${openListStatus}`">
        <span class="status-dot"></span>
        <span class="status-text">
          <template v-if="openListStatus === 'checking'">Checking OpenList...</template>
          <template v-else-if="openListStatus === 'connected'">
            OpenList Connected ({{ driverCount }} drivers)
          </template>
          <template v-else>OpenList Disconnected</template>
        </span>
      </div>
    </div>
  </header>
</template>

<style scoped>
.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid;
  transition: all 0.2s;
}

.status-indicator--checking {
  background: #f3f4f6;
  border-color: #d1d5db;
  color: #6b7280;
}

.status-indicator--connected {
  background: #d1fae5;
  border-color: #10b981;
  color: #065f46;
}

.status-indicator--disconnected {
  background: #fee2e2;
  border-color: #ef4444;
  color: #991b1b;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 2s infinite;
}

.status-indicator--connected .status-dot {
  background: #10b981;
}

.status-indicator--disconnected .status-dot {
  background: #ef4444;
}

.status-indicator--checking .status-dot {
  background: #9ca3af;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-text {
  white-space: nowrap;
}
</style>
