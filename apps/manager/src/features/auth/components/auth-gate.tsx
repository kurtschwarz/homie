import type { ReactNode } from 'react'
import { ErrorMessage } from '@/shared/ui/error-message'
import { useSession } from '../use-session'
import { LoginPage } from './login-page'
import { SetupPage } from './setup-page'

/**
 * Sits above the routes, not inside the shell: setup and login have nothing
 * to do with the sidebar, and nothing under it should mount before there's a
 * user to show it to.
 *
 * Children are plain nodes rather than a render prop - anything below that
 * needs the user calls `useSession` and reads it from the cache.
 */
export function AuthGate({ children }: { children: ReactNode }) {
  const { isPending, error, setupRequired, user } = useSession()

  if (isPending) return <p className="p-6 text-hm-body text-hm-ink-2">Connecting to API…</p>
  if (error)
    return (
      <div className="p-6">
        <ErrorMessage label="Connection failed" message={error.message} />
      </div>
    )
  if (setupRequired) return <SetupPage />
  if (!user) return <LoginPage />
  return <>{children}</>
}
