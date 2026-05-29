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
