import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { MountPoint, MountQuotaResult } from '@/types/mount'
import endpoints from './endpoints'

// 创建 Mount
export function createMount(data: Partial<MountPoint>): Promise<ApiResponse<MountPoint>> {
  return request.post(endpoints.mount, data)
}

// 列表所有 Mount
export function listMounts(): Promise<ApiResponse<MountPoint[]>> {
  return request.get(endpoints.mount)
}

// 更新 Mount
export function updateMount(id: number, data: Partial<MountPoint>): Promise<ApiResponse<MountPoint>> {
  return request.put(endpoints.mountDetail(id), data)
}

// 删除 Mount
export function deleteMount(id: number): Promise<ApiResponse<null>> {
  return request.delete(endpoints.mountDetail(id))
}

// 查询配额（通过 mount ID）
export function queryMountQuota(mountId: number): Promise<ApiResponse<MountQuotaResult>> {
  return request.get(endpoints.mountQuota(mountId))
}

// 同步配额（通过 mount ID）
export function syncMountQuota(mountId: number): Promise<ApiResponse<MountQuotaResult>> {
  return request.post(endpoints.mountQuotaSync(mountId))
}
