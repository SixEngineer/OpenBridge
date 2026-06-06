import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/views/PortalView.vue'),
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

// Admin guard removed — Debug page is now visible to all users
// Only the reset-data button inside Debug is gated by store.isAdmin

export default router