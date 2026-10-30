import goldFrameUrl from '@homie/ui/assets/gold-frame.svg?url'
import drawHintUrl from '@homie/ui/assets/homie-draw-hint.lottie?url'
import { Button } from '@homie/ui/components/button'
import { HomieAnimation } from '@homie/ui/components/homie-animation'
import { PersonBadge } from '@homie/ui/components/person-badge'
import { useIdle } from '@homie/ui/hooks/use-idle'
import { cn } from '@homie/ui/lib/utils'
import {
  Check,
  Eraser,
  Paintbrush,
  Plus,
  Rainbow,
  SlidersHorizontal,
  Spline,
  Trash2,
  Undo2,
  X,
} from 'lucide-react'
import * as React from 'react'
import { Confetti } from './confetti'
import { MixPanel } from './mix-panel'
import { loadLastMix, loadPalette, saveLastMix, savePalette } from './palette'
import { canvasToJpeg, uploadDrawing } from './save-drawing'
import { COLOURS, SIZE_PX, type SizeName, type Tool } from './tools'
import { useArtists } from './use-artists'
import { usePaintCanvas } from './use-paint-canvas'

const IDLE_MS = 3 * 60_000

const RAINBOW = 'conic-gradient(#f00, #ff0, #0f0, #0ff, #00f, #f0f, #f00)'

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
  const pickerRef = React.useRef<HTMLButtonElement>(null)
  const [mixOpen, setMixOpen] = React.useState(false)
  const [lastMix, setLastMix] = React.useState<string | null>(loadLastMix)
  const [palette, setPalette] = React.useState<string[]>(loadPalette)

  const closeMix = React.useCallback(() => setMixOpen(false), [])
  const changePalette = (next: string[]) => {
    setPalette(next)
    savePalette(next)
  }
  const applyMix = (hex: string) => {
    setColor(hex)
    setLastMix(hex)
    saveLastMix(hex)
    setTool((current) => (current === 'eraser' ? 'brush' : current))
    setMixOpen(false)
  }

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

  const { data: artists = [] } = useArtists()
  const [artistId, setArtistId] = React.useState<string | null>(null)
  const chosen = artists.find((person) => person.id === artistId)
  const [review, setReview] = React.useState<{ jpeg: Blob; url: string } | null>(null)
  const [saveState, setSaveState] = React.useState<'idle' | 'saving' | 'failed'>('idle')

  React.useEffect(() => {
    if (!review) return
    return () => URL.revokeObjectURL(review.url)
  }, [review])

  const finish = async () => {
    const canvas = canvasRef.current
    if (!canvas || !dirty) return onClose()
    setSaveState('idle')
    try {
      const jpeg = await canvasToJpeg(canvas)
      setReview({ jpeg, url: URL.createObjectURL(jpeg) })
    } catch {
      setSaveState('failed')
    }
  }

  const keepPainting = () => {
    setReview(null)
    setSaveState('idle')
  }

  const addToSlideshow = async () => {
    const canvas = canvasRef.current
    if (!canvas || !review || saveState === 'saving') return
    setSaveState('saving')
    try {
      await uploadDrawing(
        review.jpeg,
        { width: canvas.width, height: canvas.height },
        new Date(),
        artistId ?? undefined,
      )
      onClose()
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
            <HomieAnimation
              src={drawHintUrl}
              aria-label="A hand drawing a squiggle"
              className="size-[calc(16rem*var(--hm-density))]"
            />
            <p className="m-0 text-hm-lead font-medium">Draw anything with your finger</p>
          </div>
        )}
      </div>

      <aside className="relative flex w-[calc(23rem*var(--hm-density))] flex-none flex-col gap-hm-24">
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
          {mixOpen && (
            <MixPanel
              initial={lastMix ?? color}
              palette={palette}
              onPaletteChange={changePalette}
              onUse={applyMix}
              onClose={closeMix}
              anchorRef={pickerRef}
            />
          )}
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
            <button
              ref={pickerRef}
              type="button"
              aria-label="Mix your own colour"
              aria-expanded={mixOpen}
              aria-pressed={lastMix !== null && color === lastMix}
              onClick={() => setMixOpen((open) => !open)}
              className={cn(
                'relative flex aspect-square cursor-pointer items-center justify-center rounded-full border-2 border-white p-0 [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)',
                lastMix !== null &&
                  color === lastMix &&
                  'outline-3 outline-offset-3 outline-hm-accent',
              )}
              style={{ background: lastMix ?? RAINBOW }}
            >
              {lastMix === null ? (
                <span
                  aria-hidden
                  className="flex size-hm-28 items-center justify-center rounded-full bg-white text-hm-ink shadow-hm-raised"
                >
                  <Plus aria-hidden className="size-hm-20" strokeWidth={2.5} />
                </span>
              ) : (
                <span
                  aria-hidden
                  className="absolute -right-hm-4 -bottom-hm-4 flex size-hm-36 items-center justify-center rounded-full border-2 border-white text-white shadow-hm-raised"
                  style={{ background: RAINBOW }}
                >
                  <Plus aria-hidden className="size-hm-20" strokeWidth={3} />
                </span>
              )}
            </button>
          </div>
        </section>

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
          <Button variant="solid" block onClick={finish}>
            <Check aria-hidden />
            Done
          </Button>
        </div>
      </aside>

      {review && (
        <div className="fixed inset-0 z-20 grid place-items-center overflow-hidden bg-[radial-gradient(ellipse_at_50%_40%,rgba(34,38,48,.95),rgba(8,9,12,.99))]">
          <Confetti />
          <div
            role="dialog"
            aria-modal="true"
            aria-labelledby="paint-review-title"
            className="relative z-10 flex flex-col items-center gap-hm-26 text-white"
          >
            <div className="flex flex-col items-center gap-hm-6 text-center">
              <h2
                id="paint-review-title"
                className="m-0 text-[calc(5.25rem*var(--hm-density))] leading-none font-semibold tracking-[-0.03em]"
              >
                Ta-da!
              </h2>
              <p className="m-0 text-hm-lead text-white/75">
                Add it to the photo slideshow?{artists.length > 0 && ' Who painted it?'}
              </p>
            </div>

            <div className="relative w-[calc(40rem*var(--hm-density))] rotate-2 aspect-[880/600]">
              <img
                src={review.url}
                alt="Your painting"
                className="absolute top-[16%] left-[15.9%] h-[68%] w-[68.2%] bg-white object-cover shadow-[inset_0_0_0_1px_rgba(0,0,0,.2)]"
              />
              <img
                src={goldFrameUrl}
                alt=""
                className="pointer-events-none absolute -top-[10%] -left-[6.82%] h-[120%] w-[113.64%] max-w-none"
              />
              <span className="absolute top-[86.6%] left-1/2 z-1 -translate-x-1/2 -translate-y-1/2 rounded-[0.25rem] bg-[linear-gradient(180deg,#FBE9A8,#D9AE4E_55%,#B8862C)] px-hm-18 py-[0.3125rem] text-[calc(1.0625rem*var(--hm-density))] font-semibold tracking-[.04em] whitespace-nowrap text-[#4A3208] shadow-[inset_0_1px_0_rgba(255,255,255,.6),0_0_0_1px_rgba(90,58,10,.6),0_3px_6px_rgba(0,0,0,.35)]">
                {chosen ? `Painted by ${chosen.name}` : 'Painted on Homie'}
              </span>
            </div>

            {artists.length > 0 && (
              <div className="flex flex-wrap justify-center gap-hm-10">
                {artists.map((person) => (
                  <button
                    key={person.id}
                    type="button"
                    aria-pressed={artistId === person.id}
                    onClick={() =>
                      setArtistId((current) => (current === person.id ? null : person.id))
                    }
                    className={cn(
                      'flex h-[calc(5.5rem*var(--hm-density))] cursor-pointer items-center gap-hm-12 rounded-full border-0 bg-hm-surface pr-hm-24 pl-hm-12 text-hm-label-lg font-semibold text-hm-ink [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)',
                      artistId === person.id &&
                        'shadow-[0_0_0_calc(0.25rem*var(--hm-density))_var(--hm-surface),0_0_0_calc(0.4375rem*var(--hm-density))_var(--hm-accent)]',
                    )}
                  >
                    <PersonBadge
                      person={person}
                      size="md"
                      title={undefined}
                      aria-hidden
                      className="size-hm-64 text-[length:calc(2.25rem*var(--hm-density))]"
                    />
                    {person.name}
                  </button>
                ))}
              </div>
            )}

            {saveState === 'failed' && (
              <p role="status" className="m-0 text-hm-label-lg font-medium text-white">
                Couldn't save. Your drawing is still here, try again.
              </p>
            )}

            <div className="flex w-[calc(38.75rem*var(--hm-density))] gap-hm-16">
              <Button
                variant="outline"
                className="h-[calc(6rem*var(--hm-density))] flex-none px-hm-28"
                onClick={keepPainting}
                disabled={saveState === 'saving'}
              >
                Keep painting
              </Button>
              <Button
                variant="solid"
                className="h-[calc(6rem*var(--hm-density))] flex-1 px-hm-24"
                onClick={addToSlideshow}
                disabled={saveState === 'saving'}
              >
                <Check aria-hidden />
                {saveState === 'saving' ? 'Saving…' : 'Add to slideshow'}
              </Button>
            </div>

            <button
              type="button"
              onClick={onClose}
              disabled={saveState === 'saving'}
              className="-mt-hm-10 h-hm-56 cursor-pointer border-0 bg-transparent p-0 text-hm-label-lg font-medium text-white/70"
            >
              Don't save it
            </button>
          </div>
        </div>
      )}
    </main>
  )
}
