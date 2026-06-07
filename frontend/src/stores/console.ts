import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { getProviderList, deleteProvider } from '@/api/provider'
import type { ProviderRecord } from '@/types/provider'

import {
  createMount,
  listMounts,
  updateMount,
  deleteMount,
  queryMountQuota,
  syncMountQuota,
} from '@/api/mount'
import type { MountPoint } from '@/types/mount'
import type { QuotaInfo } from '@/types/quota'

import { alertItems, metricCards, systemStatuses, taskDigests } from '@/mock/dashboard'
import { quotaRecords } from '@/mock/quota'

import { createTask } from '@/api/task'
import { getUserInfo } from '@/api/user'

export const useConsoleStore = defineStore('console', () => {
  const metrics = ref(metricCards)
  const statuses = ref(systemStatuses)
  const tasks = ref(taskDigests)
  const alerts = ref(alertItems)
  const quotas = ref(quotaRecords)

  const providers = ref<ProviderRecord[]>([])

  // ── Mount 多挂载点支持 ──

  // 所有挂载点的完整列表（来自后端）
  const allMounts = ref<MountPoint[]>([])

  // 最近一次配额数据（持久化到 localStorage）
  const QUOTA_KEY = 'openbridge_current_quota'
  const QUOTA_MODE_KEY = 'openbridge_current_quota_mode'
  const QUOTA_EXTRA_KEY = 'openbridge_current_quota_extra'
  const storedQuota = (() => {
    try {
      const raw = localStorage.getItem(QUOTA_KEY)
      return raw ? JSON.parse(raw) : null
    } catch { return null }
  })()
  const storedQuotaMode = (() => {
    try {
      return localStorage.getItem(QUOTA_MODE_KEY)
    } catch { return null }
  })()
  const storedQuotaExtra = (() => {
    try {
      const raw = localStorage.getItem(QUOTA_EXTRA_KEY)
      return raw ? JSON.parse(raw) : null
    } catch { return null }
  })()
  const currentQuota = ref<QuotaInfo | null>(storedQuota)
  const currentQuotaMode = ref<string | null>(storedQuotaMode)
  const currentQuotaExtra = ref<{ inherit_chain?: number[]; virtual_config?: Record<string, number> } | null>(storedQuotaExtra)
  const quotaLoading = ref(false)

  // 挂载点映射（provider_account_id → mount_id[]）
  const MOUNT_MAP_KEY = 'openbridge_mount_id_by_provider'
  const storedMountMap = (() => {
    try {
      const raw = localStorage.getItem(MOUNT_MAP_KEY)
      return raw ? JSON.parse(raw) : {}
    } catch { return {} }
  })()
  // 维护一个兼容性映射，但实际使用 allMounts
  const mountIdByProvider = ref<Record<number, number>>(storedMountMap)

  // 挂载点详情列表（完全由 API 驱动，不依赖 localStorage）
  const MOUNTS_KEY = 'openbridge_mounts'
  // 清理旧的 localStorage 数据，避免过期 mount 显示在 Dashboard 展开列表中
  localStorage.removeItem(MOUNTS_KEY)
  interface MountInfo {
    id: number
    name: string
    mode: string
    providerName: string
    providerId: number
  }
  const mounts = ref<MountInfo[]>([])

  const mountCreating = ref(false)

  // ── Auth state ──
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

  // Default download directory
  const DD_KEY = 'openbridge_default_download_dir'
  const defaultDownloadDir = ref(localStorage.getItem(DD_KEY) || '')

  function setDefaultDownloadDir(dir: string) {
    defaultDownloadDir.value = dir
    localStorage.setItem(DD_KEY, dir)
  }

  const userRole = ref<number | null>(null)
  // OpenList 角色: 0=GENERAL, 1=GUEST, 2=ADMIN
  const isAdmin = computed(() => userRole.value === 2)

  async function fetchCurrentUser() {
    try {
      const res = await getUserInfo()
      if (res.code === 1000) {
        userRole.value = res.data.role
      }
    } catch {
      // 静默处理，非管理员用户可能无法访问
    }
  }

  // 页面刷新时，如果已登录则自动获取用户角色和挂载点数据
  if (isLoggedIn.value) {
    fetchCurrentUser()
    fetchAllMounts()
  }

  // ── Provider actions ──

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

  // 持久化 mount 映射
  function saveMountMapping() {
    localStorage.setItem(MOUNT_MAP_KEY, JSON.stringify(mountIdByProvider.value))
  }

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

  // ── Mount actions ──

  /** 从后端加载所有挂载点，按 provider 分组缓存到 allMounts */
  async function fetchAllMounts() {
    try {
      const res = await listMounts()
      if (res.code === 1000) {
        allMounts.value = res.data
        // 同步 mounts 兼容性列表
        rebuildMountInfo()
      }
    } catch (error) {
      console.error('获取挂载点列表失败', error)
    }
  }

  /** 根据 allMounts 重建 mounts 和 mountIdByProvider 兼容数据 */
  function rebuildMountInfo() {
    const newIdMap: Record<number, number> = {}
    // 只保留每个 provider 的最后一个 mount 作为兼容映射值
    for (const m of allMounts.value) {
      newIdMap[m.provider_account_id] = m.id
    }
    mountIdByProvider.value = newIdMap
    mounts.value = allMounts.value.map(m => ({
      id: m.id,
      name: m.name,
      mode: m.quota_mode,
      providerName: m.provider_type,
      providerId: m.provider_account_id,
    }))
    saveMountMapping()
  }

  /** 获取指定 provider 的所有挂载点 */
  function getMountsByProvider(providerId: number): MountPoint[] {
    return allMounts.value.filter(m => m.provider_account_id === providerId)
  }

  /** 持久化最近一次配额数据 */
  function saveQuotaData() {
    if (currentQuota.value) {
      localStorage.setItem(QUOTA_KEY, JSON.stringify(currentQuota.value))
    }
    if (currentQuotaMode.value) {
      localStorage.setItem(QUOTA_MODE_KEY, currentQuotaMode.value)
    } else {
      localStorage.removeItem(QUOTA_MODE_KEY)
    }
    if (currentQuotaExtra.value) {
      localStorage.setItem(QUOTA_EXTRA_KEY, JSON.stringify(currentQuotaExtra.value))
    } else {
      localStorage.removeItem(QUOTA_EXTRA_KEY)
    }
  }

  /** 查询指定挂载点的配额 */
  async function queryQuotaByMount(mountId: number) {
    quotaLoading.value = true
    try {
      const res = await queryMountQuota(mountId)
      if (res.code === 1000) {
        currentQuota.value = res.data.quota
        currentQuotaMode.value = res.data.mode
        currentQuotaExtra.value = {
          inherit_chain: res.data.inherit_chain,
          virtual_config: res.data.virtual_config,
        }
        saveQuotaData()
      }
      return res
    } catch (error) {
      // 静默处理（mount 可能已被删除等正常情况）
      return null
    } finally {
      quotaLoading.value = false
    }
  }

  /** 同步指定挂载点的配额 */
  async function syncQuotaByMount(mountId: number) {
    quotaLoading.value = true
    try {
      const res = await syncMountQuota(mountId)
      if (res.code === 1000) {
        currentQuota.value = res.data.quota
        currentQuotaMode.value = res.data.mode
        currentQuotaExtra.value = {
          inherit_chain: res.data.inherit_chain,
          virtual_config: res.data.virtual_config,
        }
        saveQuotaData()
      }
      return res
    } catch (error) {
      console.error('同步配额失败', error)
      return null
    } finally {
      quotaLoading.value = false
    }
  }

  /** 为指定 provider 创建挂载点 */
  async function createMountForProvider(
    providerId: number,
    providerName: string,
    providerType: string,
    rootPath?: string,
    quotaMode: string = 'real',
    virtualTotal?: number,
    inheritFromId?: number
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
      if (quotaMode === 'inherit' && inheritFromId !== undefined) {
        payload.inherit_from_id = inheritFromId
      }
      const res = await createMount(payload)
      if (res.code === 1000) {
        // 刷新挂载点列表
        await fetchAllMounts()
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

  /** 删除指定挂载点 */
  async function deleteMountById(mountId: number): Promise<boolean> {
    try {
      const res = await deleteMount(mountId)
      if (res.code === 1000) {
        await fetchAllMounts()
        // 如果当前显示的配额就是这个 mount，清空
        return true
      }
      return false
    } catch (error) {
      console.error('删除 Mount 失败', error)
      return false
    }
  }

  /** 更新指定挂载点 */
  async function updateMountById(mountId: number, data: Partial<MountPoint>): Promise<MountPoint | null> {
    try {
      const res = await updateMount(mountId, data)
      if (res.code === 1000) {
        await fetchAllMounts()
        return res.data
      }
      return null
    } catch (error) {
      console.error('更新 Mount 失败', error)
      return null
    }
  }

  // ── Task actions ──

  async function addTask(data: { path: string; dir?: string }) {
    const res = await createTask(data)
    if (res.code === 1000) {
      const taskId = (res.data as any).task_id || (res.data as any).TaskID
      if (taskId) recordTaskId(taskId)
    }
    return res
  }

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
    currentQuotaMode,
    currentQuotaExtra,
    quotaLoading,
    mountIdByProvider,
    mounts,
    mountCreating,
    allMounts,
    getMountsByProvider,
    fetchAllMounts,
    queryQuotaByMount,
    syncQuotaByMount,
    createMountForProvider,
    deleteMountById,
    updateMountById,
    addTask,
    downloadTaskIds,
    recordTaskId,
    removeTaskId,
    isLoggedIn,
    currentUser,
    login,
    logout,
    userRole,
    isAdmin,
    fetchCurrentUser,
    defaultDownloadDir,
    setDefaultDownloadDir,
  }
})
