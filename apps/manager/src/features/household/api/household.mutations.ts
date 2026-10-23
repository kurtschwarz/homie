import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import { householdApi } from './household.api'
import { householdKeys } from './household.queries'

function useHouseholdInvalidation() {
  const queryClient = useQueryClient()
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: householdKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useSavePerson() {
  return useMutation({
    mutationFn: householdApi.savePerson,
    onSuccess: useHouseholdInvalidation(),
  })
}

export function useDeletePerson() {
  return useMutation({
    mutationFn: householdApi.deletePerson,
    onSuccess: useHouseholdInvalidation(),
  })
}

export function useSaveTagMapping() {
  return useMutation({
    mutationFn: householdApi.saveTagMapping,
    onSuccess: useHouseholdInvalidation(),
  })
}

export function useDeleteTagMapping() {
  return useMutation({
    mutationFn: householdApi.deleteTagMapping,
    onSuccess: useHouseholdInvalidation(),
  })
}

export function useSyncCalendar() {
  return useMutation({
    mutationFn: householdApi.syncCalendar,
    onSuccess: useHouseholdInvalidation(),
  })
}
