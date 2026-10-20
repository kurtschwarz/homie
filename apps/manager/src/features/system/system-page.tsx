import * as React from "react";
import { Button } from "@homie/ui/components/button";
import { useAppData } from "../../app-data";
import { Page } from "../../shell/page";

/**
 * Deliberately independent of settings: a bad token makes every settings
 * query fail, and this is the page that fixes it.
 */
export function SystemPage() {
  const { health, token, saveToken } = useAppData();
  const [value, setValue] = React.useState(token);

  return <Page title="System" description="API connection and manager access.">
    <section className="flex flex-col gap-4">
      <h2 className="m-0 text-hm-title-sm">API connection</h2>
      <p className="m-0 text-hm-body text-hm-ink-2">
        {health.kind === "connecting" && "Connecting to API…"}
        {health.kind === "error" && `API unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" && <><span className="text-hm-accent">{health.health.status}</span>{` · v${health.health.version} · up ${health.health.uptimeSeconds}s`}</>}
      </p>
    </section>

    <form className="flex max-w-3xl flex-col items-start gap-4" onSubmit={(event) => { event.preventDefault(); saveToken(value); }}>
      <div className="flex w-full flex-col gap-2">
        <label htmlFor="api-token" className="text-hm-label text-hm-ink-2">API token</label>
        <input id="api-token" type="password" autoComplete="current-password" value={value}
          onChange={(event) => setValue(event.target.value)} placeholder="Paste token from API log"
          className="h-11 w-full rounded-hm-control border border-hm-line bg-hm-surface px-3 text-hm-body text-hm-ink focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-hm-accent" />
      </div>
      <Button type="submit" variant="solid">Save token</Button>
    </form>
  </Page>;
}
