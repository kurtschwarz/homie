import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import type { Calendar } from '../types'
import { calendarsApi } from './calendars.api'
import { calendarKeys } from './calendars.queries'

function useInvalidateCalendars() {
  const queryClient = useQueryClient()
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: calendarKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useCreateCalendarAccount() {
  return useMutation({ mutationFn: calendarsApi.create })
}

export function useUpdateCalendarAccount() {
  return useMutation({ mutationFn: calendarsApi.update, onSuccess: useInvalidateCalendars() })
}

export function useDeleteCalendarAccount() {
  return useMutation({ mutationFn: calendarsApi.remove, onSuccess: useInvalidateCalendars() })
}

export function useTestDraftCalendarAccount() {
  return useMutation({ mutationFn: calendarsApi.testDraft })
}

export function useTestStoredCalendarAccount() {
  return useMutation({ mutationFn: calendarsApi.testStored })
}

export function useDiscoverCalendars() {
  return useMutation({ mutationFn: calendarsApi.discover, onSuccess: useInvalidateCalendars() })
}

export function useCalendarVisibility() {
  return useMutation({
    mutationFn: calendarsApi.setVisibility,
    onSuccess: useInvalidateCalendars(),
  })
}

export function useSyncCalendarAccount() {
  const invalidate = useInvalidateCalendars()
  return useMutation({
    mutationFn: ({ calendars }: { accountId: string; calendars: Calendar[] }) =>
      Promise.all(calendars.map((calendar) => calendarsApi.refresh(calendar.id))),
    onSuccess: invalidate,
  })
}
