import { useQuery } from '@tanstack/react-query'
import { graphql } from '@/shared/lib/graphql'

export type HealthStatus = 'OK' | 'SETUP_REQUIRED'

export type Health = {
  status: HealthStatus
  version: string
  uptimeSeconds: number
}

const QUERY = `{ health { status version uptimeSeconds } }`

/**
 * The API's report on itself, polled every 3 seconds so the sidebar dot and
 * the System page notice an outage without anyone reloading.
 *
 * Consumers show `isError` in preference to the cached `data` behind it: a
 * stale `OK` reads as a working API, which is worse than admitting the
 * connection is down.
 */
export function useHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: () => graphql<{ health: Health }>(QUERY),
    select: (data) => data.health,
    refetchInterval: 3000,
  })
}
