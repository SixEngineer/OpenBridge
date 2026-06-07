export interface NavItem {
  label: string
  path: string
  description: string
  i18nKey?: string
  i18nDescKey?: string
}

export type HealthState = 'active' | 'disabled' | 'expired' | 'error'
