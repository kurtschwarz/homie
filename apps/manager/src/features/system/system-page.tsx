import { Button } from "@homie/ui/components/button";
import { useAppData } from "../../app-data";
import { Page } from "../../shell/page";
import { ErrorMessage } from "../../components/error-message";
import { useSessions } from "../../lib/use-sessions";

/** Deliberately independent of settings: this is the page that stays usable when the API is unreachable. */
export function SystemPage() {
  const { health, user, logout } = useAppData();
  const { state, revoke } = useSessions();

  return <Page title="System" description="API connection and manager access.">
    <section className="flex flex-col gap-4">
      <h2 className="m-0 text-hm-title-sm">API connection</h2>
      <p className="m-0 text-hm-body text-hm-ink-2">
        {health.kind === "connecting" && "Connecting to API…"}
        {health.kind === "error" && `API unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" && <><span className="text-hm-accent">{health.health.status}</span>{` · v${health.health.version} · up ${health.health.uptimeSeconds}s`}</>}
      </p>
    </section>

    {user && <section className="flex max-w-3xl flex-col items-start gap-4">
      <h2 className="m-0 text-hm-title-sm">Account</h2>
      <p className="m-0 text-hm-body text-hm-ink-2">Signed in as <span className="text-hm-ink">{user.username}</span> ({user.role}).</p>
      <Button variant="outline" onClick={() => void logout()}>Log out</Button>
    </section>}

    {user && <section className="flex max-w-3xl flex-col gap-4">
      <h2 className="m-0 text-hm-title-sm">Active sessions</h2>
      {state.kind === "loading" && <p className="m-0 text-hm-body text-hm-ink-2">Loading sessions…</p>}
      {state.kind === "error" && <ErrorMessage label="Sessions unavailable" message={state.message} />}
      {state.kind === "ready" && <>
        {state.error && <ErrorMessage label="Couldn't revoke session" message={state.error} />}
        <ul className="m-0 flex list-none flex-col gap-2 p-0">
          {state.sessions.map((session) => <li key={session.id}
            className="flex items-center justify-between gap-4 rounded-hm-16 border border-hm-line px-4 py-3 text-hm-body text-hm-ink-2">
            <span className="flex flex-col">
              <span className="text-hm-ink">{session.userAgent ?? "Unknown device"}{session.current && " · this browser"}</span>
              <span className="text-hm-caption text-hm-ink-3">Last used {new Date(session.lastSeenAt * 1000).toLocaleString()}</span>
            </span>
            <Button variant="outline" onClick={() => void revoke(session.id)}>Revoke</Button>
          </li>)}
        </ul>
      </>}
    </section>}
  </Page>;
}
