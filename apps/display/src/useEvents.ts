import { useEffect, useState } from "react";
import type { Swatch } from "@homie/ui/lib/people";
import type { CalendarEvent } from "./events";

type EventRecord = {
  id: string;
  title: string;
  location: string | null;
  start: number;
  end: number;
  allDay: boolean;
  people: { id: string; name: string; emoji: string; swatch: string }[];
};

type EventsState =
  | { kind: "loading" }
  | { kind: "ready"; events: CalendarEvent[] }
  | { kind: "error" };

const QUERY = `
  query Events {
    events(days: 21) {
      id
      title
      location
      start
      end
      allDay
      people { id name emoji swatch }
    }
  }
`;

function toCalendarEvent(record: EventRecord): CalendarEvent {
  return {
    id: record.id,
    title: record.title,
    location: record.location ?? undefined,
    start: new Date(record.start * 1000),
    end: new Date(record.end * 1000),
    allDay: record.allDay,
    people: record.people.map((p) => ({ id: p.id, name: p.name, emoji: p.emoji, swatch: p.swatch as Swatch })),
  };
}

/**
 * Fetches the family schedule through GraphQL and refreshes every 5 minutes —
 * shorter than the server's 15-minute sync, because the derived view (what's
 * Now vs Next) decays even without new data. A transient failure keeps the
 * last good schedule: a stale schedule is useful, an empty one is alarming.
 */
export function useEvents(): EventsState {
  const [state, setState] = useState<EventsState>({ kind: "loading" });

  useEffect(() => {
    let cancelled = false;
    let timeout: ReturnType<typeof setTimeout> | undefined;

    async function fetchEvents() {
      let retryDelay = 5 * 60_000;
      try {
        const response = await fetch("/graphql", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ query: QUERY }),
        });
        if (!response.ok) throw new Error(`Events request failed: ${response.status}`);
        const body = (await response.json()) as {
          data?: { events?: EventRecord[] };
          errors?: unknown[];
        };
        if (body.errors?.length || !body.data?.events) throw new Error("Events query failed");
        const events = body.data.events.map(toCalendarEvent);
        if (!cancelled) setState({ kind: "ready", events });
      } catch {
        retryDelay = 30_000;
        if (!cancelled) {
          setState((prev) => (prev.kind === "ready" && prev.events.length > 0 ? prev : { kind: "error" }));
        }
      } finally {
        if (!cancelled) timeout = setTimeout(fetchEvents, retryDelay);
      }
    }

    void fetchEvents();
    return () => {
      cancelled = true;
      if (timeout) clearTimeout(timeout);
    };
  }, []);

  return state;
}
