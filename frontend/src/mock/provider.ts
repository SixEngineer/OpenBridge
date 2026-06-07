import type { ProviderRecord } from '@/types/provider'

export const providerRecords: ProviderRecord[] = [
  { id: 1, name: 'Baidu Netdisk', provider_type: 'baidu', net_disk: 'baidu', account_id: '', status: 'active', total_quota: 1024000, used_quota: 901120, available_quota: 122880, created_at: '', updated_at: '' },
  { id: 2, name: 'Aliyun Drive', provider_type: 'mock', net_disk: 'mock', account_id: '', status: 'active', total_quota: 1048576, used_quota: 430080, available_quota: 618496, created_at: '', updated_at: '' },
  { id: 3, name: 'OneDrive', provider_type: 'mock', net_disk: 'mock', account_id: '', status: 'disabled', total_quota: 1024000, used_quota: 706560, available_quota: 317440, created_at: '', updated_at: '' },
]
