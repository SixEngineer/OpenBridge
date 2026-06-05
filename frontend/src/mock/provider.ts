import type { ProviderRecord } from '@/types/provider'

export const providerRecords: ProviderRecord[] = [
  { id: 1, name: 'Baidu Netdisk', provider_type: 'baidu', net_disk: 'baidu', account_id: '', status: 'active', total_quota: 1000, used_quota: 880, available_quota: 120, created_at: '', updated_at: '' },
  { id: 2, name: 'Aliyun Drive', provider_type: 'mock', net_disk: 'mock', account_id: '', status: 'active', total_quota: 1024, used_quota: 420, available_quota: 604, created_at: '', updated_at: '' },
  { id: 3, name: 'OneDrive', provider_type: 'mock', net_disk: 'mock', account_id: '', status: 'disabled', total_quota: 1000, used_quota: 690, available_quota: 310, created_at: '', updated_at: '' },
]
