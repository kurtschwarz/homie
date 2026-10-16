import { useHealth } from "./useHealth";

const theme = {
  bg: "var(--homie-bg)",
  ink: "#E2CFB6",
  ink2: "#8E7B66",
  accent: "#E0A060",
};

export function App() {
  const health = useHealth();

  return (
    <div
      style={{
        width: "1920px",
        height: "1080px",
        background: theme.bg,
        color: theme.ink,
        fontFamily: "'Lexend', system-ui, sans-serif",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 16,
      }}
    >
      <div style={{ fontSize: 64, fontWeight: 500 }}>Homie</div>
      <div style={{ fontSize: 28, color: theme.ink2 }}>
        {health.kind === "connecting" && "connecting to api…"}
        {health.kind === "error" && `api unreachable: ${health.message} — retrying…`}
        {health.kind === "ok" && (
          <span>
            <span style={{ color: theme.accent }}>{health.health.status}</span>
            {" · v"}
            {health.health.version}
            {" · up "}
            {health.health.uptimeSeconds}
            {"s"}
          </span>
        )}
      </div>
    </div>
  );
}
