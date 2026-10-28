import homieIdleUrl from '@homie/ui/assets/homie-idle.svg?url'
import { Button } from '@homie/ui/components/button'
import { useIdle } from '@homie/ui/hooks/use-idle'
import { cn } from '@homie/ui/lib/utils'
import {
  Check,
  Download,
  Eraser,
  Paintbrush,
  Rainbow,
  SlidersHorizontal,
  Spline,
  Trash2,
  Undo2,
  X,
} from 'lucide-react'
import * as React from 'react'
import { canvasToJpeg, uploadDrawing } from './save-drawing'
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
  const [settingsOpen, setSettingsOpen] = React.useState(false)
  const settingsButtonRef = React.useRef<HTMLButtonElement>(null)
  const settingsPopoverRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    if (!settingsOpen) return
    const dismiss = (event: PointerEvent) => {
      const target = event.target as Node
      if (settingsPopoverRef.current?.contains(target)) return
      if (settingsButtonRef.current?.contains(target)) return
      setSettingsOpen(false)
    }
    document.addEventListener('pointerdown', dismiss, true)
    return () => document.removeEventListener('pointerdown', dismiss, true)
  }, [settingsOpen])
  const { canvasRef, dirty, canUndo, undo, clear, handlers } = usePaintCanvas({
    tool,
    color,
    size: SIZE_PX[size],
    smooth,
  })

  const [saveState, setSaveState] = React.useState<'idle' | 'saving' | 'saved' | 'failed'>('idle')

  React.useEffect(() => {
    if (dirty) setSaveState((state) => (state === 'saved' ? 'idle' : state))
  }, [dirty])

  const save = async () => {
    const canvas = canvasRef.current
    if (!canvas || saveState === 'saving') return
    setSaveState('saving')
    try {
      const jpeg = await canvasToJpeg(canvas)
      await uploadDrawing(jpeg, { width: canvas.width, height: canvas.height }, new Date())
      clear()
      setSaveState('saved')
    } catch {
      setSaveState('failed')
    }
  }

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

      <aside className="flex w-[calc(23rem*var(--hm-density))] flex-none flex-col gap-hm-24">
        <header className="relative flex items-center justify-between">
          <h1 className="m-0 min-w-0 truncate text-hm-title font-semibold tracking-[-0.035em]">
            Painting
          </h1>
          <div className="flex flex-none items-center gap-hm-8">
            <Button
              ref={settingsButtonRef}
              variant="outline"
              aria-label="Settings"
              aria-expanded={settingsOpen}
              className={cn(
                'size-hm-64 px-0',
                settingsOpen && 'border-hm-accent bg-hm-accent-soft text-hm-accent',
              )}
              onClick={() => setSettingsOpen((open) => !open)}
            >
              <SlidersHorizontal aria-hidden />
            </Button>
            <Button
              variant="outline"
              aria-label="Close painting"
              className="size-hm-64 px-0"
              onClick={onClose}
            >
              <X aria-hidden />
            </Button>
          </div>
          {settingsOpen && (
            <div
              ref={settingsPopoverRef}
              role="dialog"
              aria-label="Advanced settings"
              className="absolute top-full right-0 z-10 mt-hm-12 flex w-[calc(34rem*var(--hm-density))] flex-col gap-hm-20 rounded-hm-panel border border-hm-line bg-hm-surface p-hm-24 shadow-hm-float"
            >
              <div className="flex items-center justify-between">
                <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Advanced</h2>
                <button
                  type="button"
                  onClick={() => setSettingsOpen(false)}
                  className="cursor-pointer border-0 bg-transparent p-0 text-hm-label-lg font-semibold text-hm-accent"
                >
                  Done
                </button>
              </div>
              <div className="flex items-center gap-hm-16">
                <span className="flex size-hm-56 flex-none items-center justify-center rounded-hm-control bg-hm-accent-soft text-hm-accent">
                  <Spline aria-hidden className="size-hm-28" strokeWidth={1.75} />
                </span>
                <div className="flex min-w-0 flex-1 flex-col gap-hm-4">
                  <span className="text-hm-label-lg font-semibold">Smooth lines</span>
                  <span className="text-hm-caption text-hm-ink-2">
                    {smooth
                      ? 'On. Wobbly strokes come out as smooth curves.'
                      : 'Off. Strokes follow your finger exactly.'}
                  </span>
                </div>
                <button
                  type="button"
                  role="switch"
                  aria-checked={smooth}
                  aria-label="Smooth lines"
                  onClick={() => setSmooth((value) => !value)}
                  className={cn(
                    'relative h-hm-40 w-hm-64 flex-none cursor-pointer rounded-full border-0 p-0 transition-colors',
                    smooth ? 'bg-hm-accent' : 'bg-hm-line',
                  )}
                >
                  <span
                    className={cn(
                      'absolute top-hm-4 size-hm-32 rounded-full bg-white shadow-hm-raised transition-[left]',
                      smooth ? 'left-[calc(100%-2.25rem*var(--hm-density))]' : 'left-hm-4',
                    )}
                  />
                </button>
              </div>
            </div>
          )}
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

        <div className="mt-auto flex flex-col gap-hm-12">
          {saveState !== 'idle' && (
            <p
              role="status"
              className={cn(
                'm-0 text-center text-hm-label-lg font-medium',
                saveState === 'failed' ? 'text-hm-accent' : 'text-hm-ink-2',
              )}
            >
              {saveState === 'saving' && 'Saving…'}
              {saveState === 'saved' && 'Saved! It will show up with the photos.'}
              {saveState === 'failed' && "Couldn't save. Your drawing is still here, try again."}
            </p>
          )}
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
          <Button
            variant="outline"
            block
            disabled={!dirty || saveState === 'saving'}
            onClick={save}
          >
            <Download aria-hidden />
            Save
          </Button>
          <Button variant="solid" block onClick={onClose}>
            <Check aria-hidden />
            Done
          </Button>
        </div>
      </aside>
    </main>
  )
}
