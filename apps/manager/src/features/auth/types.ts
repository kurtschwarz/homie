import type { HealthStatus } from '@/lib/use-health'

export type User = { id: string; username: string; role: 'ADMIN' }

/**
 * Health status rides along with the session because the gate needs both
 * answers in one round trip: whether first-run setup is still required, and
 * who is signed in.
 */
export type SessionResponse = {
  session: { user: User } | null
  health: { status: HealthStatus }
}

export type Credentials = { username: string; password: string }
export type SetupInput = Credentials & { code: string }
