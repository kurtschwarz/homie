import { useHealth } from "./useHealth";

export function App() {
  const health = useHealth();

  return (
    <div style={{ fontFamily: "system-ui, sans-serif", padding: 32 }}>
      <h1>Homie Manager</h1>
      <p>
        {health.kind === "connecting" && "connecting to api…"}
        {health.kind === "error" && `api unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" &&
          `${health.health.status} · v${health.health.version} · up ${health.health.uptimeSeconds}s`}
      </p>
    </div>
  );
}
