import * as React from "react";

import { type Person, varsFor } from "../lib/people";
import { cn } from "../lib/utils";
import { PersonBadges } from "./person-badge";

/** Seven day columns on kiosk and desktop, a single stacked list on mobile. Today first. */
export function WeekView({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="week-view"
      className={cn("grid min-h-0 grid-cols-7 gap-hm-16 mobile:grid-cols-1", className)}
      {...props}
    />
  );
}

/** One day. `today` gives it the accent-soft ground and an accent name. Empty days say "Nothing planned". */
export function WeekDay({
  name,
  date,
  today = false,
  children,
  className,
  ...props
}: React.ComponentProps<"section"> & { name: string; date: string; today?: boolean }) {
  const empty = React.Children.count(children) === 0;
  return (
    <section
      data-slot="week-day"
      data-today={today || undefined}
      className={cn(
        "flex min-w-0 flex-col gap-hm-12 rounded-hm-panel px-hm-16 py-hm-20",
        today && "bg-hm-accent-soft",
        className,
      )}
      {...props}
    >
      <h2 className="m-0 mx-1 mb-hm-8 flex flex-col gap-0.5 font-normal">
        <span
          className={cn("text-hm-body font-semibold", today ? "text-hm-accent" : "text-hm-ink")}
        >
          {name}
        </span>
        <span className="text-hm-caption text-hm-ink-3">{date}</span>
      </h2>
      {empty ? (
        <div className="px-1 py-hm-16 text-hm-caption text-hm-ink-3">Nothing planned</div>
      ) : (
        children
      )}
    </section>
  );
}

/** An event in a week column, tinted by the first person it's for (never by event type). `past` fades it. */
export function EventCard({
  time,
  title,
  people,
  whoLabel,
  past = false,
  className,
  ...props
}: React.ComponentProps<"div"> & {
  time: string;
  title: string;
  people: Person[];
  whoLabel?: string;
  past?: boolean;
}) {
  return (
    <div
      data-slot="event-card"
      className={cn(
        people[0] && varsFor(people[0]),
        "flex flex-col gap-hm-8 rounded-hm-16 bg-(--p-tint) px-hm-16 pt-hm-14 pb-hm-16",
        past && "opacity-(--hm-opacity-past)",
        className,
      )}
      {...props}
    >
      <div className="flex items-center justify-between gap-hm-8">
        <span className="text-hm-small text-hm-ink-2">{time}</span>
        <PersonBadges people={people} size="sm" />
      </div>
      <div className="line-clamp-3 text-hm-label">{title}</div>
      {whoLabel && <div className="text-hm-micro text-hm-ink-2">{whoLabel}</div>}
    </div>
  );
}
