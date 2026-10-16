import { useHealth } from "./useHealth";

export function App() {
  const health = useHealth();

  return (
    <main className="flex h-hm-screen-h w-hm-screen-w flex-col items-center justify-center gap-hm-16 bg-hm-bg text-hm-ink">
      <h1 className="m-0 text-hm-display font-medium">Homie</h1>
      <div className="text-hm-body text-hm-ink-2">
        {health.kind === "connecting" && "connecting to api…"}
        {health.kind === "error" && `api unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" && (
          <span>
            <span className="text-hm-accent">{health.health.status}</span>
            {" · v"}
            {health.health.version}
            {" · up "}
            {health.health.uptimeSeconds}
            {"s"}
          </span>
        )}
      </div>
    </main>
  );
}
