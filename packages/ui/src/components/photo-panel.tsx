"use client"

import * as React from "react"
import { MapPin } from "lucide-react"

import { formatTakenAt } from "../lib/time"
import { cn } from "../lib/utils"

export interface Photo {
  src: string
  alt: string
  /** Reverse-geocoded from EXIF GPS. */
  place?: string
  /** EXIF date taken. */
  takenAt?: Date
}

type Shape = "portrait" | "square" | "landscape"
const KIOSK_WIDTH: Record<Shape, string> = {
  portrait: "kiosk:w-hm-panel-portrait",
  square: "kiosk:w-hm-panel-square",
  landscape: "kiosk:w-hm-panel-landscape",
}
const shapeOf = (w: number, h: number): Shape => (w / h < 0.9 ? "portrait" : w / h > 1.15 ? "landscape" : "square")

/**
 * The family photo slideshow. Crossfades every `intervalMs` (12 s); a tap advances and restarts the timer.
 * Kiosk: full height, width follows the photo's shape (720 / 900 / 1040px).
 * Desktop: fills the space beside the events column. Mobile: a full-width 4:3 card below the events.
 * Evening and Night dim the photos through --hm-photo-filter.
 */
export function PhotoPanel({
  photos,
  intervalMs = 12_000,
  className,
  ...props
}: React.ComponentProps<"section"> & { photos: Photo[]; intervalMs?: number }) {
  const [index, setIndex] = React.useState(0)
  const [shapes, setShapes] = React.useState<Record<number, Shape>>({})
  const [tick, setTick] = React.useState(0)

  React.useEffect(() => {
    if (photos.length < 2) return
    const id = setInterval(() => setIndex((i) => (i + 1) % photos.length), intervalMs)
    return () => clearInterval(id)
  }, [photos.length, intervalMs, tick])

  const current = photos[index]!
  const shape = shapes[index] ?? "landscape"

  return (
    <section
      data-slot="photo-panel"
      aria-label="Photo slideshow"
      className={cn(
        "relative min-w-0 flex-1 overflow-hidden rounded-hm-panel bg-hm-photo [filter:var(--hm-photo-filter)]",
        "transition-[width,filter] duration-900 ease-hm-panel",
        "kiosk:flex-none",
        KIOSK_WIDTH[shape],
        "mobile:order-1 mobile:aspect-[4/3] mobile:w-full mobile:flex-none",
        className,
      )}
      {...props}
    >
      {photos.map((p, i) => (
        <img
          key={p.src}
          src={p.src}
          alt={p.alt}
          onLoad={(e) => {
            const { naturalWidth: w, naturalHeight: h } = e.currentTarget
            if (w && h) setShapes((s) => (s[i] === shapeOf(w, h) ? s : { ...s, [i]: shapeOf(w, h) }))
          }}
          className={cn("absolute inset-0 size-full object-cover transition-opacity duration-900", i === index ? "opacity-100" : "opacity-0")}
        />
      ))}
      <button
        type="button"
        aria-label="Next photo"
        onClick={() => {
          setIndex((i) => (i + 1) % Math.max(1, photos.length))
          setTick((t) => t + 1)
        }}
        className="absolute inset-0 size-full cursor-pointer border-0 bg-transparent p-0"
      />
      {current && (current.place || current.takenAt) && (
        <div className="pointer-events-none absolute bottom-hm-28 left-hm-28 flex max-w-[calc(100%-56px*var(--hm-density))] items-start gap-hm-14 rounded-hm-18 bg-hm-scrim py-hm-16 pr-hm-24 pl-hm-18 text-hm-on-scrim">
          <MapPin aria-hidden className="mt-0.5 size-hm-28 shrink-0" strokeWidth={2} />
          <div className="flex min-w-0 flex-col gap-0.5">
            <div className="truncate text-hm-label-lg">{current.place ?? "Unknown place"}</div>
            {current.takenAt && <div className="text-hm-caption text-hm-on-scrim-2">{formatTakenAt(current.takenAt)}</div>}
          </div>
        </div>
      )}
    </section>
  )
}
