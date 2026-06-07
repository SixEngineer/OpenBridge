import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

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

// 用户登录（后端：POST /user/login, body: {username, password}）
export function userLogin(data: {
  username: string
  password: string
}): Promise<ApiResponse<null>> {
  return request.post('/user/login', data)
}

// 清空所有用户数据（后端：DELETE /user/reset）
export function userReset(): Promise<ApiResponse<null>> {
  return request.delete('/user/reset')
}

// 获取当前用户信息（后端：GET /user/info → 转发 OpenList /api/me）
export function getUserInfo(): Promise<ApiResponse<UserInfo>> {
  return request.get('/user/info')
}
