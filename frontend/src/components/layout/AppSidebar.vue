<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { NavItem } from '@/types/common'

const route = useRoute()

const items: NavItem[] = [
  { label: 'Dashboard', path: '/dashboard', description: 'System overview' },
  { label: 'OpenList', path: '/openlist', description: 'Bridge source access' },
  { label: 'Providers', path: '/providers', description: 'Adapter registry' },
  { label: 'Tasks', path: '/tasks', description: 'Download control center' },
  { label: 'Quota', path: '/quota', description: 'Capacity monitoring' },
  { label: 'Settings', path: '/settings', description: 'Policy and configuration' },
  { label: 'Debug', path: '/debug', description: 'Trace and diagnostics' },
]

const activePath = computed(() => route.path)
</script>

<template>
  <aside class="sidebar">
    <div>
      <div class="brand">
        <span class="brand__mark">OB</span>
        <div>
          <p class="brand__title">OpenBridge</p>
          <p class="brand__subtitle">Download orchestration console</p>
        </div>
      </div>

      <nav class="nav">
        <RouterLink
          v-for="item in items"
          :key="item.path"
          :to="item.path"
          class="nav__link"
          :class="{ 'nav__link--active': activePath === item.path }"
        >
          <span class="nav__label">{{ item.label }}</span>
          <span class="nav__description">{{ item.description }}</span>
        </RouterLink>
      </nav>
    </div>

    <div class="sidebar__footer">
      <p class="sidebar__eyebrow">Focus</p>
      <p class="sidebar__note">Make the platform observable before the backend is complete.</p>
    </div>
  </aside>
</template>
