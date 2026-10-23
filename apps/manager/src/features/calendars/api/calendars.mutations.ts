import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import { calendarsApi } from './calendars.api'
import { calendarKeys } from './calendars.queries'

export function useSyncCalendar() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: calendarsApi.sync,
    // A sync refreshes the events and can reveal tags not seen before.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: calendarKeys.all }),
        queryClient.invalidateQueries({ queryKey: eventKeys.all }),
      ]),
  })
}
