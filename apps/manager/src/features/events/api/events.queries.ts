import { queryOptions } from '@tanstack/react-query'
import { eventsApi } from './events.api'

export const eventKeys = {
  all: ['events'] as const,
  list: () => [...eventKeys.all, 'list'] as const,
}

export const eventQueries = {
  /**
   * The manager only reads the week to show someone what their badge is
   * actually doing, so a failed fetch costs nothing but the week strip.
   *
   * Polling pauses while the tab is hidden and catches up on focus, which is
   * the right trade for a page someone tabs away from. Saves that can change
   * attribution invalidate this key, so an edit shows up without waiting out
   * the interval.
   */
  list: () =>
    queryOptions({
      queryKey: eventKeys.list(),
      queryFn: eventsApi.list,
      refetchInterval: 5 * 60_000,
    }),
}
