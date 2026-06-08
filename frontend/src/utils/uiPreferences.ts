export const UI_MOTION_ENABLED_KEY = 'openbridge_ui_motion_enabled'

export function readMotionEnabled(): boolean {
  try {
    const raw = localStorage.getItem(UI_MOTION_ENABLED_KEY)
    if (raw === null) return true
    return raw !== 'false'
  } catch {
    return true
  }
}

export function applyMotionPreference(enabled = readMotionEnabled()) {
  document.documentElement.dataset.motion = enabled ? 'on' : 'off'
}

export function writeMotionEnabled(enabled: boolean) {
  localStorage.setItem(UI_MOTION_ENABLED_KEY, String(enabled))
  applyMotionPreference(enabled)
}
