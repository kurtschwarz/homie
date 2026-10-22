import { useEffect, useState } from "react";
import { graphql } from "../../lib/graphql";
import { formatSetupCode } from "./setup-code";

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
        const data = await graphql<SetupResponse>(QUERY);
        if (!data.health) throw new Error("Setup status query failed");

        if (!cancelled) {
          if (data.health.status === "SETUP_REQUIRED" && data.setupCode) {
            setState({ kind: "setup", code: formatSetupCode(data.setupCode) });
          } else if (data.health.status === "OK") {
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
