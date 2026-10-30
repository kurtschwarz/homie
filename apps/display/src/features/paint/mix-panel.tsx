import { cn } from '@homie/ui/lib/utils'
import { Plus, X } from 'lucide-react'
import * as React from 'react'
import { type Hsv, hexToHsv, hsvToHex } from './colour'
import { addColour, PALETTE_SIZE, removeColour } from './palette'

function useDrag(onMove: (x: number, y: number) => void) {
  const apply = (event: React.PointerEvent<HTMLDivElement>) => {
    const rect = event.currentTarget.getBoundingClientRect()
    const x = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width))
    const y = Math.min(1, Math.max(0, (event.clientY - rect.top) / rect.height))
    onMove(x, y)
  }
  return {
    onPointerDown: (event: React.PointerEvent<HTMLDivElement>) => {
      event.currentTarget.setPointerCapture(event.pointerId)
      apply(event)
    },
    onPointerMove: (event: React.PointerEvent<HTMLDivElement>) => {
      if (event.currentTarget.hasPointerCapture(event.pointerId)) apply(event)
    },
  }
}

export function MixPanel({
  initial,
  palette,
  onPaletteChange,
  onUse,
  onClose,
  anchorRef,
}: {
  initial: string
  palette: string[]
  onPaletteChange: (palette: string[]) => void
  onUse: (hex: string) => void
  onClose: () => void
  anchorRef: React.RefObject<HTMLElement | null>
}) {
  const panelRef = React.useRef<HTMLDivElement>(null)
  const [hsv, setHsv] = React.useState<Hsv>(() => hexToHsv(initial))
  const [editing, setEditing] = React.useState(false)
  const [highlight, setHighlight] = React.useState<number | null>(null)
  const hex = hsvToHex(hsv)

  React.useEffect(() => {
    const dismiss = (event: PointerEvent) => {
      const target = event.target as Node
      if (panelRef.current?.contains(target)) return
      if (anchorRef.current?.contains(target)) return
      onClose()
    }
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('pointerdown', dismiss, true)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', dismiss, true)
      document.removeEventListener('keydown', onKey)
    }
  }, [anchorRef, onClose])

  React.useEffect(() => {
    if (highlight === null) return
    const timer = setTimeout(() => setHighlight(null), 900)
    return () => clearTimeout(timer)
  }, [highlight])

  const squareDrag = useDrag((x, y) => setHsv((current) => ({ ...current, s: x, v: 1 - y })))
  const hueDrag = useDrag((x) => setHsv((current) => ({ ...current, h: x * 360 })))

  const save = () => {
    const result = addColour(palette, hex)
    if (result.palette !== palette) onPaletteChange(result.palette)
    if (result.index !== -1) setHighlight(result.index)
  }

  const slots = Array.from({ length: PALETTE_SIZE }, (_, index) => index)

  return (
    <div
      ref={panelRef}
      role="dialog"
      aria-label="Mix your own colour"
      className="absolute right-full bottom-0 z-10 mr-hm-12 flex w-[calc(30rem*var(--hm-density))] flex-col gap-hm-16 rounded-hm-panel border border-hm-line bg-hm-surface p-hm-24 shadow-hm-float"
    >
      <div className="flex items-center justify-between">
        <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Mix your own</h2>
        <button
          type="button"
          aria-label="Close"
          onClick={onClose}
          className="flex size-hm-40 cursor-pointer items-center justify-center rounded-full border-0 bg-transparent p-0 text-hm-ink-2"
        >
          <X aria-hidden className="size-hm-24" />
        </button>
      </div>

      <div
        role="slider"
        aria-label="Saturation and brightness"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.round(hsv.s * 100)}
        aria-valuetext={hex}
        tabIndex={0}
        {...squareDrag}
        className="relative h-[calc(16rem*var(--hm-density))] w-full touch-none rounded-hm-control"
        style={{
          background: `linear-gradient(to top, #000, transparent), linear-gradient(to right, #fff, hsl(${hsv.h} 100% 50%))`,
        }}
      >
        <span
          aria-hidden
          className="pointer-events-none absolute size-hm-32 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white shadow-hm-raised"
          style={{ left: `${hsv.s * 100}%`, top: `${(1 - hsv.v) * 100}%`, background: hex }}
        />
      </div>

      <div
        role="slider"
        aria-label="Hue"
        aria-valuemin={0}
        aria-valuemax={360}
        aria-valuenow={Math.round(hsv.h)}
        tabIndex={0}
        {...hueDrag}
        className="relative h-hm-40 w-full touch-none rounded-full"
        style={{
          background:
            'linear-gradient(to right, #f00, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00)',
        }}
      >
        <span
          aria-hidden
          className="pointer-events-none absolute top-1/2 size-hm-48 -translate-x-1/2 -translate-y-1/2 rounded-full border-4 border-white shadow-hm-raised"
          style={{ left: `${(hsv.h / 360) * 100}%`, background: `hsl(${hsv.h} 100% 50%)` }}
        />
      </div>

      <div className="flex items-center gap-hm-16">
        <span
          aria-hidden
          className="size-hm-64 flex-none rounded-full border border-hm-line"
          style={{ background: hex }}
        />
        <button
          type="button"
          onClick={() => onUse(hex)}
          className="h-hm-64 flex-1 cursor-pointer rounded-hm-control border-0 bg-hm-accent text-hm-label-lg font-semibold text-white [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)"
        >
          Use it
        </button>
      </div>

      <div className="flex flex-col gap-hm-12 border-t border-hm-line pt-hm-16">
        <div className="flex items-center justify-between">
          <h3 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">My colours</h3>
          {palette.length > 0 && (
            <button
              type="button"
              onClick={() => setEditing((value) => !value)}
              className="cursor-pointer border-0 bg-transparent p-0 text-hm-label-lg font-semibold text-hm-accent"
            >
              {editing ? 'Done' : 'Edit'}
            </button>
          )}
        </div>
        <div className="grid grid-cols-4 gap-hm-12">
          {slots.map((index) => {
            const saved = palette[index]
            if (saved) {
              return (
                <div key={saved} className="relative aspect-square">
                  <button
                    type="button"
                    aria-label={`Saved colour ${saved}`}
                    onClick={() => setHsv(hexToHsv(saved))}
                    className={cn(
                      'size-full cursor-pointer rounded-full border border-hm-line [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)',
                      highlight === index && 'outline-3 outline-offset-3 outline-hm-accent',
                    )}
                    style={{ background: saved }}
                  />
                  {editing && (
                    <button
                      type="button"
                      aria-label={`Delete saved colour ${saved}`}
                      onClick={() => onPaletteChange(removeColour(palette, index))}
                      className="absolute -top-hm-6 -right-hm-6 flex size-hm-32 cursor-pointer items-center justify-center rounded-full border-0 bg-hm-ink p-0 text-white shadow-hm-raised"
                    >
                      <X aria-hidden className="size-hm-20" />
                    </button>
                  )}
                </div>
              )
            }
            if (index === palette.length && !editing) {
              return (
                <button
                  key="add"
                  type="button"
                  aria-label="Save this colour"
                  onClick={save}
                  className="flex aspect-square cursor-pointer items-center justify-center rounded-full border-2 border-dashed border-hm-line bg-transparent p-0 text-hm-ink-2 [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)"
                >
                  <Plus aria-hidden className="size-hm-28" />
                </button>
              )
            }
            return (
              <span
                key={`empty-${index}`}
                aria-hidden
                className="aspect-square rounded-full border border-hm-line bg-hm-line/40"
              />
            )
          })}
        </div>
      </div>
    </div>
  )
}
