import { graphql } from '@/shared/lib/graphql'
import type { ActiveSession, Health } from '../types'

const HEALTH_QUERY = `{ health { status version uptimeSeconds } }`
const SESSIONS_QUERY = `{ sessions { id userAgent createdAt lastSeenAt current } }`

export const systemApi = {
  health: async (): Promise<Health> => {
    const data = await graphql<{ health: Health }>(HEALTH_QUERY)
    return data.health
  },

  sessions: async (): Promise<ActiveSession[]> => {
    const data = await graphql<{ sessions: ActiveSession[] }>(SESSIONS_QUERY)
    return data.sessions
  },

  revokeSession: (id: string) => graphql(`mutation($id: ID!) { revokeSession(id: $id) }`, { id }),
}
