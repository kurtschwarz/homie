import type { Query, QueryKey } from '@tanstack/react-query'
import { QueryClient } from '@tanstack/react-query'

/**
 * Preserves today's hand-rolled polling behavior: a failed request waits for
 * the next poll rather than retrying in a burst, polling continues while the
 * kiosk tab is hidden, and nothing refetches just because the window
 * regained focus.
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

/** Polls at `ms`, backing off to `afterErrorMs` once a refetch has failed. */
export const pollEvery =
  (ms: number, afterErrorMs = 30_000) =>
  <TQueryFnData, TError, TData, TQueryKey extends QueryKey>(
    query: Query<TQueryFnData, TError, TData, TQueryKey>,
  ) =>
    query.state.status === 'error' ? afterErrorMs : ms
