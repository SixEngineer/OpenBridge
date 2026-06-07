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

const THEME_KEY = 'openbridge_theme'
const isDark = ref(localStorage.getItem(THEME_KEY) === 'dark')

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.dataset.theme = isDark.value ? 'dark' : ''
  localStorage.setItem(THEME_KEY, isDark.value ? 'dark' : 'light')
}

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
    // 未登录时不显示"已连接"（OpenList 跑了但没登录 OpenBridge 没意义）
    openListStatus.value = store.isLoggedIn && (res.code === 1000 || res.code === 0) ? 'connected' : 'disconnected'
  } catch (e) {
    openListStatus.value = 'disconnected'
  }
}

onMounted(() => {
  if (isDark.value) {
    document.documentElement.dataset.theme = 'dark'
  }
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
      <button class="theme-toggle" @click="toggleTheme" :title="isDark ? '切换亮色模式' : '切换暗色模式'">
        <svg v-if="isDark" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
        <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
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
  background: var(--surface);
  border-bottom: 1px solid var(--border);
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
  background: var(--surface);
  border-color: var(--border);
  color: var(--muted);
}

.status-indicator--connected {
  background: rgba(16, 185, 129, 0.1);
  border-color: rgba(16, 185, 129, 0.3);
  color: #10b981;
}

.status-indicator--disconnected {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
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
  background: var(--muted);
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  white-space: nowrap;
}

.lang-switcher,
.theme-toggle {
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  transition: all 0.2s;
}
.lang-switcher:hover,
.theme-toggle:hover {
  background: var(--surface-strong);
  border-color: var(--muted);
}

.theme-toggle {
  display: flex;
  align-items: center;
  line-height: 1;
}

[data-theme="dark"] .topbar {
  background: #1a2a42;
  border-bottom-color: rgba(255,255,255,0.08);
}
[data-theme="dark"] .topbar__eyebrow {
  color: #5bc0be;
}
[data-theme="dark"] .topbar__title {
  color: #94a3b8;
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
