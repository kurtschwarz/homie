import { MODES } from '../lib/modes'
import { DAY, formatTime, type Segment } from '../lib/schedule'

const TICKS = [0, 6, 12, 18, 24]

export function DayTimeline({ segments, nowMinutes }: { segments: Segment[]; nowMinutes: number }) {
  const nowPercent = (nowMinutes / DAY) * 100
  const first = segments[0]
  const last = segments[segments.length - 1]
  // The day wraps, so the last band continues the first; only the first carries the name.
  const lastContinuesFirst = segments.length > 1 && first?.from === 0 && first.mode === last?.mode

  return (
    <div className="flex flex-col gap-2">
      <div className="relative pt-11">
        <div
          className="absolute top-0 -translate-x-1/2 whitespace-nowrap rounded-full bg-[#d3673a] px-3.5 py-1.5 text-hm-caption font-bold text-white"
          style={{ left: `${Math.min(Math.max(nowPercent, 10), 90)}%` }}
        >
          Now · {formatTime(nowMinutes)}
        </div>
        <div
          role="img"
          aria-label={segments
            .map((s) => `${MODES[s.mode].label} ${formatTime(s.from)} to ${formatTime(s.to % DAY)}`)
            .join(', ')}
          className="flex h-12 overflow-hidden rounded-hm-control"
        >
          {segments.map((segment, index) => {
            const { band, label } = MODES[segment.mode]
            const hideLabel = lastContinuesFirst && index === segments.length - 1
            return (
              <div
                key={`${segment.mode}-${segment.from}`}
                className="flex min-w-0 items-center px-3 text-hm-caption font-bold"
                style={{
                  width: `${((segment.to - segment.from) / DAY) * 100}%`,
                  background: band.bg,
                  color: band.ink,
                }}
              >
                {!hideLabel && <span className="truncate">{label}</span>}
              </div>
            )
          })}
        </div>
        <div
          className="pointer-events-none absolute bottom-0 top-[34px] w-0.5 -translate-x-1/2 bg-[#d3673a]"
          style={{ left: `${nowPercent}%` }}
        />
      </div>
      <div className="relative h-5 text-hm-caption font-semibold text-hm-ink-2">
        {TICKS.map((hour) => (
          <span
            key={hour}
            className={`absolute ${hour === 0 ? '' : hour === 24 ? '-translate-x-full' : '-translate-x-1/2'}`}
            style={{ left: `${(hour / 24) * 100}%` }}
          >
            {formatTime(hour * 60).replace(':00 ', ' ')}
          </span>
        ))}
      </div>
    </div>
  )
}
