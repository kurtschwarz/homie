'use client'

import { MapPin } from 'lucide-react'
import * as React from 'react'

import { formatTakenAt } from '../lib/time'
import { cn } from '../lib/utils'

export interface Photo {
  src: string
  alt: string
  /** Reverse-geocoded from EXIF GPS. */
  place?: string
  /** EXIF date taken. */
  takenAt?: Date
  focus?: { x: number; y: number; width: number; height: number }
}

type Shape = 'portrait' | 'square' | 'landscape'
const PANEL_WIDTH: Record<Shape, string> = {
  portrait: 'w-hm-panel-portrait',
  square: 'w-hm-panel-square',
  landscape: 'w-hm-panel-landscape',
}
const shapeOf = (aspectRatio: number): Shape =>
  aspectRatio < 0.9 ? 'portrait' : aspectRatio > 1.15 ? 'landscape' : 'square'

/**
 * The family photo slideshow. The display API controls the current frame and schedule; a tap requests an advance.
 * Full height, with the width following the photo's shape (720 / 900 / 1040px).
 * Kiosk-only: the home screen it belongs to never renders on another surface.
 * Evening and Night dim the photos through --hm-photo-filter.
 */
export function PhotoPanel({
  photos,
  currentSrc,
  onAdvance,
  action,
  className,
  ...props
}: React.ComponentProps<'section'> & {
  photos: Photo[]
  currentSrc?: string
  onAdvance?: () => void
  action?: React.ReactNode
}) {
  const [aspectRatios, setAspectRatios] = React.useState<Record<string, number>>({})
  const [panelHeight, setPanelHeight] = React.useState(0)
  const [frameWidths, setFrameWidths] = React.useState<Record<Shape, number>>()
  const panelRef = React.useRef<HTMLElement>(null)

  React.useEffect(() => {
    const panel = panelRef.current
    if (!panel) return
    // Read the resting widths from the tokens rather than repeating them here.
    const styles = getComputedStyle(panel)
    const token = (name: string) => Number.parseFloat(styles.getPropertyValue(name))
    const widths = {
      portrait: token('--hm-size-panel-portrait'),
      square: token('--hm-size-panel-square'),
      landscape: token('--hm-size-panel-landscape'),
    }
    if (Object.values(widths).every((width) => width > 0)) setFrameWidths(widths)

    // Height is the one panel dimension that holds still, so it is the only one
    // the images are allowed to depend on.
    const observer = new ResizeObserver(([entry]) => {
      if (!entry) return
      const { height } = entry.contentRect
      if (height) setPanelHeight(height)
    })
    observer.observe(panel)
    return () => observer.disconnect()
  }, [])

  const current = photos.find((photo) => photo.src === currentSrc) ?? photos[0]
  const activeSrc = currentSrc ?? current?.src
  // Landscape until the image loads and reports its real aspect, matching the
  // widest panel so the first frame doesn't visibly grow into place.
  const currentAspectRatio = current ? aspectRatios[current.src] : undefined
  const shape = currentAspectRatio ? shapeOf(currentAspectRatio) : 'landscape'

  return (
    <section
      ref={panelRef}
      data-slot="photo-panel"
      aria-label="Photo slideshow"
      className={cn(
        'relative min-w-0 flex-none overflow-hidden rounded-hm-panel bg-hm-photo [filter:var(--hm-photo-filter)]',
        'transition-[width,filter] duration-900 ease-hm-panel',
        PANEL_WIDTH[shape],
        className,
      )}
      {...props}
    >
      {photos.map((p, i) => {
        const aspectRatio = aspectRatios[p.src]
        const geometry =
          aspectRatio && panelHeight && frameWidths
            ? focalGeometry(aspectRatio, frameWidths[shapeOf(aspectRatio)], panelHeight, p.focus)
            : undefined
        return (
          <img
            key={p.src}
            src={p.src}
            alt={p.alt}
            onLoad={(e) => {
              const { naturalWidth: w, naturalHeight: h } = e.currentTarget
              if (w && h) {
                setAspectRatios((ratios) =>
                  ratios[p.src] === w / h ? ratios : { ...ratios, [p.src]: w / h },
                )
              }
            }}
            style={geometry}
            className={cn(
              'absolute transition-opacity duration-[1200ms] ease-in-out',
              // max-w-none: preflight caps images at 100%, which would pull the
              // width back onto the animating panel. Cover the panel the old way
              // until the natural size is known.
              geometry ? 'max-w-none' : 'inset-0 size-full object-cover',
              p.src === activeSrc ? 'opacity-100' : 'opacity-0',
            )}
          />
        )
      })}
      <button
        type="button"
        aria-label="Next photo"
        onClick={() => {
          onAdvance?.()
        }}
        className="absolute inset-0 size-full cursor-pointer border-0 bg-transparent p-0"
      />
      {current && (current.place || current.takenAt) && (
        <div className="pointer-events-none absolute bottom-hm-28 left-hm-28 flex max-w-[calc(100%-3.5rem*var(--hm-density))] items-start gap-hm-14 rounded-hm-18 bg-hm-scrim py-hm-16 pr-hm-24 pl-hm-18 text-hm-on-scrim">
          <MapPin aria-hidden className="mt-0.5 size-hm-28 shrink-0" strokeWidth={2} />
          <div className="flex min-w-0 flex-col gap-0.5">
            <div className="truncate text-hm-label-lg">{current.place ?? 'Unknown place'}</div>
            {current.takenAt && (
              <div className="text-hm-caption text-hm-on-scrim-2">
                {formatTakenAt(current.takenAt)}
              </div>
            )}
          </div>
        </div>
      )}
      {action}
    </section>
  )
}

/**
 * Sizes an image to cover `frameWidth` × `frameHeight` and offsets it so the
 * focal region lands in the middle of that frame, clamped to the image's edges.
 * The frame is the panel's *resting* width for this photo's shape, never the
 * width mid-animation, so a resize reveals and hides pixels instead of
 * rescaling. Falls back to the centre when a photo has no focal region.
 */
function focalGeometry(
  aspectRatio: number,
  frameWidth: number,
  frameHeight: number,
  focus?: Photo['focus'],
): { width: number; height: number; left: number; top: number } {
  let width = frameHeight * aspectRatio
  let height = frameHeight
  if (width < frameWidth) {
    width = frameWidth
    height = frameWidth / aspectRatio
  }
  const focusX = focus ? focus.x + focus.width / 2 : 0.5
  const focusY = focus ? focus.y + focus.height / 2 : 0.5
  return {
    width,
    height,
    left: clamp(frameWidth / 2 - focusX * width, frameWidth - width, 0),
    top: clamp(frameHeight / 2 - focusY * height, frameHeight - height, 0),
  }
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}
