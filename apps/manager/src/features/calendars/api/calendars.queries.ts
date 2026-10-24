import { queryOptions } from '@tanstack/react-query'
import { calendarsApi } from './calendars.api'

export const calendarKeys = {
  all: ['calendars'] as const,
  list: () => [...calendarKeys.all, 'list'] as const,
}

export const calendarQueries = {
  list: () =>
    queryOptions({
      queryKey: calendarKeys.list(),
      queryFn: calendarsApi.list,
    }),
}
