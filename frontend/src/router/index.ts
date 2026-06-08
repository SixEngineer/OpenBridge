import { createRouter, createWebHistory } from 'vue-router'
import { useConsoleStore } from '@/stores/console'

if (typeof window !== 'undefined' && 'scrollRestoration' in window.history) {
  window.history.scrollRestoration = 'manual'
}

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    }
    return { top: 0 }
  },
  routes: [
    {
      path: '/',
      component: () => import('@/views/PortalView.vue'),
      meta: { public: true },
    },
    {
      path: '/login',
      component: () => import('@/views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('@/components/layout/AppShell.vue'),
      children: [
        { path: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'openlist', component: () => import('@/views/OpenListView.vue') },
        { path: 'providers', component: () => import('@/views/ProviderView.vue') },
        { path: 'tasks', component: () => import('@/views/DownloadTasksView.vue') },
        { path: 'quota', component: () => import('@/views/QuotaView.vue'), meta: { adminOnly: true } },
        { path: 'rclone', component: () => import('@/views/RcloneView.vue'), meta: { adminOnly: true } },
        { path: 'settings', component: () => import('@/views/SettingsView.vue'), meta: { adminOnly: true } },
        { path: 'debug', component: () => import('@/views/DebugView.vue'), meta: { adminOnly: true } },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.public) return true

  const store = useConsoleStore()
  const valid = await store.validateSession({ forceRemote: true, touch: true })
  if (!valid) return '/login'

  if (store.userRole === null) {
    await store.fetchCurrentUser()
  }

  if (to.meta.adminOnly && !store.isAdmin) {
    return '/dashboard'
  }

  return true
})

export default router
