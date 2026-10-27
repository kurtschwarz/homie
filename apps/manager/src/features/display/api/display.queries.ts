import { queryOptions } from '@tanstack/react-query'
import { displayApi } from './display.api'

export const displayKeys = {
  all: ['display'] as const,
  schedule: () => [...displayKeys.all, 'schedule'] as const,
  status: () => [...displayKeys.all, 'status'] as const,
  wifi: () => [...displayKeys.all, 'wifi'] as const,
}

export const displayQueries = {
  schedule: () => queryOptions({ queryKey: displayKeys.schedule(), queryFn: displayApi.schedule }),
  status: () => queryOptions({ queryKey: displayKeys.status(), queryFn: displayApi.status }),
  wifi: () => queryOptions({ queryKey: displayKeys.wifi(), queryFn: displayApi.wifi }),
}
