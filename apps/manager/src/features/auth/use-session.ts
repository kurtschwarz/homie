import { useQuery } from '@tanstack/react-query'
import { authQueries } from './api/auth.queries'

/**
 * Who is signed in, and whether first-run setup is still outstanding. Health
 * status is the source of truth for the latter.
 *
 * Several places read this - the gate, the System page - and the cache shares
 * one request between them, so there is nothing to hold in context.
 */
export function useSession() {
  const query = useQuery(authQueries.session())

  return {
    isPending: query.isPending,
    error: query.error,
    setupRequired: query.data?.health.status === 'SETUP_REQUIRED',
    user: query.data?.session?.user ?? null,
  }
}
