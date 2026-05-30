import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { MountPoint, MountQuotaResult } from '@/types/mount'

// 创建 Mount
export function createMount(data: Partial<MountPoint>): Promise<ApiResponse<MountPoint>> {
  return request.post('/mount', data)
}

// 查询配额（通过 mount ID）
export function queryMountQuota(mountId: number): Promise<ApiResponse<MountQuotaResult>> {
  return request.get(`/mount/${mountId}/quota`)
}

// 同步配额（通过 mount ID）
export function syncMountQuota(mountId: number): Promise<ApiResponse<MountQuotaResult>> {
  return request.post(`/mount/${mountId}/quota/sync`)
}
