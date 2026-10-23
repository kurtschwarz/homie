import { queryOptions } from '@tanstack/react-query'
import { householdApi } from './household.api'

export const householdKeys = {
  all: ['household'] as const,
  settings: () => [...householdKeys.all, 'settings'] as const,
}

export const householdQueries = {
  settings: () =>
    queryOptions({
      queryKey: householdKeys.settings(),
      queryFn: householdApi.get,
    }),
}
