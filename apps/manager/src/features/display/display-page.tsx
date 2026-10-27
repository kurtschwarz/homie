import { useQuery } from '@tanstack/react-query'
import { useCallback, useEffect, useRef, useState } from 'react'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import { ToggleSwitch } from '@/shared/ui/toggle-switch'
import { useUpdateDisplaySchedule } from './api/display.mutations'
import { displayQueries } from './api/display.queries'
import { DayTimeline } from './components/day-timeline'
import { ModeRow } from './components/mode-row'
import { RightNow } from './components/right-now'
import { WifiCard } from './components/wifi-card'
import { orderModes, overnightSummary, segments, startOptions } from './lib/schedule'
import type { DisplayModeSetting, DisplaySchedule } from './types'

const SAVE_DELAY_MS = 600

function minutesNow() {
  const now = new Date()
  return now.getHours() * 60 + now.getMinutes()
}

function useNowMinutes() {
  const [minutes, setMinutes] = useState(minutesNow)
  useEffect(() => {
    const timer = setInterval(() => setMinutes(minutesNow()), 30_000)
    return () => clearInterval(timer)
  }, [])
  return minutes
}

function DisplayContent({ initial }: { initial: DisplaySchedule }) {
  const [draft, setDraft] = useState<DisplaySchedule>(() => ({
    ...initial,
    modes: orderModes(initial.modes),
  }))
  const [unsaved, setUnsaved] = useState(false)
  const { mutate, isPending, isError, error, isSuccess } = useUpdateDisplaySchedule()
  const nowMinutes = useNowMinutes()
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const pending = useRef<DisplaySchedule | null>(null)

  const flush = useCallback(() => {
    clearTimeout(timer.current)
    if (pending.current) mutate(pending.current)
    pending.current = null
    setUnsaved(false)
  }, [mutate])

  // Flush on unmount so leaving the page mid-drag does not drop the last edit.
  useEffect(() => flush, [flush])

  function update(next: DisplaySchedule) {
    setDraft(next)
    setUnsaved(true)
    pending.current = next
    clearTimeout(timer.current)
    timer.current = setTimeout(flush, SAVE_DELAY_MS)
  }

  function updateMode(index: number, patch: Partial<DisplayModeSetting>) {
    update({
      ...draft,
      modes: draft.modes.map((setting, i) => (i === index ? { ...setting, ...patch } : setting)),
    })
  }

  const status =
    unsaved || isPending
      ? 'Saving…'
      : isError
        ? `Not saved: ${error.message}`
        : isSuccess
          ? 'Saved to Homie'
          : ''

  return (
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
      <section
        aria-labelledby="schedule-heading"
        className="flex flex-col gap-5 rounded-hm-16 border border-hm-line bg-hm-surface p-5"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 flex-col gap-1">
            <h2 id="schedule-heading" className="m-0 text-hm-title-sm">
              Through the day
            </h2>
            <p className="m-0 text-hm-body text-hm-ink-2">
              Homie switches between these on its own. Each mode keeps its own brightness.
            </p>
          </div>
          <span role="status" className="shrink-0 text-hm-caption text-hm-ink-2">
            {status}
          </span>
        </div>
        <DayTimeline segments={segments(draft)} nowMinutes={nowMinutes} />
        <div className="flex flex-col gap-4">
          {draft.modes.map((setting, index) => (
            <ModeRow
              key={setting.mode}
              setting={setting}
              options={startOptions(draft.modes, index)}
              disabled={setting.mode === 'OFF' && !draft.powerSavingEnabled}
              onChange={(patch) => updateMode(index, patch)}
            />
          ))}
        </div>
        <div className="flex flex-col gap-3 border-t border-hm-line pt-4">
          <div className="flex items-center justify-between gap-3 text-hm-body">
            Turn the screen off overnight
            <ToggleSwitch
              label="Turn the screen off overnight"
              checked={draft.powerSavingEnabled}
              onCheckedChange={(powerSavingEnabled) => update({ ...draft, powerSavingEnabled })}
            />
          </div>
          <div className="flex items-center justify-between gap-3 text-hm-body">
            Wake the screen for alerts
            <ToggleSwitch
              label="Wake the screen for alerts"
              checked={draft.wakeForAlerts}
              disabled={!draft.powerSavingEnabled}
              onCheckedChange={(wakeForAlerts) => update({ ...draft, wakeForAlerts })}
            />
          </div>
        </div>
        <p className="m-0 text-hm-caption text-hm-ink-2">{overnightSummary(draft)}</p>
      </section>
      <div className="flex flex-col gap-6">
        <WifiCard />
        <RightNow schedule={draft} nowMinutes={nowMinutes} />
      </div>
    </div>
  )
}

export function DisplayPage() {
  const result = useQuery(displayQueries.schedule())

  return (
    <Page title="Display" description="How the kitchen screen changes through the day">
      {result.isPending ? (
        <p className="m-0 text-hm-body text-hm-ink-2">Loading display settings…</p>
      ) : result.error ? (
        <ErrorMessage label="Display unavailable" message={result.error.message} />
      ) : (
        <DisplayContent initial={result.data} />
      )}
    </Page>
  )
}
