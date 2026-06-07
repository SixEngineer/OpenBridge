import request from '@/utils/request'
import endpoints from './endpoints'
import type { ApiResponse } from '@/types/api'
import type { RcloneProfile, RcloneProfileInput } from '@/types/rclone'

export function listRcloneProfiles(): Promise<ApiResponse<RcloneProfile[]>> {
  return request.get(endpoints.rcloneProfiles)
}

export function createRcloneProfile(data: RcloneProfileInput): Promise<ApiResponse<RcloneProfile>> {
  return request.post(endpoints.rcloneProfiles, data)
}

export function updateRcloneProfile(id: number, data: RcloneProfileInput): Promise<ApiResponse<RcloneProfile>> {
  return request.put(endpoints.rcloneProfile(id), data)
}

export function deleteRcloneProfile(id: number): Promise<ApiResponse<null>> {
  return request.delete(endpoints.rcloneProfile(id))
}

export function applyRcloneProfile(id: number): Promise<ApiResponse<RcloneProfile>> {
  return request.post(endpoints.rcloneProfileApply(id))
}

export function mountRcloneProfile(id: number): Promise<ApiResponse<RcloneProfile>> {
  return request.post(endpoints.rcloneProfileMount(id))
}
