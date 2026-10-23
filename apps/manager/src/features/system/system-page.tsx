import { Button } from '@homie/ui/components/button'
import { Page } from '@/app/shell/page'
import { useAppData } from '@/app-data'
import { useHealth } from '@/lib/use-health'
import { useSessions } from '@/lib/use-sessions'
import { ErrorMessage } from '@/shared/ui/error-message'

/** Deliberately independent of settings: this is the page that stays usable when the API is unreachable. */
export function SystemPage() {
  const { user, logout } = useAppData()
  const health = useHealth()
  const { sessions, isPending, error, revoke, revokeError } = useSessions()

  return (
    <Page title="System" description="API connection and manager access.">
      <section className="flex flex-col gap-4">
        <h2 className="m-0 text-hm-title-sm">API connection</h2>
        <p className="m-0 text-hm-body text-hm-ink-2">
          {health.isPending && 'Connecting to API…'}
          {health.isError && `API unreachable: ${health.error.message} — retrying…`}
          {health.isSuccess && (
            <>
              <span className="text-hm-accent">{health.data.status}</span>
              {` · v${health.data.version} · up ${health.data.uptimeSeconds}s`}
            </>
          )}
        </p>
      </section>

      {user && (
        <section className="flex max-w-3xl flex-col items-start gap-4">
          <h2 className="m-0 text-hm-title-sm">Account</h2>
          <p className="m-0 text-hm-body text-hm-ink-2">
            Signed in as <span className="text-hm-ink">{user.username}</span> ({user.role}).
          </p>
          <Button variant="outline" onClick={() => void logout()}>
            Log out
          </Button>
        </section>
      )}

      {user && (
        <section className="flex max-w-3xl flex-col gap-4">
          <h2 className="m-0 text-hm-title-sm">Active sessions</h2>
          {isPending && <p className="m-0 text-hm-body text-hm-ink-2">Loading sessions…</p>}
          {error && <ErrorMessage label="Sessions unavailable" message={error.message} />}
          {sessions && (
            <>
              {revokeError && (
                <ErrorMessage label="Couldn't revoke session" message={revokeError.message} />
              )}
              <ul className="m-0 flex list-none flex-col gap-2 p-0">
                {sessions.map((session) => (
                  <li
                    key={session.id}
                    className="flex items-center justify-between gap-4 rounded-hm-16 border border-hm-line px-4 py-3 text-hm-body text-hm-ink-2"
                  >
                    <span className="flex flex-col">
                      <span className="text-hm-ink">
                        {session.userAgent ?? 'Unknown device'}
                        {session.current && ' · this browser'}
                      </span>
                      <span className="text-hm-caption text-hm-ink-3">
                        Last used {new Date(session.lastSeenAt * 1000).toLocaleString()}
                      </span>
                    </span>
                    <Button variant="outline" onClick={() => void revoke(session.id)}>
                      Revoke
                    </Button>
                  </li>
                ))}
              </ul>
            </>
          )}
        </section>
      )}
    </Page>
  )
}
