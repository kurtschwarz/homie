import { useMutation, useQueryClient } from '@tanstack/react-query'
import { authKeys } from '@/features/auth'
import { systemApi } from './system.api'
import { systemKeys } from './system.queries'

export function useRevokeSession() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: systemApi.revokeSession,
    // Revoking the session you are signed in with logs you out, so the gate's
    // session query is refreshed too rather than leaving it to find out on the
    // next request that fails.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: systemKeys.sessions() }),
        queryClient.invalidateQueries({ queryKey: authKeys.session() }),
      ]),
  })
}
