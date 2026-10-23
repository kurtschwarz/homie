import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { graphql } from '@/shared/lib/graphql'

export type ActiveSession = {
  id: string
  userAgent: string | null
  createdAt: number
  lastSeenAt: number
  current: boolean
}

const SESSIONS_QUERY = `{ sessions { id userAgent createdAt lastSeenAt current } }`

/** The signed-in devices list for the System page, with revoke. */
export function useSessions() {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['sessions'],
    queryFn: () => graphql<{ sessions: ActiveSession[] }>(SESSIONS_QUERY),
    select: (data) => data.sessions,
  })

  const revoke = useMutation({
    mutationFn: (id: string) => graphql(`mutation($id: ID!) { revokeSession(id: $id) }`, { id }),
    // Revoking the session you are signed in with logs you out, so the gate's
    // session query is refreshed too rather than leaving it to notice on the
    // next request that fails.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: ['sessions'] }),
        queryClient.invalidateQueries({ queryKey: ['session'] }),
      ]),
  })

  return {
    sessions: query.data,
    isPending: query.isPending,
    error: query.error,
    // `mutate` rather than `mutateAsync`: a failed revoke belongs on
    // `revokeError` next to the list, and the caller has nothing to await.
    revoke: revoke.mutate,
    revokeError: revoke.error,
  }
}
