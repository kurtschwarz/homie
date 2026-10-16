import { useHealth } from "./useHealth";

export function App() {
  const health = useHealth();

  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-hm-16 p-hm-screen text-hm-ink">
      <h1 className="m-0 text-hm-title-lg">Homie Manager</h1>
      <p className="m-0 text-hm-body text-hm-ink-2">
        {health.kind === "connecting" && "connecting to api…"}
        {health.kind === "error" && `api unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" &&
          <>
            <span className="text-hm-accent">{health.health.status}</span>
            {` · v${health.health.version} · up ${health.health.uptimeSeconds}s`}
          </>}
      </p>
    </main>
  );
}
