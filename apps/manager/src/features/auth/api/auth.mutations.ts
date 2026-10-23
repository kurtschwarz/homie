import { useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi } from './auth.api'
import { authKeys } from './auth.queries'

/**
 * Setup and login both resolve the session again before settling, so the gate
 * has the new answer by the time the form stops submitting. Their errors stay
 * on the mutation for the form to show: a wrong code or password belongs next
 * to the field, not in a torn-down gate.
 */
export function useSetup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: authApi.setup,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: authKeys.session() }),
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: authApi.login,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: authKeys.session() }),
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: authApi.logout,
    onSuccess: async () => {
      // Everything outside this feature was read as the user who just left, so
      // drop it: the next person to sign in must not find their people,
      // sessions or events in the cache. The session query is kept and
      // refetched, which is what flips the gate back to the login form.
      queryClient.removeQueries({
        predicate: (cached) => cached.queryKey[0] !== authKeys.all[0],
      })
      await queryClient.invalidateQueries({ queryKey: authKeys.session() })
    },
  })
}
