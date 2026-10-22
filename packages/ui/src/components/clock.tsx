"use client";

import type * as React from "react";

import { useNow } from "../hooks/use-now";
import { clockParts } from "../lib/time";
import { cn } from "../lib/utils";

/**
 * The time with a colon that blinks every 2 s (stopped under prefers-reduced-motion).
 * `default` is the home clock (`clock` role: 88 / 48 / 40px); `sm` is the screen-header clock (`clock-sm`).
 * Pass `date` to control it, or leave it out to follow the current time.
 */
export function Clock({
  date,
  size = "default",
  className,
  ...props
}: React.ComponentProps<"div"> & { date?: Date; size?: "default" | "sm" }) {
  const live = useNow();
  const { hours, minutes, ampm } = clockParts(date ?? live);
  return (
    <div
      data-slot="clock"
      className={cn(
        "flex items-baseline leading-none",
        size === "sm" ? "gap-hm-10" : "gap-hm-12",
        className,
      )}
      {...props}
    >
      <span className={size === "sm" ? "text-hm-clock-sm" : "text-hm-clock"}>
        {hours}
        <span aria-hidden className="inline-block motion-safe:animate-hm-blink">
          :
        </span>
        {minutes}
      </span>
      <span
        className={cn(
          "leading-none font-normal text-hm-ink-3",
          size === "sm" ? "text-hm-label" : "text-hm-item",
        )}
      >
        {ampm}
      </span>
    </div>
  );
}
