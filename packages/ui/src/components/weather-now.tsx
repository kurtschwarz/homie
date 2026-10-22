import type * as React from "react";

import { cn } from "../lib/utils";
import { Eyebrow } from "./eyebrow";
import { CONDITION_LABEL, type WeatherCondition, WeatherIcon } from "./weather-icon";

/**
 * The Weather screen's Now card: big icon and temperature, condition, a sub-line, three metrics
 * (wind, rain next 3 h, sunset), and an optional Callout as children. Kiosk fixes it at 640px; on desktop and
 * mobile it fills its column (see WeatherLayout).
 */
export function WeatherNow({
  condition,
  temperature,
  sub,
  metrics,
  children,
  className,
  ...props
}: React.ComponentProps<"section"> & {
  condition: WeatherCondition;
  temperature: number;
  sub?: string;
  metrics: { label: string; value: string }[];
}) {
  return (
    <section
      data-slot="weather-now"
      aria-label="Right now"
      className={cn(
        "flex flex-col gap-hm-28 rounded-hm-32 bg-hm-surface p-hm-40 kiosk:w-hm-now-card",
        className,
      )}
      {...props}
    >
      <div className="flex items-center gap-hm-28 text-hm-ink">
        <WeatherIcon condition={condition} strokeWidth={1.5} className="size-hm-128 shrink-0" />
        <span className="text-hm-display-xl">{temperature}°</span>
      </div>
      <div className="flex flex-col gap-hm-8">
        <div className="text-hm-title">{CONDITION_LABEL[condition]}</div>
        {sub && <div className="text-hm-strong font-normal text-hm-ink-2">{sub}</div>}
      </div>
      <div className="grid grid-cols-[repeat(3,auto)] justify-between gap-hm-16 border-t-2 border-hm-line pt-hm-24">
        {metrics.map((m) => (
          <div key={m.label} className="flex flex-col gap-hm-6 whitespace-nowrap">
            <Eyebrow as="span">{m.label}</Eyebrow>
            <span className="text-hm-item">{m.value}</span>
          </div>
        ))}
      </div>
      {children && <div className="mt-auto">{children}</div>}
    </section>
  );
}
