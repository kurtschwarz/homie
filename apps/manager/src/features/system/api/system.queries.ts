import { queryOptions } from '@tanstack/react-query'
import { systemApi } from './system.api'

export const systemKeys = {
  all: ['system'] as const,
  health: () => [...systemKeys.all, 'health'] as const,
  sessions: () => [...systemKeys.all, 'sessions'] as const,
}

export const systemQueries = {
  /**
   * Polled every 3 seconds so the sidebar dot and the System page notice an
   * outage without anyone reloading. `staleTime: 0` opts out of the global
   * default: a health answer is worth nothing once it is seconds old.
   *
   * Consumers show `isError` in preference to the cached data behind it - a
   * stale `OK` reads as a working API, which is worse than admitting the
   * connection is down.
   */
  health: () =>
    queryOptions({
      queryKey: systemKeys.health(),
      queryFn: systemApi.health,
      refetchInterval: 3000,
      staleTime: 0,
    }),

  sessions: () =>
    queryOptions({
      queryKey: systemKeys.sessions(),
      queryFn: systemApi.sessions,
    }),
}
