import { useCallback, useEffect, useRef, useState } from "react";

export type Photo = {
  id: string;
  url: string;
  width: number;
  height: number;
  takenAt: number;
  place?: string | null;
  focus?: { x: number; y: number; width: number; height: number } | null;
};

export type PhotoShowing = {
  id: string;
  photo: Photo;
  startedAt: number;
  endsAt: number;
  endedAt?: number | null;
  source: "SCHEDULE" | "TAP";
};

export type PhotoShowingState = { current: PhotoShowing | null; next: PhotoShowing | null };

const ROTATION_QUERY = `query PhotoShowing { photoShowing { current { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } next { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } } }`;
const ROTATION_MUTATION = (source: "SCHEDULE" | "TAP") => `mutation TrackPhotoShowing { trackPhotoShowing(source: ${source}) { current { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } next { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } } }`;

type Response<T> = { data?: T; errors?: unknown[] };

async function request<T>(query: string): Promise<T> {
  const response = await fetch("/graphql", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query }),
  });
  if (!response.ok) throw new Error(`Photo rotation request failed: ${response.status}`);
  const body = (await response.json()) as Response<T>;
  if (body.errors?.length || !body.data) throw new Error("Photo rotation request failed");
  return body.data;
}

/** Polls the API-owned slideshow timeline and advances expired slots. */
export function usePhotoShowing() {
  const [showing, setShowing] = useState<PhotoShowingState | null>(null);
  const advancing = useRef(false);
  const refreshInFlight = useRef(false);
  const advance = useCallback(async (source: "SCHEDULE" | "TAP") => {
    if (advancing.current) return;
    advancing.current = true;
    try {
      const { trackPhotoShowing: result } = await request<{ trackPhotoShowing: PhotoShowingState }>(ROTATION_MUTATION(source));
      setShowing(result);
    } catch {
      // The next poll retries if this slot expired; taps remain responsive to API recovery.
    } finally {
      advancing.current = false;
    }
  }, []);

  const refresh = useCallback(async () => {
    if (refreshInFlight.current) return;
    refreshInFlight.current = true;
    try {
      const { photoShowing: result } = await request<{ photoShowing: PhotoShowingState }>(ROTATION_QUERY);
      setShowing(result);
      if (result.current && result.next && result.current.endsAt <= Math.floor(Date.now() / 1000)) {
        await advance("SCHEDULE");
      }
    } catch {
      // Retain the last known frame while the API is temporarily unavailable.
    } finally {
      refreshInFlight.current = false;
    }
  }, [advance]);

  useEffect(() => {
    void refresh();
    const interval = window.setInterval(refresh, 1000);
    return () => window.clearInterval(interval);
  }, [refresh]);

  return { showing, advance };
}
