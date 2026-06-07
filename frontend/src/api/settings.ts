import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import endpoints from './endpoints'

export interface SettingsInfo {
  openlist_base_url: string
  aria2_rpc_url: string
  rclone_path: string
}

export function getSettings(): Promise<ApiResponse<SettingsInfo>> {
  return request.get(endpoints.settings)
}

export function updateOpenListSettings(data: {
  base_url: string
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsOpenList, data)
}

export function updateAria2Settings(data: {
  rpc_url: string
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsAria2, data)
}

export function updateRcloneSettings(data: {
  path: string
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsRclone, data)
}
