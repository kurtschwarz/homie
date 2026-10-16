import { useEffect, useState } from "react";

type Health = {
  status: string;
  version: string;
  uptimeSeconds: number;
};

type HealthState =
  | { kind: "connecting" }
  | { kind: "ok"; health: Health }
  | { kind: "error"; message: string };

const QUERY = `{ health { status version uptimeSeconds } }`;
const RETRY_MS = 3000;

export function useHealth(): HealthState {
  const [state, setState] = useState<HealthState>({ kind: "connecting" });

  useEffect(() => {
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;

    const poll = async () => {
      try {
        const res = await fetch("/graphql", {
          method: "POST",
          headers: { "content-type": "application/json" },
          body: JSON.stringify({ query: QUERY }),
        });
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const body = await res.json();
        if (body.errors) throw new Error(body.errors[0]?.message ?? "GraphQL error");
        if (!cancelled) setState({ kind: "ok", health: body.data.health });
      } catch (err) {
        if (!cancelled) {
          setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
        }
      } finally {
        if (!cancelled) timer = setTimeout(poll, RETRY_MS);
      }
    };

    poll();
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, []);

  return state;
}
