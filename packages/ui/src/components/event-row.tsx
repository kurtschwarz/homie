import type * as React from "react";

import type { Person } from "../lib/people";
import { cn } from "../lib/utils";
import { PersonBadges } from "./person-badge";

/**
 * One upcoming event as a row: when, who, what. `time` is "5:45 pm" or "All day"; pass `day` ("Tomorrow",
 * "Saturday") only when it isn't today. `compact` is the Coming up sheet's variant (`item` titles, no ring).
 */
export function EventRow({
  time,
  day,
  people,
  title,
  compact = false,
  className,
  ...props
}: React.ComponentProps<"div"> & {
  time: string;
  day?: string;
  people: Person[];
  title: string;
  compact?: boolean;
}) {
  return (
    <div
      data-slot="event-row"
      className={cn(
        "grid grid-cols-[var(--hm-time-col)_auto_minmax(0,1fr)] items-center gap-hm-20 border-t-2 border-hm-line",
        compact ? "py-hm-14" : "py-hm-16",
        className,
      )}
      {...props}
    >
      <div className="flex flex-col gap-0.5 leading-[1.15]">
        {day && (
          <span className="text-hm-eyebrow leading-[1.15] uppercase text-hm-ink-3">{day}</span>
        )}
        <span className="text-hm-body leading-[1.15] text-hm-ink-2">{time}</span>
      </div>
      <PersonBadges people={people} ring={!compact} />
      <div className={cn("truncate", compact ? "text-hm-item" : "text-hm-item-lg")}>{title}</div>
    </div>
  );
}

export function EventList({ className, ...props }: React.ComponentProps<"div">) {
  return <div data-slot="event-list" className={cn("flex flex-col", className)} {...props} />;
}
