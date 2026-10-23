import { queryOptions } from '@tanstack/react-query'
import { authApi } from './auth.api'

export const authKeys = {
  all: ['auth'] as const,
  session: () => [...authKeys.all, 'session'] as const,
}

export const authQueries = {
  session: () =>
    queryOptions({
      queryKey: authKeys.session(),
      queryFn: authApi.session,
    }),
}
