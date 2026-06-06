import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

// 用户登录（后端：POST /user/login, body: {username, password}）
// 登录成功后服务端会设置 session cookie，前端无需额外处理
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
