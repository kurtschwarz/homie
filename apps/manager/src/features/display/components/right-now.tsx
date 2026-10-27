import { useEffect, useState } from 'react'
import { usePreviewDisplayMode } from '../api/display.mutations'
import { MODE_ORDER, MODES } from '../lib/modes'
import { brightnessOf, formatTime, modeAt } from '../lib/schedule'
import type { DisplayModeKey, DisplaySchedule } from '../types'

const PREVIEW_SECONDS = 10

type Choice = 'NOW' | DisplayModeKey

function MockDevice({
  mode,
  brightness,
  nowMinutes,
}: {
  mode: DisplayModeKey
  brightness: number
  nowMinutes: number
}) {
  const { palette } = MODES[mode]
  const clock = formatTime(nowMinutes).replace(/ [AP]M$/, '')
  const date = new Date().toLocaleDateString(undefined, {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  })
  return (
    <div className="rounded-hm-16 bg-black p-2">
      <div
        className="flex aspect-video flex-col justify-between rounded-hm-control p-4"
        style={{
          background: palette.bg,
          color: palette.ink,
          filter: `brightness(${0.35 + (brightness / 100) * 0.65})`,
        }}
      >
        {mode === 'OFF' ? null : (
          <>
            <div>
              <p className="m-0 text-[32px] font-bold leading-none">{clock}</p>
              <p className="m-0 mt-1 text-hm-caption" style={{ color: palette.ink2 }}>
                {date}
              </p>
            </div>
            <div className="flex items-end gap-3">
              {mode !== 'NIGHT' && (
                <div
                  className="size-14 rounded-hm-control"
                  style={{ background: palette.surface }}
                />
              )}
              <p className="m-0 text-hm-caption">
                <span style={{ color: palette.accent }}>Next</span> · Soccer practice, 4:30 PM
              </p>
            </div>
          </>
        )}
      </div>
    </div>
  )
}

export function RightNow({
  schedule,
  nowMinutes,
}: {
  schedule: DisplaySchedule
  nowMinutes: number
}) {
  const [choice, setChoice] = useState<Choice>('NOW')
  const [running, setRunning] = useState(false)
  const preview = usePreviewDisplayMode()

  const shown = choice === 'NOW' ? modeAt(schedule, nowMinutes) : choice

  // The display polls, so the preview takes up to 2s to appear and 2s to clear.
  useEffect(() => {
    if (!running) return
    const timer = setTimeout(() => setRunning(false), (PREVIEW_SECONDS + 2) * 1000)
    return () => clearTimeout(timer)
  }, [running])

  function show() {
    setRunning(true)
    preview.mutate({ mode: shown, seconds: PREVIEW_SECONDS }, { onError: () => setRunning(false) })
  }

  return (
    <section
      aria-labelledby="right-now-heading"
      className="flex flex-col gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-5"
    >
      <h2 id="right-now-heading" className="m-0 text-hm-caption font-semibold">
        Right now
      </h2>
      <fieldset aria-label="Preview mode" className="m-0 flex flex-wrap gap-2 border-0 p-0">
        {(['NOW', ...MODE_ORDER] as Choice[]).map((option) => (
          <button
            key={option}
            type="button"
            aria-pressed={choice === option}
            onClick={() => setChoice(option)}
            className={`h-9 cursor-pointer rounded-full border px-4 text-hm-caption font-semibold ${choice === option ? 'border-hm-accent bg-hm-accent-soft text-hm-ink' : 'border-hm-line bg-hm-surface text-hm-ink-2'}`}
          >
            {option === 'NOW' ? 'Now' : MODES[option].switcherLabel}
          </button>
        ))}
      </fieldset>
      <MockDevice mode={shown} brightness={brightnessOf(schedule, shown)} nowMinutes={nowMinutes} />
      <button
        type="button"
        disabled={running}
        onClick={show}
        className="h-11 cursor-pointer rounded-hm-control border border-hm-line bg-hm-surface px-4 text-hm-body font-semibold text-hm-ink disabled:cursor-not-allowed disabled:opacity-60"
      >
        {running
          ? `Showing ${MODES[shown].label} on Homie…`
          : `Show ${MODES[shown].label} on Homie for ${PREVIEW_SECONDS} seconds`}
      </button>
      {preview.error && (
        <p role="alert" className="m-0 text-hm-caption text-hm-ink-2">
          Could not preview: {preview.error.message}
        </p>
      )}
    </section>
  )
}
