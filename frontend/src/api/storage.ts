import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'

// 获取存储驱动列表
export function getDrivers(): Promise<ApiResponse<any[]>> {
  return request.get('/storage/drivers')
}

// 获取驱动详情
export function getDriverInfo(name: string): Promise<ApiResponse<any>> {
  return request.get('/storage/driverInfo', { params: { name } })
}

// 获取文件列表
export function getFiles(params: {
  path: string
  page?: number
  per_page?: number
}): Promise<ApiResponse<any>> {
  return request.get('/storage/files', { params })
}

// 获取文件信息
export function getFileInfo(path: string): Promise<ApiResponse<any>> {
  return request.get('/storage/file', { params: { path } })
}
