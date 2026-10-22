import * as React from "react"

import { cn } from "../lib/utils"
import { WeatherIcon, type WeatherCondition } from "./weather-icon"

export interface HourForecast {
  label: string
  condition: WeatherCondition
  temperature: number
  /** Chance of rain, 0–100. Shown only at 20% or more. */
  pop?: number
  now?: boolean
}

/** Twelve hour tiles spread across the width; a sideways scroller on mobile. The first is { label: "Now", now: true }. */
export function HourlyForecast({ hours, className, ...props }: React.ComponentProps<"div"> & { hours: HourForecast[] }) {
  return (
    <div
      data-slot="hourly-forecast"
      className={cn("grid auto-cols-fr grid-flow-col gap-hm-8 mobile:auto-cols-[60px] mobile:overflow-x-auto mobile:[scrollbar-width:none]", className)}
      {...props}
    >
      {hours.map((h, i) => (
        <div key={i} className={cn("flex flex-col items-center gap-hm-12 rounded-hm-18 py-hm-18", h.now && "bg-hm-surface")}>
          <span className="text-hm-small whitespace-nowrap text-hm-ink-3">{h.label}</span>
          <WeatherIcon condition={h.condition} className="size-hm-36" />
          <span className="text-hm-strong">{h.temperature}°</span>
          <span className="min-h-[calc(var(--hm-fs-micro)*1.3)] text-hm-micro font-semibold text-hm-accent">{h.pop != null && h.pop >= 20 ? `${h.pop}%` : ""}</span>
        </div>
      ))}
    </div>
  )
}

export interface DayForecast {
  name: string
  condition: WeatherCondition
  low: number
  high: number
  pop?: number
}

/** One row per day with a low–high range bar on a scale shared by the whole week. */
export function DailyForecast({ days, className, ...props }: React.ComponentProps<"div"> & { days: DayForecast[] }) {
  const min = Math.min(...days.map((d) => d.low))
  const max = Math.max(...days.map((d) => d.high))
  const span = Math.max(1, max - min)
  return (
    <div data-slot="daily-forecast" className={cn("flex flex-col", className)} {...props}>
      {days.map((d) => (
        <div
          key={d.name}
          className={cn(
            "grid h-hm-76 items-center gap-hm-20 border-t-2 border-hm-line text-hm-body",
            "grid-cols-[84px_24px_44px_36px_minmax(0,1fr)_36px] kiosk:grid-cols-[9.375rem_2.5rem_5rem_4rem_minmax(0,1fr)_4rem] mobile:grid-cols-[60px_22px_40px_32px_minmax(0,1fr)_32px]",
          )}
        >
          <span className="font-medium">{d.name}</span>
          <WeatherIcon condition={d.condition} className="size-hm-36 text-hm-ink-2" />
          <span className="text-hm-caption font-semibold text-hm-accent">{d.pop != null && d.pop >= 20 ? `${d.pop}%` : ""}</span>
          <span className="text-right text-hm-ink-3">{d.low}°</span>
          <div className="relative h-hm-12 rounded-full bg-hm-line">
            <div
              className="absolute inset-y-0 rounded-full bg-hm-accent"
              style={{ left: `${Math.round(((d.low - min) / span) * 100)}%`, width: `${Math.max(4, Math.round(((d.high - d.low) / span) * 100))}%` }}
            />
          </div>
          <span className="font-medium">{d.high}°</span>
        </div>
      ))}
    </div>
  )
}
