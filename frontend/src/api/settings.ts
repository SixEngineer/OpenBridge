import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import endpoints from './endpoints'

export interface SettingsInfo {
  openlist_base_url: string
  app_version: string
  aria2_rpc_url: string
  aria2_path: string
  aria2_auto_start: boolean
  rclone_path: string
  session_device_limit: number
  auto_open_browser: boolean
  filetree_cache_size_kb: number
  filetree_cache_depth: number
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
  path: string
  auto_start: boolean
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsAria2, data)
}

export function updateRcloneSettings(data: {
  path: string
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsRclone, data)
}

export function updateSessionSettings(data: {
  device_limit: number
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsSession, data)
}

export function updateAppSettings(data: {
  auto_open_browser: boolean
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsApp, data)
}

export function updateFileTreeSettings(data: {
  cache_size_kb: number
  cache_depth: number
}): Promise<ApiResponse<SettingsInfo>> {
  return request.put(endpoints.settingsFileTree, data)
}
