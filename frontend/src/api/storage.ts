import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import endpoints from './endpoints'

// 获取存储驱动列表（加时间戳防止浏览器缓存）
export function getDrivers(): Promise<ApiResponse<any[]>> {
  return request.get(endpoints.storageDrivers, { params: { _t: Date.now() } })
}

// 获取驱动详情
export function getDriverInfo(name: string): Promise<ApiResponse<any>> {
  return request.get(endpoints.storageDriverInfo, { params: { name } })
}

// 获取文件列表
export function getFiles(params: {
  path: string
  page?: number
  per_page?: number
}, options?: {
  timeout?: number
}): Promise<ApiResponse<any>> {
  return request.get(endpoints.storageFiles, {
    params: {
      ...params,
      _t: Date.now(),
    },
    timeout: options?.timeout,
  })
}

// 获取文件信息
export function getFileInfo(path: string): Promise<ApiResponse<any>> {
  return request.get(endpoints.storageFile, { params: { path } })
}
