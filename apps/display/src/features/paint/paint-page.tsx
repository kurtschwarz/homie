import homieIdleUrl from '@homie/ui/assets/homie-idle.svg?url'
import { Button } from '@homie/ui/components/button'
import { useIdle } from '@homie/ui/hooks/use-idle'
import { cn } from '@homie/ui/lib/utils'
import { Check, Eraser, Paintbrush, Rainbow, Spline, Trash2, Undo2, X } from 'lucide-react'
import * as React from 'react'
import { COLOURS, SIZE_PX, type SizeName, type Tool } from './tools'
import { usePaintCanvas } from './use-paint-canvas'

const IDLE_MS = 3 * 60_000

const TOOL_OPTIONS: { tool: Tool; label: string; icon: typeof Paintbrush }[] = [
  { tool: 'brush', label: 'Brush', icon: Paintbrush },
  { tool: 'rainbow', label: 'Rainbow', icon: Rainbow },
  { tool: 'eraser', label: 'Eraser', icon: Eraser },
]

const SIZE_OPTIONS: { size: SizeName; dot: string }[] = [
  { size: 'small', dot: 'size-hm-10' },
  { size: 'medium', dot: 'size-hm-20' },
  { size: 'large', dot: 'size-hm-36' },
]

const tile =
  'flex cursor-pointer flex-col items-center justify-center gap-hm-2 rounded-hm-control leading-none border-2 bg-hm-surface text-[calc(1.125rem*var(--hm-density))] font-medium [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)'

export function PaintPage({ onClose }: { onClose: () => void }) {
  const [tool, setTool] = React.useState<Tool>('brush')
  const [size, setSize] = React.useState<SizeName>('medium')
  const [color, setColor] = React.useState(COLOURS[7]?.value ?? '#5590db')
  const [smooth, setSmooth] = React.useState(true)
  const { canvasRef, dirty, canUndo, undo, clear, handlers } = usePaintCanvas({
    tool,
    color,
    size: SIZE_PX[size],
    smooth,
  })

  useIdle(true, onClose, IDLE_MS)

  return (
    <main className="fixed inset-0 flex h-dvh w-screen gap-hm-column bg-hm-shell p-hm-screen text-hm-ink">
      <div className="relative min-w-0 flex-1 overflow-hidden rounded-hm-panel bg-white shadow-hm-float">
        <canvas
          ref={canvasRef}
          aria-label="Drawing canvas"
          className="absolute inset-0 size-full touch-none"
          {...handlers}
        />
        {!dirty && (
          <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center gap-hm-12 text-hm-ink-3">
            <img src={homieIdleUrl} alt="" className="h-hm-128 w-auto" />
            <p className="m-0 text-hm-lead font-medium">Draw anything with your finger</p>
          </div>
        )}
      </div>

      <aside className="flex w-[calc(21rem*var(--hm-density))] flex-none flex-col gap-hm-24">
        <header className="flex items-center justify-between">
          <h1 className="m-0 text-hm-title-lg font-semibold tracking-[-0.035em]">Painting</h1>
          <Button
            variant="outline"
            aria-label="Close painting"
            className="size-hm-76 px-0"
            onClick={onClose}
          >
            <X aria-hidden />
          </Button>
        </header>

        <section className="flex flex-col gap-hm-12">
          <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Tools</h2>
          <div className="grid grid-cols-3 gap-hm-12">
            {TOOL_OPTIONS.map(({ tool: option, label, icon: Icon }) => (
              <button
                key={option}
                type="button"
                aria-pressed={tool === option}
                onClick={() => setTool(option)}
                className={cn(
                  tile,
                  'h-[calc(6.5rem*var(--hm-density))]',
                  tool === option
                    ? 'border-hm-accent bg-hm-accent-soft text-hm-accent'
                    : 'border-hm-line text-hm-ink-2',
                )}
              >
                <Icon aria-hidden className="size-hm-40" strokeWidth={1.75} />
                {label}
              </button>
            ))}
          </div>
        </section>

        <section className="flex flex-col gap-hm-12">
          <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Size</h2>
          <div className="grid grid-cols-3 gap-hm-12">
            {SIZE_OPTIONS.map(({ size: option, dot }) => (
              <button
                key={option}
                type="button"
                aria-label={`${option} size`}
                aria-pressed={size === option}
                onClick={() => setSize(option)}
                className={cn(
                  tile,
                  'h-[calc(5rem*var(--hm-density))]',
                  size === option ? 'border-hm-accent bg-hm-accent-soft' : 'border-hm-line',
                )}
              >
                <span
                  aria-hidden
                  className={cn('rounded-full', dot)}
                  style={{ background: tool === 'eraser' ? 'var(--hm-ink-3)' : color }}
                />
              </button>
            ))}
          </div>
        </section>

        <section className="flex flex-col gap-hm-12">
          <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Colours</h2>
          <div className="grid grid-cols-4 gap-hm-12">
            {COLOURS.map(({ name, value }) => (
              <button
                key={value}
                type="button"
                aria-label={name}
                aria-pressed={color === value}
                onClick={() => {
                  setColor(value)
                  if (tool === 'eraser') setTool('brush')
                }}
                className={cn(
                  'aspect-square cursor-pointer rounded-full border border-hm-line [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)',
                  color === value && 'outline-3 outline-offset-3 outline-hm-accent',
                )}
                style={{ background: value }}
              />
            ))}
          </div>
        </section>

        <button
          type="button"
          aria-pressed={smooth}
          onClick={() => setSmooth((value) => !value)}
          className={cn(
            tile,
            'h-hm-64 flex-row',
            smooth
              ? 'border-hm-accent bg-hm-accent-soft text-hm-accent'
              : 'border-hm-line text-hm-ink-2',
          )}
        >
          <Spline aria-hidden className="size-hm-28" strokeWidth={1.75} />
          Smooth lines
        </button>

        <div className="mt-auto flex flex-col gap-hm-12">
          <div className="grid grid-cols-2 gap-hm-12">
            <Button variant="outline" disabled={!canUndo} onClick={undo}>
              <Undo2 aria-hidden />
              Undo
            </Button>
            <Button variant="outline" disabled={!dirty} onClick={clear}>
              <Trash2 aria-hidden />
              Clear
            </Button>
          </div>
          <Button variant="solid" block onClick={onClose}>
            <Check aria-hidden />
            Done
          </Button>
        </div>
      </aside>
    </main>
  )
}
