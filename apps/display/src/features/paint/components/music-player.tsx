import { cn } from '@homie/ui/lib/utils'
import { Pause, Play, SkipBack, SkipForward, Volume2, X } from 'lucide-react'
import * as React from 'react'

const px = (value: number) => `calc(${value / 16}rem * var(--hm-density))`

const RECORD_FACE =
  'radial-gradient(circle, var(--hm-accent) 0 16%, #1b1c20 17% 30%, #2a2c32 31% 34%, #1b1c20 35% 60%, #2a2c32 61% 64%, #1b1c20 65%)'

function Record({ playing, size }: { playing: boolean; size: number }) {
  return (
    <span
      aria-hidden
      className="block flex-none rounded-full motion-safe:animate-[spin_2.2s_linear_infinite]"
      style={{
        width: px(size),
        height: px(size),
        background: RECORD_FACE,
        animationPlayState: playing ? 'running' : 'paused',
      }}
    />
  )
}

function Tonearm({ playing }: { playing: boolean }) {
  return (
    <span
      aria-hidden
      className="pointer-events-none absolute -top-hm-4 -right-hm-6 block h-hm-52 w-hm-6 origin-top rounded-full bg-hm-ink-3 transition-transform duration-400 ease-out motion-reduce:transition-none"
      style={{ transform: `rotate(${playing ? 12 : -24}deg)` }}
    />
  )
}

function IconButton({
  label,
  size,
  accent,
  onClick,
  children,
}: {
  label: string
  size: number
  accent?: boolean
  onClick?: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      aria-label={label}
      onClick={onClick}
      className={cn(
        'flex flex-none cursor-pointer items-center justify-center rounded-full border-0 p-0 [-webkit-tap-highlight-color:transparent] active:scale-(--hm-press-scale)',
        accent ? 'bg-hm-accent text-hm-on-accent' : 'bg-transparent text-hm-ink',
      )}
      style={{ width: px(size), height: px(size) }}
    >
      {children}
    </button>
  )
}

export function MusicPlayer() {
  const [playing, setPlaying] = React.useState(false)
  const [open, setOpen] = React.useState(false)
  const rootRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    if (!open) return
    const dismiss = (event: PointerEvent) => {
      if (rootRef.current?.contains(event.target as Node)) return
      setOpen(false)
    }
    document.addEventListener('pointerdown', dismiss, true)
    return () => document.removeEventListener('pointerdown', dismiss, true)
  }, [open])

  const status = playing ? 'Playing' : 'Paused'
  const toggle = () => setPlaying((value) => !value)
  const Icon = playing ? Pause : Play

  return (
    <div ref={rootRef} className="relative flex h-full items-center gap-hm-16">
      {open && (
        <div
          role="dialog"
          aria-label="Music"
          className="absolute bottom-full left-0 z-10 mb-hm-12 flex flex-col items-center gap-hm-20 rounded-hm-panel border border-hm-line bg-hm-surface p-hm-24 shadow-hm-float"
          style={{ width: px(400) }}
        >
          <div className="flex w-full items-center justify-between">
            <h2 className="m-0 text-hm-eyebrow text-hm-ink-2 uppercase">Music</h2>
            <button
              type="button"
              aria-label="Close music"
              onClick={() => setOpen(false)}
              className="flex cursor-pointer items-center justify-center rounded-full border-0 bg-transparent p-0 text-hm-ink-2"
              style={{ width: px(60), height: px(60) }}
            >
              <X aria-hidden className="size-hm-28" />
            </button>
          </div>
          <div className="relative">
            <Record playing={playing} size={170} />
          </div>
          <div className="flex flex-col items-center gap-hm-4">
            <span className="font-semibold" style={{ fontSize: px(26) }}>
              Family playlist
            </span>
            <span className="text-hm-ink-2" style={{ fontSize: px(19) }}>
              {status}
            </span>
          </div>
          <div className="h-hm-8 w-full rounded-hm-3 bg-hm-line">
            <div className="h-full w-[35%] rounded-hm-3 bg-hm-accent" />
          </div>
          <div className="flex items-center justify-center gap-hm-22">
            <IconButton label="Previous" size={72}>
              <SkipBack aria-hidden className="size-hm-32" />
            </IconButton>
            <IconButton label={playing ? 'Pause' : 'Play'} size={96} accent onClick={toggle}>
              <Icon aria-hidden className="size-hm-40" />
            </IconButton>
            <IconButton label="Next" size={72}>
              <SkipForward aria-hidden className="size-hm-32" />
            </IconButton>
          </div>
          <div className="flex w-full items-center gap-hm-16 text-hm-ink-2">
            <Volume2 aria-hidden className="size-hm-28 flex-none" />
            <div className="relative flex-1 rounded-full bg-hm-line" style={{ height: px(12) }}>
              <div className="h-full w-[60%] rounded-full bg-hm-accent" />
              <span
                aria-hidden
                className="absolute top-1/2 left-[60%] -translate-x-1/2 -translate-y-1/2 rounded-full bg-white shadow-hm-raised"
                style={{ width: px(34), height: px(34) }}
              />
            </div>
          </div>
        </div>
      )}

      <button
        type="button"
        aria-label="Open music player"
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
        className="relative flex min-w-0 flex-1 cursor-pointer items-center gap-hm-16 border-0 bg-transparent p-0 text-left text-hm-ink [-webkit-tap-highlight-color:transparent]"
      >
        <span className="relative flex-none">
          <Record playing={playing} size={80} />
          <Tonearm playing={playing} />
        </span>
        <span className="flex min-w-0 flex-col gap-hm-2">
          <span className="truncate font-semibold" style={{ fontSize: px(22) }}>
            Family playlist
          </span>
          <span className="text-hm-ink-2" style={{ fontSize: px(18) }}>
            {status}
          </span>
        </span>
      </button>

      <IconButton label={playing ? 'Pause' : 'Play'} size={72} accent onClick={toggle}>
        <Icon aria-hidden className="size-hm-32" />
      </IconButton>
    </div>
  )
}
