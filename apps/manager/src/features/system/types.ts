export type HealthStatus = 'OK' | 'SETUP_REQUIRED'

export type Health = {
  status: HealthStatus
  version: string
  uptimeSeconds: number
}

export type ActiveSession = {
  id: string
  userAgent: string | null
  createdAt: number
  lastSeenAt: number
  current: boolean
}
