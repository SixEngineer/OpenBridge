export type RcloneMode = 'ordinary' | 'union' | 'combine'

export interface RcloneProfile {
  id: number
  name: string
  mode: RcloneMode
  mount_ids: number[]
  username: string
  target_path: string
  password_saved: boolean
  is_mounted: boolean
  mount_pid: number
  mount_rc_addr: string
  last_applied_at?: string | null
  last_mounted_at?: string | null
  last_error?: string
  created_at: string
  updated_at: string
  apply_commands: string[]
  mount_command: string
}

export interface RcloneProfileInput {
  name: string
  mode: RcloneMode
  mount_ids: number[]
  username: string
  password?: string
  target_path: string
}
