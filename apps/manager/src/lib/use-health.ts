import { useEffect, useState } from "react";
import { graphql } from "./graphql";

type Health = {
  status: string;
  version: string;
  uptimeSeconds: number;
};

export type HealthState =
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
        const data = await graphql<{ health: Health }>(QUERY);
        if (!cancelled) setState({ kind: "ok", health: data.health });
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
