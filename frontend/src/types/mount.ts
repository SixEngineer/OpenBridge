import type { QuotaInfo } from './quota'

export interface MountPoint {
  id: number
  name: string
  provider_account_id: number
  provider_type: string
  mount_path: string
  provider_root_path: string
  quota_mode: 'real' | 'inherit' | 'virtual'
  inherit_from_id?: number
  virtual_total: number
  virtual_used: number
  read_only: boolean
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface MountQuotaResult {
  mount_id: number
  mode: string
  allowed_max: number
  quota: QuotaInfo
  inherit_chain?: number[]
  virtual_config?: Record<string, number>
}
