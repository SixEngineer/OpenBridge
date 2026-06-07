import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { ProviderRecord } from '@/types/provider'
import endpoints from './endpoints'

// 获取 Provider 列表
export function getProviderList(): Promise<ApiResponse<ProviderRecord[]>> {
  return request.get(endpoints.providerList)
}

// 注册 Provider
export function registerProvider(data: any): Promise<ApiResponse<any>> {
  return request.post(endpoints.provider, data)
}

// 删除 Provider
export function deleteProvider(id: number): Promise<ApiResponse<null>> {
  return request.delete(endpoints.provider, { params: { id } })
}

// 更新 Provider
export function updateProvider(data: Partial<ProviderRecord> & { id: number }): Promise<ApiResponse<null>> {
  return request.put(endpoints.provider + '/', data)
}

// 获取单个 Provider 信息
export function getProviderInfo(id: string): Promise<ApiResponse<ProviderRecord>> {
  return request.get(endpoints.providerInfo, { params: { id } })
}