export interface QuotaInfo {
  provider: string
  total: number
  used: number
  available: number
  updated_at: string
}

export interface QuotaSnapshot {
  id: number
  provider_account_id: number
  provider_type: string
  total_quota: number
  used_quota: number
  available_quota: number
  sync_status: 'success' | 'failed'
  error_message?: string
  created_at: string
}