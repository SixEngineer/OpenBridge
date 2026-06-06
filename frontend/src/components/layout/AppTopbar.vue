<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import { getDrivers } from '@/api/storage'

const router = useRouter()
const store = useConsoleStore()
const { locale, t } = useI18n()

const openListStatus = ref<'checking' | 'connected' | 'disconnected'>('checking')

function toggleLang() {
  locale.value = locale.value === 'en' ? 'zh-CN' : 'en'
  localStorage.setItem('openbridge_lang', locale.value)
}

function handleLogout() {
  store.logout()
  router.push('/login')
}

async function checkOpenList() {
  try {
    const res = await getDrivers()
    if (res.code === 1000 || res.code === 0) {
      openListStatus.value = 'connected'
    } else {
      openListStatus.value = 'disconnected'
    }
  } catch (e) {
    openListStatus.value = 'disconnected'
  }
}

onMounted(() => {
  checkOpenList()
  setInterval(checkOpenList, 30000)
})
</script>

<template>
  <header class="topbar">
    <div>
      <p class="topbar__eyebrow">{{ t('topbar.title') }}</p>
      <h1 class="topbar__title">{{ t('topbar.subtitle') }}</h1>
    </div>

    <div class="topbar__meta">
      <div
        class="status-indicator"
        :class="`status-indicator--${openListStatus}`"
      >
        <span class="status-dot"></span>
        <span class="status-text">
          <template v-if="openListStatus === 'checking'">{{ t('topbar.openlist_checking') }}</template>
          <template v-else-if="openListStatus === 'connected'">
            {{ t('topbar.openlist_connected') }}
          </template>
          <template v-else>{{ t('topbar.openlist_disconnected') }}</template>
        </span>
      </div>
      <button class="lang-switcher" @click="toggleLang" :title="locale === 'en' ? t('topbar.switch_to_cn') : t('topbar.switch_to_en')">
        {{ locale === 'en' ? '中文' : 'EN' }}
      </button>
      <router-link v-if="!store.isLoggedIn" to="/login" class="topbar-login-btn">{{ t('topbar.login') }}</router-link>
      <button v-else class="topbar-logout-btn" @click="handleLogout">{{ t('topbar.logout') }}</button>
    </div>
  </header>
</template>

<style scoped>
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
}

.topbar__eyebrow {
  margin: 0 0 2px;
  font-size: 12px;
  font-weight: 600;
  color: #3b82f6;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.topbar__title {
  margin: 0;
  font-size: 14px;
  font-weight: 400;
  color: #6b7280;
}

.topbar__meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

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
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  white-space: nowrap;
}

.lang-switcher {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: white;
  color: #374151;
  transition: all 0.2s;
}
.lang-switcher:hover {
  background: #f3f4f6;
  border-color: #9ca3af;
}

.topbar-login-btn,
.topbar-logout-btn {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid #3b82f6;
  background: #3b82f6;
  color: white;
  text-decoration: none;
  transition: all 0.2s;
}
.topbar-login-btn:hover,
.topbar-logout-btn:hover {
  background: #2563eb;
  border-color: #2563eb;
}
</style>
