import type { HealthState } from './common'

export interface MetricCardData {
  title: string
  value: string
  detail: string
  trend: string
}

export interface SystemStatus {
  name: string
  state: HealthState
  detail: string
}

export interface TaskDigest {
  id: string
  name: string
  provider: string
  progress: number
  status: 'running' | 'queued' | 'error' | 'complete'
}

export interface AlertItem {
  title: string
  detail: string
  level: 'info' | 'warning' | 'critical'
}
