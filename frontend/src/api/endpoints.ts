/**
 * 后端 API 端点集中配置
 *
 * 所有接口路径都定义在此文件中。
 * 如果后端接口发生变化，只需修改此文件即可。
 *
 * baseURL 由 VITE_API_BASE_URL 环境变量控制（默认 /api/v1），
 * 各端点只需定义相对于 baseURL 的路径。
 */
const endpoints = {
  // ── Provider ──
  providerList: '/provider/list',
  providerInfo: '/provider/info',
  provider: '/provider',

  // ── Mount ──
  mount: '/mount',
  mountDetail: (id: number) => `/mount/${id}` as const,
  mountQuota: (mountId: number) => `/mount/${mountId}/quota` as const,
  mountQuotaSync: (mountId: number) => `/mount/${mountId}/quota/sync` as const,

  // ── Quota ──
  quotaQuery: '/quota/query',
  quotaSync: '/quota/sync',

  // ── Storage ──
  storageDrivers: '/storage/drivers',
  storageDriverInfo: '/storage/driverInfo',
  storageFiles: '/storage/files',
  storageFile: '/storage/file',

  // ── System ──
  systemPickPath: '/system/pick-path',
  systemMetrics: '/system/metrics',
  systemRestart: '/system/restart',
  systemExit: '/system/exit',

  // ── Download / Task ──
  downloadTasks: '/download/tasks',
  downloadResolve: '/download/resolve',
  downloadFolderZip: '/download/folder-zip',
  downloadTaskDetail: (taskId: string) => `/download/tasks/${taskId}` as const,
  downloadTaskRetry: (taskId: string) => `/download/tasks/${taskId}/retry` as const,
  downloadTaskOpen: (taskId: string) => `/download/tasks/${taskId}/open` as const,
  downloadTaskOpenLocation: (taskId: string) => `/download/tasks/${taskId}/open-location` as const,
  downloadAria2Status: '/download/aria2-status',

  // ── Settings ──
  settings: '/settings',
  settingsApp: '/settings/app',
  settingsOpenList: '/settings/openlist',
  settingsAria2: '/settings/aria2',
  settingsRclone: '/settings/rclone',
  settingsSession: '/settings/session',
  settingsFileTree: '/settings/filetree',

  // ── User ──
  userLogin: '/user/login',
  userReset: '/user/reset',
  userInfo: '/user/info',
  userSessionStatus: '/user/session-status',

  // ── Rclone ──
  rcloneProfiles: '/rclone/profiles',
  rcloneProfile: (id: number) => `/rclone/profiles/${id}` as const,
  rcloneProfileApply: (id: number) => `/rclone/profiles/${id}/apply` as const,
  rcloneProfileMount: (id: number) => `/rclone/profiles/${id}/mount` as const,
  rcloneProfileUnmount: (id: number) => `/rclone/profiles/${id}/unmount` as const,
} as const

export default endpoints
