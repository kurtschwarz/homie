import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import { associationsApi } from './associations.api'
import { associationKeys } from './associations.queries'

function useInvalidateAssociations() {
  const queryClient = useQueryClient()
  // An association decides which events carry a person's badge.
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: associationKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useCreateCalendarAssociation() {
  return useMutation({ mutationFn: associationsApi.create, onSuccess: useInvalidateAssociations() })
}

export function useUpdateCalendarAssociation() {
  return useMutation({ mutationFn: associationsApi.update, onSuccess: useInvalidateAssociations() })
}

export function useDeleteCalendarAssociation() {
  return useMutation({ mutationFn: associationsApi.remove, onSuccess: useInvalidateAssociations() })
}
