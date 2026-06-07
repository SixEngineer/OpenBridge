import type { AlertItem, MetricCardData, SystemStatus, TaskDigest } from '@/types/dashboard'

export const metricCards: MetricCardData[] = [
  { title: 'Active Providers', value: '04', detail: 'Adapters currently online', trend: '+1 this week' },
  { title: 'Running Tasks', value: '12', detail: 'Downloads in progress', trend: '3 finishing soon' },
  { title: 'Quota Alerts', value: '02', detail: 'Providers nearing limits', trend: 'Watch threshold 85%' },
  { title: 'Success Rate', value: '96.4%', detail: 'Recent task completion rate', trend: '+2.1% from yesterday' },
]

export const systemStatuses: SystemStatus[] = [
  { name: 'OpenList', state: 'active', detail: 'Connection tested and responding' },
  { name: 'aria2', state: 'active', detail: 'RPC worker available' },
  { name: 'Quota Sync', state: 'disabled', detail: 'Last refresh delayed by 18 minutes' },
]

export const taskDigests: TaskDigest[] = [
  { id: 'DL-1024', name: 'Semester archive bundle', provider: 'Baidu Netdisk', progress: 78, status: 'running' },
  { id: 'DL-1021', name: 'OpenList export backup', provider: 'Aliyun Drive', progress: 100, status: 'complete' },
  { id: 'DL-1018', name: 'Research dataset mirror', provider: 'OneDrive', progress: 42, status: 'queued' },
  { id: 'DL-1012', name: 'Presentation media pack', provider: 'Google Drive', progress: 16, status: 'error' },
]

export const alertItems: AlertItem[] = [
  { title: 'Baidu quota approaching threshold', detail: 'Used 88% of the mirrored storage budget.', level: 'warning' },
  { title: 'Resolver fallback triggered', detail: 'One task used fallback headers during direct-link resolution.', level: 'info' },
  { title: 'Task DL-1012 needs retry', detail: 'Remote source rejected token refresh during transfer.', level: 'critical' },
]
