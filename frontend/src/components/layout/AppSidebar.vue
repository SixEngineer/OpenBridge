<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useConsoleStore } from '@/stores/console'
import type { NavItem } from '@/types/common'

const { t } = useI18n()
const route = useRoute()
const store = useConsoleStore()

const allItems: NavItem[] = [
  { label: 'Dashboard', path: '/dashboard', description: 'System overview', i18nKey: 'dashboard', i18nDescKey: 'dashboard_desc' },
  { label: 'OpenList', path: '/openlist', description: 'Bridge source access', i18nKey: 'openlist', i18nDescKey: 'openlist_desc' },
  { label: 'Providers', path: '/providers', description: 'Adapter registry', i18nKey: 'providers', i18nDescKey: 'providers_desc' },
  { label: 'Tasks', path: '/tasks', description: 'Download control center', i18nKey: 'tasks', i18nDescKey: 'tasks_desc' },
  { label: 'Quota', path: '/quota', description: 'Capacity monitoring', i18nKey: 'quota', i18nDescKey: 'quota_desc' },
  { label: 'Settings', path: '/settings', description: 'Policy and configuration', i18nKey: 'settings', i18nDescKey: 'settings_desc' },
  { label: 'Debug', path: '/debug', description: 'Trace and diagnostics', i18nKey: 'debug', i18nDescKey: 'debug_desc' },
]

const items = computed(() => allItems)

const activePath = computed(() => route.path)

function closeOnNav() {
  store.sidebarOpen = false
}
</script>

<template>
  <!-- Mobile overlay -->
  <div
    v-if="store.sidebarOpen"
    class="sidebar-overlay"
    @click="store.sidebarOpen = false"
  ></div>

  <aside class="sidebar" :class="{ 'sidebar--open': store.sidebarOpen }">
    <div>
      <router-link to="/" class="brand" @click="closeOnNav">
        <span class="brand__mark">OB</span>
        <div>
          <p class="brand__title">{{ t('app.name') }}</p>
          <p class="brand__subtitle">{{ t('app.tagline') }}</p>
        </div>
      </router-link>

      <nav class="nav">
        <RouterLink
          v-for="item in items"
          :key="item.path"
          :to="item.path"
          class="nav__link"
          :class="{ 'nav__link--active': activePath === item.path }"
          @click="closeOnNav"
        >
          <span class="nav__label">{{ item.i18nKey ? t('sidebar.items.' + item.i18nKey) : item.label }}</span>
          <span class="nav__description">{{ item.i18nDescKey ? t('sidebar.items.' + item.i18nDescKey) : item.description }}</span>
        </RouterLink>
      </nav>
    </div>
  </aside>
</template>

<style scoped>
.sidebar-overlay {
  display: none;
}

@media (max-width: 860px) {
  .sidebar-overlay {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    z-index: 99;
  }
}
</style>
