import type * as React from 'react'

import { cn } from '../lib/utils'

/** A full screen (Week, Weather): `screen-padding` around, `section-gap` between header and content. */
export function Screen({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="screen"
      className={cn(
        'flex min-h-dvh flex-col gap-hm-section p-hm-screen kiosk:h-dvh kiosk:w-screen kiosk:overflow-hidden',
        className,
      )}
      {...props}
    />
  )
}

/**
 * The home screen. Kiosk and desktop: photo panel beside the events column, `column-gap` apart.
 * Mobile: one column, events first and the photo card after (PhotoPanel orders itself).
 * Kiosk scales from its 1920 × 1080 design size to fill the viewport without scrolling.
 */
export function HomeLayout({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="home-layout"
      className={cn(
        'flex min-h-dvh gap-hm-column p-hm-screen mobile:flex-col kiosk:min-h-0 kiosk:h-dvh kiosk:w-screen kiosk:overflow-hidden',
        className,
      )}
      {...props}
    />
  )
}

/** The events column: header at the top, NextUp in the middle, the list at the bottom. Position: relative for the sheet. */
export function HomeEvents({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="home-events"
      className={cn(
        'relative flex min-w-0 flex-col justify-between gap-hm-section',
        'w-[clamp(360px,42%,560px)] flex-none kiosk:w-[clamp(22.5rem,42%,35rem)] kiosk:flex-1 kiosk:pt-hm-24 kiosk:pr-hm-32 mobile:w-full',
        className,
      )}
      {...props}
    />
  )
}

/** The Weather screen body: the Now card beside the forecasts, stacked on mobile. */
export function WeatherLayout({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot="weather-layout"
      className={cn(
        'grid min-h-0 flex-1 items-start gap-hm-column',
        'grid-cols-[minmax(300px,420px)_minmax(0,1fr)] kiosk:grid-cols-[var(--hm-size-now-card)_minmax(0,1fr)] mobile:grid-cols-1',
        className,
      )}
      {...props}
    />
  )
}
