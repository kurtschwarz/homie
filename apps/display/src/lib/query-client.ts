import type { Query, QueryKey } from '@tanstack/react-query'
import { QueryClient } from '@tanstack/react-query'

/**
 * Shared TanStack Query client for the display app.
 *
 * Defaults preserve the earlier hand-rolled polling behavior:
 * - `retry: false` — a failed request waits for the next poll instead of
 *   retrying in a burst.
 * - `refetchOnWindowFocus: false` — regaining window focus does not trigger a
 *   refetch.
 * - `refetchIntervalInBackground: true` — polling continues while the kiosk
 *   tab is hidden.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      refetchOnWindowFocus: false,
      refetchIntervalInBackground: true,
    },
  },
})

/**
 * Builds a `refetchInterval` function that polls at a steady rate and backs
 * off while the query is in an error state.
 *
 * @param ms - Polling interval in milliseconds while the query is healthy.
 * @param afterErrorMs - Polling interval in milliseconds once the query's last
 * fetch has failed. Defaults to 30 seconds.
 * @returns A function suitable for the `refetchInterval` query option.
 *
 * @example
 * ```ts
 * useQuery({ queryKey, queryFn, refetchInterval: pollEvery(5_000) })
 * ```
 */
export const pollEvery =
  (ms: number, afterErrorMs = 30_000) =>
  <TQueryFnData, TError, TData, TQueryKey extends QueryKey>(
    query: Query<TQueryFnData, TError, TData, TQueryKey>,
  ) =>
    query.state.status === 'error' ? afterErrorMs : ms
