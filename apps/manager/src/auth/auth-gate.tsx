import type { ReactNode } from 'react'
import { ErrorMessage } from '../components/error-message'
import { SetupPage } from '../features/setup/setup-page'
import { type User, useSession } from '../lib/use-session'
import { LoginPage } from './login-page'

/**
 * Sits above the routes, not inside the shell: setup and login have
 * nothing to do with the sidebar, and nothing under it should mount before
 * there's a user to show it to. Children are a render prop so the
 * resolved user and a logout callback can flow straight into
 * AppDataProvider without a second context.
 */
export function AuthGate({
  children,
}: {
  children: (user: User | null, logout: () => Promise<void>) => ReactNode
}) {
  const { state, setup, login, logout } = useSession()

  if (state.kind === 'loading')
    return <p className="p-6 text-hm-body text-hm-ink-2">Connecting to API…</p>
  if (state.kind === 'error')
    return (
      <div className="p-6">
        <ErrorMessage label="Connection failed" message={state.message} />
      </div>
    )
  if (state.setupRequired) return <SetupPage onSetup={setup} />
  if (!state.user) return <LoginPage onLogin={login} />
  return <>{children(state.user, logout)}</>
}
