import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
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
        { path: 'quota', component: () => import('@/views/QuotaView.vue') },
        { path: 'settings', component: () => import('@/views/SettingsView.vue') },
        { path: 'debug', component: () => import('@/views/DebugView.vue') },
      ],
    },
  ],
})

// Auth guard: unauthenticated users are redirected to /login
const AUTH_KEY = 'openbridge_auth'

router.beforeEach((to) => {
  if (to.meta.public) return true
  const storedAuth = localStorage.getItem(AUTH_KEY)
  if (storedAuth) return true
  return '/login'
})

export default router