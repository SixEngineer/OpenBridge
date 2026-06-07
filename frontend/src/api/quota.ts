import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { QuotaInfo } from '@/types/quota'
import endpoints from './endpoints'

// 查询配额
export function queryQuota(provider: string): Promise<ApiResponse<QuotaInfo>> {
  return request.post(endpoints.quotaQuery, { name: provider })
}

// 同步配额
export function syncQuota(provider: string): Promise<ApiResponse<QuotaInfo>> {
  return request.post(endpoints.quotaSync, { name: provider })
}