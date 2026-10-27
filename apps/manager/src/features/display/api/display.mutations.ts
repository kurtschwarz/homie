import { useMutation, useQueryClient } from '@tanstack/react-query'
import { displayApi } from './display.api'
import { displayKeys } from './display.queries'

export function useUpdateDisplaySchedule() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: displayApi.updateSchedule,
    onSuccess: (schedule) => queryClient.setQueryData(displayKeys.schedule(), schedule),
  })
}

export function usePreviewDisplayMode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: displayApi.preview,
    onSuccess: (status) => queryClient.setQueryData(displayKeys.status(), status),
  })
}
