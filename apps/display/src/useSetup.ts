import { useEffect, useState } from "react";
import { formatSetupCode } from "./setup";

type SetupState =
  | { kind: "loading" }
  | { kind: "setup"; code: string }
  | { kind: "ready" }
  | { kind: "error" };

type SetupResponse = {
  health: { status: "OK" | "SETUP_REQUIRED" };
  setupCode: string | null;
};

const QUERY = `{ health { status } setupCode }`;
const REFRESH_MS = 5000;
const RETRY_MS = 30_000;

/** Polls first-run status and retains a visible code through transient failures. */
export function useSetup(): SetupState {
  const [state, setState] = useState<SetupState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    async function poll() {
      let delay = REFRESH_MS;
      try {
        const response = await fetch("/graphql", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: QUERY }),
        });
        if (!response.ok) throw new Error(`Setup status request failed: ${response.status}`);
        const body = (await response.json()) as { data?: SetupResponse; errors?: unknown[] };
        if (body.errors?.length || !body.data?.health) throw new Error("Setup status query failed");

        if (!cancelled) {
          if (body.data.health.status === "SETUP_REQUIRED" && body.data.setupCode) {
            setState({ kind: "setup", code: formatSetupCode(body.data.setupCode) });
          } else if (body.data.health.status === "OK") {
            setState({ kind: "ready" });
          }
        }
      } catch {
        delay = RETRY_MS;
        if (!cancelled) setState((previous) => previous.kind === "setup" ? previous : { kind: "error" });
      } finally {
        if (!cancelled) timeout = setTimeout(poll, delay);
      }
    }

    void poll();
    return () => {
      cancelled = true;
      if (timeout) clearTimeout(timeout);
    };
  }, []);

  return state;
}
