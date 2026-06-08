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
import { getSessionStatus, getUserInfo, type SessionStatus } from '@/api/user'
import type { Router } from 'vue-router'
import {
  clearLocalSession,
  clearLogoutReason,
  ensureDeviceId,
  isSessionExpired,
  readLocalSession,
  readLogoutReason,
  readSessionTimeoutMinutes,
  writeLocalSession,
  writeLogoutReason,
  writeSessionTimeoutMinutes,
  type LocalSession,
} from '@/utils/session'

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

  interface EffectiveProviderQuota {
    total: number
    used: number
    available: number
    mode: 'real' | 'virtual'
    providerType: string
  }

  const mountCreating = ref(false)

  // ── Auth state ──
  const localSession = ref<LocalSession | null>(readLocalSession())
  const sessionTimeoutMinutes = ref(readSessionTimeoutMinutes())
  const logoutReason = ref(readLogoutReason())
  const isLoggedIn = computed(() => localSession.value !== null)
  const currentUser = computed(() => localSession.value?.username ?? '')
  const currentOpenListBaseURL = computed(() => localSession.value?.openListBaseURL ?? '')
  const openListSessionKey = computed(() => {
    const session = localSession.value
    if (!session) return ''
    return `${session.username}|${session.openListBaseURL}|${session.backendInstanceId}`
  })
  const lastSessionCheckAt = ref(0)
  let sessionMonitorStarted = false
  let sessionMonitorTimer: number | null = null

  function applySession(session: LocalSession | null) {
    localSession.value = session
    if (session) {
      writeLocalSession(session)
    } else {
      clearLocalSession()
    }
  }

  function touchSessionActivity(now = Date.now()) {
    if (!localSession.value) return
    applySession({
      ...localSession.value,
      deviceId: localSession.value.deviceId,
      lastActiveAt: now,
      timeoutMinutes: sessionTimeoutMinutes.value,
    })
  }

  function login(username: string, status: SessionStatus) {
    const now = Date.now()
    const deviceId = status.device_id || ensureDeviceId()
    clearLogoutReason()
    logoutReason.value = ''
    applySession({
      username,
      deviceId,
      issuedAt: now,
      lastActiveAt: now,
      timeoutMinutes: sessionTimeoutMinutes.value,
      backendFingerprint: status.fingerprint,
      backendInstanceId: status.backend_instance_id,
      openListBaseURL: status.openlist_base_url,
    })
    lastSessionCheckAt.value = now
  }

  function logout(reason = 'manual') {
    applySession(null)
    logoutReason.value = reason
    writeLogoutReason(reason)
    lastSessionCheckAt.value = 0
    userRole.value = null
    providers.value = []
    allMounts.value = []
    mounts.value = []
  }

  function setSessionTimeout(minutes: number) {
    writeSessionTimeoutMinutes(minutes)
    sessionTimeoutMinutes.value = readSessionTimeoutMinutes()
    if (localSession.value) {
      applySession({
        ...localSession.value,
        timeoutMinutes: sessionTimeoutMinutes.value,
      })
    }
  }

  // ── UI state ──
  const sidebarOpen = ref(false)
  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
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

  async function fetchSessionStatus() {
    const res = await getSessionStatus()
    return res.data
  }

  async function validateSession(options: { forceRemote?: boolean; touch?: boolean } = {}) {
    const session = localSession.value
    if (!session) return false

    const now = Date.now()
    if (isSessionExpired(session, now)) {
      logout('expired')
      return false
    }

    if (options.touch) {
      touchSessionActivity(now)
    }

    const shouldCheckRemote = options.forceRemote || now-lastSessionCheckAt.value > 30_000
    if (!shouldCheckRemote) {
      return true
    }

    try {
      const status = await fetchSessionStatus()
      if (
        !status.authenticated ||
        status.device_id !== session.deviceId ||
        status.openlist_base_url !== session.openListBaseURL ||
        (status.username && status.username !== session.username)
      ) {
        logout(status.reason || 'session_changed')
        return false
      }

      applySession({
        ...session,
        backendFingerprint: status.fingerprint,
        backendInstanceId: status.backend_instance_id,
        openListBaseURL: status.openlist_base_url,
        lastActiveAt: now,
      })
      lastSessionCheckAt.value = now
      return true
    } catch (error) {
      console.error('校验会话失败', error)
      return true
    }
  }

  function consumeLogoutReason() {
    const reason = logoutReason.value
    logoutReason.value = ''
    clearLogoutReason()
    return reason
  }

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

  async function bootstrapSession() {
    if (!localSession.value) return
    const valid = await validateSession({ forceRemote: true })
    if (!valid) return
    await Promise.allSettled([
      fetchCurrentUser(),
      fetchAllMounts(),
      fetchProviders(),
    ])
  }

  function startSessionMonitor(router: Router) {
    if (sessionMonitorStarted) return
    sessionMonitorStarted = true

    const ensureValidSession = async (forceRemote = false) => {
      if (!localSession.value) return
      const valid = await validateSession({ forceRemote, touch: true })
      if (!valid && router.currentRoute.value.path !== '/login') {
        router.replace('/login')
      }
    }

    void bootstrapSession()
    void ensureValidSession(true)

    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') {
        void ensureValidSession(true)
      }
    })

    window.addEventListener('storage', (event) => {
      if (event.key === null) return
      if (event.key === 'openbridge_auth' || event.key === 'openbridge_session_timeout_minutes') {
        localSession.value = readLocalSession()
        sessionTimeoutMinutes.value = readSessionTimeoutMinutes()
      }
    })

    sessionMonitorTimer = window.setInterval(() => {
      void ensureValidSession(true)
    }, 30_000)
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

  function getEffectiveProviderQuota(providerOrId: ProviderRecord | number): EffectiveProviderQuota | null {
    const provider = typeof providerOrId === 'number'
      ? providers.value.find(p => p.id === providerOrId)
      : providerOrId

    if (!provider) return null

    const virtualMount = allMounts.value
      .filter(m => m.provider_account_id === provider.id && m.enabled && m.quota_mode === 'virtual')
      .sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())[0]

    if (virtualMount) {
      return {
        total: virtualMount.virtual_total,
        used: virtualMount.virtual_used,
        available: Math.max(virtualMount.virtual_total - virtualMount.virtual_used, 0),
        mode: 'virtual',
        providerType: provider.provider_type,
      }
    }

    return {
      total: provider.total_quota,
      used: provider.used_quota,
      available: provider.available_quota,
      mode: 'real',
      providerType: provider.provider_type,
    }
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
    mountPath: string,
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
        mount_path: mountPath.trim(),
        provider_root_path: rootPath || '/',
        quota_mode: quotaMode as 'real' | 'virtual',
      }
      if (quotaMode === 'virtual' && virtualTotal !== undefined) {
        payload.virtual_total = virtualTotal
        payload.virtual_used = 0
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
    getEffectiveProviderQuota,
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
    currentOpenListBaseURL,
    openListSessionKey,
    login,
    logout,
    validateSession,
    fetchSessionStatus,
    startSessionMonitor,
    sessionTimeoutMinutes,
    setSessionTimeout,
    logoutReason,
    consumeLogoutReason,
    userRole,
    isAdmin,
    fetchCurrentUser,
    defaultDownloadDir,
    setDefaultDownloadDir,
    sidebarOpen,
    toggleSidebar,
  }
})
