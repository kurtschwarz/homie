import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import { tagsApi } from './tags.api'
import { tagKeys } from './tags.queries'

function useTagsInvalidation() {
  const queryClient = useQueryClient()
  // Tag mappings decide which badge an event shows with.
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: tagKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useSaveTagMapping() {
  return useMutation({
    mutationFn: tagsApi.save,
    onSuccess: useTagsInvalidation(),
  })
}

export function useDeleteTagMapping() {
  return useMutation({
    mutationFn: tagsApi.remove,
    onSuccess: useTagsInvalidation(),
  })
}
