import { dayLabel, formatRange, formatTime, until } from "@homie/ui/lib/time";
import type { Person } from "@homie/ui/lib/people";

export type CalendarEvent = {
  id: string;
  title: string;
  location?: string;
  start: Date;
  end: Date;
  allDay: boolean;
  people: Person[];
};

export type Headline = {
  event: CalendarEvent;
  label: string;
  detail: string;
};

/**
 * Splits events into the one headline (now or next) and everything else.
 * Only timed events are headline candidates: an all-day event as the 112px
 * hero would be wrong, per the prototype.
 */
export function splitEvents(events: CalendarEvent[], now: Date): { headline?: Headline; upcoming: CalendarEvent[] } {
  const candidates = events.filter((e) => !e.allDay && e.end > now);
  const live = candidates.find((e) => e.start <= now);
  const headlineEvent = live ?? candidates.find((e) => e.start > now);

  if (!headlineEvent) {
    return { upcoming: events };
  }

  const label = live
    ? `Now · until ${formatTime(headlineEvent.end)}`
    : `Next · ${isToday(headlineEvent.start, now) ? until(headlineEvent.start.getTime() - now.getTime()) : dayLabel(headlineEvent.start, now)}`;

  const detail = [formatRange(headlineEvent.start, headlineEvent.end), headlineEvent.location].filter(Boolean).join(" · ");

  return {
    headline: { event: headlineEvent, label, detail },
    upcoming: events.filter((e) => e.id !== headlineEvent.id),
  };
}

export type EventRowData = {
  id: string;
  time: string;
  day?: string;
  people: Person[];
  title: string;
};

/** Row labels for EventRow: day is omitted for today, per event-row.tsx's contract. */
export function toRows(events: CalendarEvent[], now: Date): EventRowData[] {
  return events.map((e) => ({
    id: e.id,
    time: e.allDay ? "All day" : formatTime(e.start),
    day: isToday(e.start, now) ? undefined : dayLabel(e.start, now),
    people: e.people,
    title: e.title,
  }));
}

function isToday(d: Date, now: Date) {
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth() && d.getDate() === now.getDate();
}
