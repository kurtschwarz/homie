import { queryOptions } from '@tanstack/react-query'
import { weatherApi } from './weather.api'

export const weatherKeys = {
  all: ['weather'] as const,
  list: () => [...weatherKeys.all, 'list'] as const,
  reading: () => [...weatherKeys.all, 'reading'] as const,
}

export const weatherQueries = {
  list: () =>
    queryOptions({
      queryKey: weatherKeys.list(),
      queryFn: weatherApi.list,
    }),
  reading: () =>
    queryOptions({
      queryKey: weatherKeys.reading(),
      queryFn: weatherApi.reading,
    }),
}
