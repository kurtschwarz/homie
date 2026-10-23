import { queryOptions } from '@tanstack/react-query'
import { calendarsApi } from './calendars.api'

export const calendarKeys = {
  all: ['calendars'] as const,
  tags: () => [...calendarKeys.all, 'tags'] as const,
}

export const calendarQueries = {
  /** Tags seen on synced calendars, mapped or not. */
  tags: () =>
    queryOptions({
      queryKey: calendarKeys.tags(),
      queryFn: calendarsApi.listTags,
    }),
}
