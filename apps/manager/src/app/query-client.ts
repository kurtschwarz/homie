import { QueryClient } from '@tanstack/react-query'

/**
 * Shared TanStack Query client for the manager app.
 *
 * `retry: false` keeps a failed request from retrying in a burst — health
 * polls every few seconds anyway, and a failed save should surface its error
 * straight away rather than after three silent attempts.
 *
 * Unlike the display's client, refetch-on-focus and background-pause are left
 * at their defaults: this is an admin page someone tabs away from and comes
 * back to, so refreshing on return is what you want, and polling while it is
 * hidden is wasted work.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
})
