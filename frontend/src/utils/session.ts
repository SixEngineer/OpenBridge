export interface LocalSession {
  username: string
  deviceId: string
  issuedAt: number
  lastActiveAt: number
  timeoutMinutes: number
  backendFingerprint: string
  backendInstanceId: string
  openListBaseURL: string
}

export const AUTH_KEY = 'openbridge_auth'
export const DEVICE_ID_KEY = 'openbridge_device_id'
export const SESSION_TIMEOUT_KEY = 'openbridge_session_timeout_minutes'
export const LOGOUT_REASON_KEY = 'openbridge_logout_reason'
export const DEFAULT_SESSION_TIMEOUT_MINUTES = 120
export const MIN_SESSION_TIMEOUT_MINUTES = 5
export const MAX_SESSION_TIMEOUT_MINUTES = 10080

export function normalizeSessionTimeoutMinutes(value: number): number {
  if (!Number.isFinite(value)) return DEFAULT_SESSION_TIMEOUT_MINUTES
  return Math.min(MAX_SESSION_TIMEOUT_MINUTES, Math.max(MIN_SESSION_TIMEOUT_MINUTES, Math.round(value)))
}

export function readLocalSession(): LocalSession | null {
  try {
    const raw = localStorage.getItem(AUTH_KEY)
    if (!raw) return null

    const parsed = JSON.parse(raw) as Partial<LocalSession>
    const deviceId = typeof parsed.deviceId === 'string' ? parsed.deviceId : ensureDeviceId()
    if (
      typeof parsed.username !== 'string' ||
      typeof parsed.issuedAt !== 'number' ||
      typeof parsed.lastActiveAt !== 'number' ||
      typeof parsed.timeoutMinutes !== 'number' ||
      typeof parsed.backendFingerprint !== 'string' ||
      typeof parsed.backendInstanceId !== 'string' ||
      typeof parsed.openListBaseURL !== 'string'
    ) {
      return null
    }

    return {
      username: parsed.username,
      deviceId,
      issuedAt: parsed.issuedAt,
      lastActiveAt: parsed.lastActiveAt,
      timeoutMinutes: normalizeSessionTimeoutMinutes(parsed.timeoutMinutes),
      backendFingerprint: parsed.backendFingerprint,
      backendInstanceId: parsed.backendInstanceId,
      openListBaseURL: parsed.openListBaseURL,
    }
  } catch {
    return null
  }
}

export function writeLocalSession(session: LocalSession) {
  localStorage.setItem(AUTH_KEY, JSON.stringify(session))
}

export function clearLocalSession() {
  localStorage.removeItem(AUTH_KEY)
}

export function ensureDeviceId(): string {
  const existing = localStorage.getItem(DEVICE_ID_KEY)
  if (existing) return existing

  const generated = createDeviceId()
  localStorage.setItem(DEVICE_ID_KEY, generated)
  return generated
}

export function readSessionTimeoutMinutes(): number {
  try {
    const raw = localStorage.getItem(SESSION_TIMEOUT_KEY)
    if (!raw) return DEFAULT_SESSION_TIMEOUT_MINUTES
    return normalizeSessionTimeoutMinutes(Number(raw))
  } catch {
    return DEFAULT_SESSION_TIMEOUT_MINUTES
  }
}

export function writeSessionTimeoutMinutes(minutes: number) {
  localStorage.setItem(SESSION_TIMEOUT_KEY, String(normalizeSessionTimeoutMinutes(minutes)))
}

export function isSessionExpired(session: LocalSession, now = Date.now()): boolean {
  return now - session.lastActiveAt > session.timeoutMinutes * 60 * 1000
}

export function writeLogoutReason(reason: string) {
  localStorage.setItem(LOGOUT_REASON_KEY, reason)
}

export function readLogoutReason(): string {
  return localStorage.getItem(LOGOUT_REASON_KEY) || ''
}

export function clearLogoutReason() {
  localStorage.removeItem(LOGOUT_REASON_KEY)
}

function createDeviceId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  return `ob-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}
