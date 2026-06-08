import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import endpoints from './endpoints'

// 用户信息（来自 OpenList /api/me）
export interface UserInfo {
  id: number
  username: string
  password: string
  base_path: string
  role: number
  disabled: boolean
  permission: number
  sso_id: string
  otp: boolean
}

export interface SessionStatus {
  authenticated: boolean
  backend_instance_id: string
  openlist_base_url: string
  fingerprint: string
  device_id: string
  device_limit: number
  active_device_count: number
  username?: string
  role?: number
  checked_at: number
  reason?: string
}

// 用户登录（后端：POST /user/login, body: {username, password}）
export function userLogin(data: {
  username: string
  password: string
}): Promise<ApiResponse<null>> {
  return request.post(endpoints.userLogin, data)
}

export type UserResetScope = 'current' | 'all'

// 重置用户数据（后端：DELETE /user/reset?scope=current|all）
export function userReset(scope: UserResetScope = 'all'): Promise<ApiResponse<null>> {
  return request.delete(endpoints.userReset, {
    params: { scope },
  })
}

// 获取当前用户信息（后端：GET /user/info → 转发 OpenList /api/me）
export function getUserInfo(): Promise<ApiResponse<UserInfo>> {
  return request.get(endpoints.userInfo)
}

export function getSessionStatus(): Promise<ApiResponse<SessionStatus>> {
  return request.get(endpoints.userSessionStatus)
}
