import type { HealthState } from './common'

export interface ProviderRecord {
  id: number
  name: string
  provider_type: string
  net_disk: string
  account_id: string
  status: HealthState
  access_token?: string
  refresh_token?: string
  token_expires_at?: string
  auth_cookie?: string
  total_quota: number
  used_quota: number
  available_quota: number
  last_quota_sync_at?: string
  last_error?: string
  created_at: string
  updated_at: string
}