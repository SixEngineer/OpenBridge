import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import endpoints from './endpoints'

export interface PickLocalPathInput {
  kind: 'file' | 'directory'
  title?: string
  current_path?: string
  filter?: string
}

export interface PickLocalPathResult {
  path: string
}

export interface SystemMetrics {
  cpu_usage: number
  memory_usage: number
  memory_used_bytes: number
  memory_total_bytes: number
  disk_usage: number
  disk_used_bytes: number
  disk_total_bytes: number
  disk_path: string
  hostname: string
  sampled_at: string
}

export function pickLocalPath(data: PickLocalPathInput): Promise<ApiResponse<PickLocalPathResult>> {
  return request.post(endpoints.systemPickPath, data)
}

export function getSystemMetrics(): Promise<ApiResponse<SystemMetrics>> {
  return request.get(endpoints.systemMetrics)
}
