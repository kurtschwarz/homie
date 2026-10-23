import { queryOptions } from '@tanstack/react-query'
import { tagsApi } from './tags.api'

export const tagKeys = {
  all: ['tags'] as const,
  list: () => [...tagKeys.all, 'list'] as const,
}

export const tagQueries = {
  /**
   * Every tag mapping, shared by the calendars editor and each person's tags
   * tab. Tag and person saves both invalidate this key, since a mapping
   * embeds its person's badge.
   */
  list: () =>
    queryOptions({
      queryKey: tagKeys.list(),
      queryFn: tagsApi.list,
    }),
}
