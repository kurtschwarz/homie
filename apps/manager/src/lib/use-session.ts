import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'
import { graphql } from '@/shared/lib/graphql'
import type { HealthStatus } from './use-health'

export type User = { id: string; username: string; role: 'ADMIN' }

type Session = {
  session: { user: User } | null
  health: { status: HealthStatus }
}

const SESSION_QUERY = `{ session { user { id username role } } health { status } }`

/**
 * Whether setup or login is needed, refreshed after every setup, login or
 * logout. Failures from setup/login reject to the caller: a wrong code or
 * password is a form error, not a reason to tear down the whole gate. Health
 * status is the source of truth for whether first-run setup is still required.
 *
 * This asks for `health { status }` alongside the session rather than sharing
 * the `health` query, because the gate needs both answers in one round trip
 * before it can decide what to render.
 */
export function useSession() {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['session'],
    queryFn: () => graphql<Session>(SESSION_QUERY),
  })

  // Awaited by each mutation below, so `mutateAsync` resolves only once the
  // refetched session is in the cache and the gate has what it needs to move on.
  const refresh = useCallback(
    () => queryClient.invalidateQueries({ queryKey: ['session'] }),
    [queryClient],
  )

  const setupMutation = useMutation({
    mutationFn: (input: { code: string; username: string; password: string }) =>
      graphql(
        `mutation($code: String!, $username: String!, $password: String!) { setup(code: $code, username: $username, password: $password) { id } }`,
        input,
      ),
    onSuccess: refresh,
  })

  const loginMutation = useMutation({
    mutationFn: (input: { username: string; password: string }) =>
      graphql(
        `mutation($username: String!, $password: String!) { login(username: $username, password: $password) { id } }`,
        input,
      ),
    onSuccess: refresh,
  })

  const logoutMutation = useMutation({
    mutationFn: () => graphql(`mutation { logout }`),
    onSuccess: async () => {
      // Everything else in the cache was read as the user who just left, so
      // drop it: the next person to log in must not find their people,
      // sessions or events sitting there. The session query itself is kept and
      // refetched, which is what flips the gate back to the login form.
      queryClient.removeQueries({ predicate: (cached) => cached.queryKey[0] !== 'session' })
      await refresh()
    },
  })

  const setup = useCallback(
    async (code: string, username: string, password: string) => {
      await setupMutation.mutateAsync({ code, username, password })
    },
    [setupMutation.mutateAsync],
  )

  const login = useCallback(
    async (username: string, password: string) => {
      await loginMutation.mutateAsync({ username, password })
    },
    [loginMutation.mutateAsync],
  )

  const logout = useCallback(async () => {
    await logoutMutation.mutateAsync()
  }, [logoutMutation.mutateAsync])

  return {
    isPending: query.isPending,
    error: query.error,
    setupRequired: query.data?.health.status === 'SETUP_REQUIRED',
    user: query.data?.session?.user ?? null,
    setup,
    login,
    logout,
  }
}
