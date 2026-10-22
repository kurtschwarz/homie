import { describe, expect, it } from "vitest";
import { splitEvents, toRows, type CalendarEvent } from "./calendar-event";

const NOW = new Date("2026-10-18T13:00:00");

function event(partial: Partial<CalendarEvent> & { start: Date; end: Date }): CalendarEvent {
  return { id: "id", title: "Event", allDay: false, people: [], ...partial };
}

describe("splitEvents", () => {
  it("makes a live timed event the headline with a Now label", () => {
    const live = event({ id: "live", title: "Live", start: new Date("2026-10-18T12:30:00"), end: new Date("2026-10-18T14:15:00") });
    const { headline } = splitEvents([live], NOW);
    expect(headline?.event.id).toBe("live");
    expect(headline?.label).toBe("Now · until 2:15 pm");
  });

  it("picks the next timed event when nothing is live", () => {
    const next = event({ id: "next", title: "Next", start: new Date("2026-10-18T15:17:00"), end: new Date("2026-10-18T16:00:00") });
    const { headline } = splitEvents([next], NOW);
    expect(headline?.event.id).toBe("next");
    expect(headline?.label).toBe("Next · in 2 hr 17 min");
  });

  it("labels a next event tomorrow as Next · Tomorrow", () => {
    const tomorrow = event({ id: "tomorrow", title: "Tomorrow", start: new Date("2026-10-19T09:00:00"), end: new Date("2026-10-19T10:00:00") });
    const { headline } = splitEvents([tomorrow], NOW);
    expect(headline?.label).toBe("Next · Tomorrow");
  });

  it("never makes an all-day event the headline", () => {
    const allDay = event({ id: "all-day", title: "All day", allDay: true, start: new Date("2026-10-18T00:00:00"), end: new Date("2026-10-19T00:00:00") });
    const { headline } = splitEvents([allDay], NOW);
    expect(headline).toBeUndefined();
  });

  it("still produces rows with time 'All day' on an all-day-only day", () => {
    const allDay = event({ id: "all-day", title: "Bins", allDay: true, start: new Date("2026-10-18T00:00:00"), end: new Date("2026-10-19T00:00:00") });
    const { upcoming } = splitEvents([allDay], NOW);
    const rows = toRows(upcoming, NOW);
    expect(rows).toEqual([{ id: "all-day", time: "All day", day: undefined, people: [], title: "Bins" }]);
  });

  it("excludes the headline event from upcoming", () => {
    const next = event({ id: "next", start: new Date("2026-10-18T15:17:00"), end: new Date("2026-10-18T16:00:00") });
    const other = event({ id: "other", start: new Date("2026-10-18T17:00:00"), end: new Date("2026-10-18T18:00:00") });
    const { headline, upcoming } = splitEvents([next, other], NOW);
    expect(headline?.event.id).toBe("next");
    expect(upcoming.map((e) => e.id)).toEqual(["other"]);
  });

  it("omits day for today's rows and includes it for later ones", () => {
    const today = event({ id: "today", start: new Date("2026-10-18T17:00:00"), end: new Date("2026-10-18T18:00:00") });
    const later = event({ id: "later", start: new Date("2026-10-19T09:00:00"), end: new Date("2026-10-19T10:00:00") });
    const rows = toRows([today, later], NOW);
    expect(rows[0]?.day).toBeUndefined();
    expect(rows[1]?.day).toBe("Tomorrow");
  });

  it("dates a row a week out so its weekday is not mistaken for the nearer one", () => {
    const thisWeek = event({ id: "near", start: new Date("2026-10-21T09:00:00"), end: new Date("2026-10-21T10:00:00") });
    const nextWeek = event({ id: "far", start: new Date("2026-10-28T09:00:00"), end: new Date("2026-10-28T10:00:00") });
    const rows = toRows([thisWeek, nextWeek], NOW);
    expect(rows[0]?.day).toBe("Wednesday");
    expect(rows[1]?.day).toBe("Wed, Oct 28");
  });

  it("excludes an event that ended one minute ago", () => {
    const ended = event({ id: "ended", start: new Date("2026-10-18T11:00:00"), end: new Date("2026-10-18T12:59:00") });
    const { headline } = splitEvents([ended], NOW);
    expect(headline).toBeUndefined();
  });

  it("produces no headline and no rows for empty input", () => {
    const { headline, upcoming } = splitEvents([], NOW);
    expect(headline).toBeUndefined();
    expect(upcoming).toEqual([]);
  });
});
