<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useConsoleStore } from '@/stores/console'

const router = useRouter()
const store = useConsoleStore()

function handleLoginClick() {
  router.push('/login')
}

function handleLogout() {
  store.logout()
  router.push('/')
}
</script>

<template>
  <header class="topbar">
    <div>
      <p class="topbar__eyebrow">OpenBridge Console</p>
      <h1 class="topbar__title">Unified control surface for providers, quota, and tasks.</h1>
    </div>

    <div class="topbar__meta">
      <template v-if="store.isLoggedIn">
        <div class="topbar__pill topbar__pill--user">
          <span class="user-avatar">{{ store.currentUser.charAt(0).toUpperCase() }}</span>
          {{ store.currentUser }}
        </div>
        <button class="topbar__btn" @click="handleLogout">Logout</button>
      </template>
      <template v-else>
        <button class="topbar__btn topbar__btn--primary" @click="handleLoginClick">Login</button>
      </template>
    </div>
  </header>
</template>

<style scoped>
.topbar__btn {
  padding: 8px 18px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.2s;
  background: transparent;
  color: #374151;
  border-color: #d1d5db;
}

.topbar__btn:hover {
  background: #f3f4f6;
}

.topbar__btn--primary {
  background: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.topbar__btn--primary:hover {
  background: #2563eb;
}

.topbar__pill--user {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.user-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #3b82f6;
  color: white;
  font-size: 13px;
  font-weight: 700;
}
</style>
