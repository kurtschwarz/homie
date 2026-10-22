import { useCallback, useEffect, useState } from "react";
import { graphql } from "./graphql";

export type ActiveSession = {
  id: string;
  userAgent: string | null;
  createdAt: number;
  lastSeenAt: number;
  current: boolean;
};

type SessionsState =
  | { kind: "loading" }
  | { kind: "error"; message: string }
  | { kind: "ready"; sessions: ActiveSession[]; error?: string };

const SESSIONS_QUERY = `{ sessions { id userAgent createdAt lastSeenAt current } }`;

/** The signed-in devices list for the System page, with revoke. */
export function useSessions() {
  const [state, setState] = useState<SessionsState>({ kind: "loading" });

  const load = useCallback(async () => {
    try {
      const data = await graphql<{ sessions: ActiveSession[] }>(SESSIONS_QUERY);
      setState({ kind: "ready", sessions: data.sessions });
    } catch (err) {
      setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const revoke = useCallback(
    async (id: string) => {
      try {
        await graphql(`mutation($id: ID!) { revokeSession(id: $id) }`, { id });
        await load();
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        setState((prev) =>
          prev.kind === "ready" ? { ...prev, error: message } : { kind: "error", message },
        );
      }
    },
    [load],
  );

  return { state, revoke };
}
