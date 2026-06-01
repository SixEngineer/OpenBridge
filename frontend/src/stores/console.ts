import { ref } from 'vue'
import { defineStore } from 'pinia'
import { getProviderList, deleteProvider } from '@/api/provider'
import type { ProviderRecord } from '@/types/provider'

import { createMount, queryMountQuota, syncMountQuota } from '@/api/mount'
import type { MountPoint } from '@/types/mount'
import type { QuotaInfo } from '@/types/quota'

import { alertItems, metricCards, systemStatuses, taskDigests } from '@/mock/dashboard'
import { quotaRecords } from '@/mock/quota'

import { createTask } from '@/api/task'

export const useConsoleStore = defineStore('console', () => {
  const metrics = ref(metricCards)
  const statuses = ref(systemStatuses)
  const tasks = ref(taskDigests)
  const alerts = ref(alertItems)
  const quotas = ref(quotaRecords)

  const providers = ref<ProviderRecord[]>([])

  const currentQuota = ref<QuotaInfo | null>(null)
  const quotaLoading = ref(false)

  // provider_account_id → mount_id 映射
  const mountIdByProvider = ref<Record<number, number>>({})
  const mountCreating = ref(false)

  // Auth state — persisted across page refreshes
  const AUTH_KEY = 'openbridge_auth'
  const storedAuth = (() => {
    try {
      const raw = localStorage.getItem(AUTH_KEY)
      return raw ? JSON.parse(raw) : null
    } catch { return null }
  })()
  const isLoggedIn = ref(storedAuth !== null)
  const currentUser = ref(storedAuth?.username ?? '')

  function login(username: string) {
    isLoggedIn.value = true
    currentUser.value = username
    localStorage.setItem(AUTH_KEY, JSON.stringify({ username }))
  }

  function logout() {
    isLoggedIn.value = false
    currentUser.value = ''
    localStorage.removeItem(AUTH_KEY)
  }

  // 获取 Provider 列表
  async function fetchProviders() {
    try {
      const res = await getProviderList()
      if (res.code === 1000) {
        providers.value = res.data
      }
    } catch (error) {
      console.error('获取 Provider 列表失败', error)
      providers.value = []
    }
  }

  // 删除 Provider
  async function removeProvider(id: number) {
    try {
      const res = await deleteProvider(id)
      if (res.code === 1000) {
        await fetchProviders()
        return true
      }
      return false
    } catch (error) {
      console.error('删除失败', error)
      return false
    }
  }

  // 通过 Mount 查询配额
  async function queryQuotaByMount(mountId: number) {
    quotaLoading.value = true
    try {
      const res = await queryMountQuota(mountId)
      if (res.code === 1000) {
        currentQuota.value = res.data.quota
      }
      return res
    } catch (error) {
      console.error('查询配额失败', error)
      return null
    } finally {
      quotaLoading.value = false
    }
  }

  // 通过 Mount 同步配额
  async function syncQuotaByMount(mountId: number) {
    quotaLoading.value = true
    try {
      const res = await syncMountQuota(mountId)
      if (res.code === 1000) {
        currentQuota.value = res.data.quota
      }
      return res
    } catch (error) {
      console.error('同步配额失败', error)
      return null
    } finally {
      quotaLoading.value = false
    }
  }

  // 为指定 Provider 创建 Mount
  async function createMountForProvider(
    providerId: number,
    providerName: string,
    providerType: string,
    rootPath?: string,
    quotaMode: string = 'real',
    virtualTotal?: number
  ): Promise<MountPoint | null> {
    mountCreating.value = true
    try {
      const payload: Partial<MountPoint> = {
        name: `${providerName} Mount`,
        provider_account_id: providerId,
        provider_type: providerType,
        mount_path: `/mnt/${providerType}`,
        provider_root_path: rootPath || '/',
        quota_mode: quotaMode as 'real' | 'inherit' | 'virtual',
      }
      if (quotaMode === 'virtual' && virtualTotal !== undefined) {
        payload.virtual_total = virtualTotal
        payload.virtual_used = 0
      }
      const res = await createMount(payload)
      if (res.code === 1000) {
        mountIdByProvider.value[providerId] = res.data.id
        return res.data
      }
      return null
    } catch (error) {
      console.error('创建 Mount 失败', error)
      return null
    } finally {
      mountCreating.value = false
    }
  }
  // 创建任务
  async function addTask(data: { path: string; dir?: string }) {
    const res = await createTask(data)
    if (res.code === 1000) {
      const taskId = (res.data as any).task_id || (res.data as any).TaskID
      if (taskId) recordTaskId(taskId)
    }
    return res
  }

  // 下载任务 ID 列表（持久化到 localStorage）
  const STORAGE_KEY = 'openbridge_download_task_ids'
  const downloadTaskIds = ref<string[]>(loadTaskIds())

  function loadTaskIds(): string[] {
    try {
      const raw = localStorage.getItem(STORAGE_KEY)
      return raw ? JSON.parse(raw) : []
    } catch { return [] }
  }

  function saveTaskIds() {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(downloadTaskIds.value))
  }

  function recordTaskId(id: string) {
    if (!downloadTaskIds.value.includes(id)) {
      downloadTaskIds.value.unshift(id)
      saveTaskIds()
    }
  }

  function removeTaskId(id: string) {
    downloadTaskIds.value = downloadTaskIds.value.filter(x => x !== id)
    saveTaskIds()
  }

  return {
    metrics,
    statuses,
    tasks,
    alerts,
    providers,
    quotas,
    fetchProviders,
    removeProvider,
    currentQuota,
    quotaLoading,
    mountIdByProvider,
    mountCreating,
    queryQuotaByMount,
    syncQuotaByMount,
    createMountForProvider,
    addTask,
    downloadTaskIds,
    recordTaskId,
    removeTaskId,
    isLoggedIn,
    currentUser,
    login,
    logout,
  }
})