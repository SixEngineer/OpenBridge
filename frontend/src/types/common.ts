export interface NavItem {
  label: string
  path: string
  description: string
}

export type HealthState = 'active' | 'disabled' | 'expired' | 'error'
