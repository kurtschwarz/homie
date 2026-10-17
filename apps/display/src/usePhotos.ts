import { useEffect, useState } from "react";
import type { Photo } from "@homie/ui/components/photo-panel";

type PhotoState =
  | { kind: "loading" }
  | { kind: "ready"; photos: Photo[] }
  | { kind: "error" };

const QUERY = `
  query Photos {
    photos {
      id
      url
      width
      height
      takenAt
      place
      focus { x y width height }
    }
  }
`;

/** Fetches the current photo library and refreshes periodically. */
export function usePhotos(): PhotoState {
  const [state, setState] = useState<PhotoState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    async function fetchPhotos() {
      let retryDelay = 5 * 60_000;
      try {
        const response = await fetch("/graphql", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: QUERY }),
        });
        if (!response.ok) throw new Error(`Photos request failed: ${response.status}`);
        const body = (await response.json()) as {
          data?: { photos?: Array<{ url: string; takenAt: number; place?: string | null; focus?: Photo["focus"] | null }> };
          errors?: unknown[];
        };
        if (body.errors?.length || !body.data?.photos) throw new Error("Photos query failed");
        const photos: Photo[] = body.data.photos.map((photo) => ({
          src: photo.url,
          alt: photo.place ?? "",
          ...(photo.place ? { place: photo.place } : {}),
          takenAt: new Date(photo.takenAt * 1000),
          ...(photo.focus ? { focus: photo.focus } : {}),
        }));
        if (!cancelled) setState({ kind: "ready", photos });
      } catch {
        retryDelay = 30_000;
        if (!cancelled) setState({ kind: "error" });
      } finally {
        if (!cancelled) timeout = setTimeout(fetchPhotos, retryDelay);
      }
    }

    void fetchPhotos();
    return () => {
      cancelled = true;
      if (timeout) clearTimeout(timeout);
    };
  }, []);

  return state;
}
