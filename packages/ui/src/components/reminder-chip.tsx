import type * as React from "react";

import type { Person } from "../lib/people";
import { cn } from "../lib/utils";
import { PersonBadge } from "./person-badge";

/**
 * One-line nudge under the clock (bins out tonight, a birthday this week). Show one at a time and rotate every 10 s
 * when there are several; render nothing when there's nothing to say. Lead with a lucide icon (it turns accent)
 * or a person.
 */
export function ReminderChip({
  icon,
  person,
  lead,
  children,
  className,
  ...props
}: React.ComponentProps<"div"> & {
  icon?: React.ReactNode;
  person?: Person;
  lead: React.ReactNode;
}) {
  return (
    <div
      data-slot="reminder-chip"
      className={cn(
        "flex min-h-hm-40 min-w-0 items-center gap-hm-12 text-hm-body-sm text-hm-ink-2 [&>svg]:size-hm-30 [&>svg]:shrink-0 [&>svg]:text-hm-accent",
        className,
      )}
      {...props}
    >
      {person ? <PersonBadge person={person} size="chip" /> : icon}
      <span className="truncate">
        <span className="font-medium text-hm-ink">{lead}</span> {children}
      </span>
    </div>
  );
}
