import { queryOptions } from '@tanstack/react-query'
import { peopleApi } from './people.api'

export const peopleKeys = {
  all: ['people'] as const,
  list: () => [...peopleKeys.all, 'list'] as const,
}

export const peopleQueries = {
  /**
   * Everyone who lives here. The sidebar count and the calendars tag editor
   * read the same entry, so person saves invalidate this key rather than each
   * reader fetching its own copy.
   */
  list: () =>
    queryOptions({
      queryKey: peopleKeys.list(),
      queryFn: peopleApi.list,
    }),
}
