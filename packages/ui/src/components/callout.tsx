import type * as React from "react";

import { cn } from "../lib/utils";

/** A soft accent-soft note with one practical tip ("Light jacket for school pickup · 13° at 3:30 pm"). */
export function Callout({
  icon,
  lead,
  children,
  className,
  ...props
}: React.ComponentProps<"div"> & { icon?: React.ReactNode; lead: React.ReactNode }) {
  return (
    <div
      data-slot="callout"
      className={cn(
        "flex items-center gap-hm-18 rounded-hm-20 bg-hm-accent-soft px-hm-26 py-hm-22 text-hm-body text-hm-ink-2 [&>svg]:size-hm-36 [&>svg]:shrink-0 [&>svg]:text-hm-accent",
        className,
      )}
      {...props}
    >
      {icon}
      <span>
        <span className="font-medium text-hm-ink">{lead}</span>
        {children ? <> · {children}</> : null}
      </span>
    </div>
  );
}
