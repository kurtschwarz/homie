import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import { peopleApi } from './people.api'
import { peopleKeys } from './people.queries'

function usePeopleInvalidation() {
  const queryClient = useQueryClient()
  // A person's badge is on their events, so saving or removing someone
  // refreshes those too.
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: peopleKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useSavePerson() {
  return useMutation({
    mutationFn: peopleApi.save,
    onSuccess: usePeopleInvalidation(),
  })
}

export function useDeletePerson() {
  return useMutation({
    mutationFn: peopleApi.remove,
    onSuccess: usePeopleInvalidation(),
  })
}
