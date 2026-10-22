"use client"

import { Button as BaseButton } from "@base-ui/react/button"

import { cn } from "../lib/utils"
import { CONDITION_LABEL, WeatherIcon, type WeatherCondition } from "./weather-icon"

/**
 * Tappable icon and temperature in the home header that opens the Weather screen.
 * Use "cloud-sun" by day and "moon" at night. Temperature is a whole number; the degree sign is added for you.
 */
export function WeatherGlance({
  condition,
  temperature,
  stale = false,
  className,
  ...props
}: BaseButton.Props & { condition: WeatherCondition; temperature: number; stale?: boolean }) {
  return (
    <BaseButton
      data-slot="weather-glance"
      aria-label={`${CONDITION_LABEL[condition]}, ${temperature} degrees. Show the forecast`}
      className={cn(
        "inline-flex min-h-hm-touch cursor-pointer items-center gap-hm-12 rounded-hm-control bg-transparent px-hm-14 py-hm-6 leading-none text-hm-ink",
        stale && "opacity-55",
        "transition-transform duration-120 active:scale-(--hm-press-scale) focus-visible:outline-3 focus-visible:outline-offset-3 focus-visible:outline-hm-accent",
        className,
      )}
      {...props}
    >
      <WeatherIcon condition={condition} className="size-[5.5rem] shrink-0 text-hm-ink-2" />
      <span className="text-hm-clock">{temperature}°</span>
    </BaseButton>
  )
}
