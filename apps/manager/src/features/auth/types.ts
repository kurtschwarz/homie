export type User = { id: string; username: string; role: 'ADMIN' }

/**
 * Mirrors the API's health status enum, which the system feature also types.
 * Auth keeps its own copy rather than importing that one: auth is a dependency
 * of the shell that renders before anything else, and reaching into the system
 * feature for two string literals would put the two features in a cycle.
 */
type SetupStatus = 'OK' | 'SETUP_REQUIRED'

/**
 * Health status rides along with the session because the gate needs both
 * answers in one round trip: whether first-run setup is still required, and
 * who is signed in.
 */
export type SessionResponse = {
  session: { user: User } | null
  health: { status: SetupStatus }
}

export type Credentials = { username: string; password: string }
export type SetupInput = Credentials & { code: string }
