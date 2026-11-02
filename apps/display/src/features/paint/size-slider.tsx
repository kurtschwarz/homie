import * as React from 'react'
import { BRUSH_MAX, BRUSH_MIN, BRUSH_SNAPS, clampBrushWidth, snapBrushWidth } from './tools'

const INSET = 'calc(2rem * var(--hm-density))'
const SPAN = BRUSH_MAX - BRUSH_MIN

const fraction = (width: number) => (width - BRUSH_MIN) / SPAN
const along = (width: number) => `calc(${INSET} + (100% - 2 * ${INSET}) * ${fraction(width)})`

function describe(width: number) {
  if (width <= 8) return 'Thin'
  if (width <= 30) return 'Medium'
  if (width <= 52) return 'Thick'
  return 'Extra thick'
}

export function SizeSlider({
  value,
  color,
  onChange,
}: {
  value: number
  color: string
  onChange: (value: number) => void
}) {
  const trackRef = React.useRef<HTMLDivElement>(null)
  const dragging = React.useRef<number | null>(null)

  const widthAt = (clientX: number) => {
    const rect = trackRef.current?.getBoundingClientRect()
    if (!rect) return value
    const inset = parseFloat(getComputedStyle(trackRef.current as Element).paddingLeft) || 0
    const left = rect.left + inset
    const span = rect.width - inset * 2
    const ratio = Math.min(1, Math.max(0, (clientX - left) / span))
    return clampBrushWidth(BRUSH_MIN + ratio * SPAN)
  }

  const onPointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    dragging.current = event.pointerId
    onChange(widthAt(event.clientX))
  }

  const onPointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    if (dragging.current !== event.pointerId) return
    onChange(widthAt(event.clientX))
  }

  const onPointerEnd = (event: React.PointerEvent<HTMLDivElement>) => {
    if (dragging.current !== event.pointerId) return
    dragging.current = null
    onChange(snapBrushWidth(widthAt(event.clientX)))
  }

  const onKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    const step = event.key === 'ArrowRight' || event.key === 'ArrowUp' ? 2 : -2
    if (!event.key.startsWith('Arrow')) return
    event.preventDefault()
    onChange(clampBrushWidth(value + step))
  }

  return (
    <div
      ref={trackRef}
      role="slider"
      tabIndex={0}
      aria-label="Brush size"
      aria-orientation="horizontal"
      aria-valuemin={BRUSH_MIN}
      aria-valuemax={BRUSH_MAX}
      aria-valuenow={value}
      aria-valuetext={`${describe(value)}, ${value} pixels`}
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onPointerUp={onPointerEnd}
      onPointerCancel={onPointerEnd}
      onKeyDown={onKeyDown}
      className="relative h-[calc(6rem*var(--hm-density))] w-[calc(23.75rem*var(--hm-density))] flex-none cursor-pointer touch-none px-hm-32 outline-none focus-visible:outline-3 focus-visible:outline-hm-accent"
    >
      <div className="absolute inset-x-hm-32 top-1/2 h-hm-8 -translate-y-1/2 rounded-hm-3 bg-hm-line" />
      <div
        className="absolute left-hm-32 top-1/2 h-hm-8 -translate-y-1/2 rounded-hm-3"
        style={{ width: `calc(${along(value)} - ${INSET})`, background: color }}
      />
      {BRUSH_SNAPS.map((snap) => (
        <span
          key={snap}
          aria-hidden
          className="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full shadow-[0_0_0_1px_rgba(0,0,0,.14)]"
          style={{
            left: along(snap),
            width: `calc(${snap / 16}rem * var(--hm-density))`,
            height: `calc(${snap / 16}rem * var(--hm-density))`,
            background: color,
            opacity: value === snap ? 1 : 0.35,
          }}
        />
      ))}
      <span
        aria-hidden
        className="pointer-events-none absolute top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white shadow-hm-raised"
        style={{
          left: along(value),
          width: `calc(${Math.max(30, value) / 16}rem * var(--hm-density))`,
          height: `calc(${Math.max(30, value) / 16}rem * var(--hm-density))`,
          background: color,
        }}
      />
    </div>
  )
}
