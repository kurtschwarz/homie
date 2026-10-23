import type { Swatch } from '@homie/ui/lib/people'

import { type ManagerEvent, onDay, weekDays } from '@/features/events'

/**
 * One-letter weekday labels, indexed by `Date.getDay()`, so the first entry is
 * Sunday whatever day the row starts on. The letters repeat (T, T and S, S),
 * which is fine: the dots read left to right from today, so position tells
 * the days apart.
 */
const INITIALS = ['S', 'M', 'T', 'W', 'T', 'F', 'S']

/**
 * A dot per day for the next seven days, starting today, filled in the
 * person's colour on days they have something. It answers "is this badge
 * doing anything?" at a glance, which is the question the list is really
 * there to settle.
 *
 * The row is one image to assistive technology, described by a count of busy
 * days; the per-day initials and dots are hidden from it.
 */
export function PersonWeek({
  events,
  swatch,
  name,
}: {
  /** Only this person's events. A day is busy if any of these starts on it. */
  events: ManagerEvent[]
  /** Colours the dot on a busy day. Quiet days use the neutral line colour. */
  swatch: Swatch
  /** Read out in the accessible label, e.g. "Lana has events on 3 of the next 7 days". */
  name: string
}) {
  const days = weekDays()
  const busy = days.map((day) => onDay(events, day).length > 0)
  const total = busy.filter(Boolean).length

  return (
    <div
      role="img"
      className="flex gap-1.5"
      aria-label={`${name} has events on ${total} of the next 7 days`}
    >
      {days.map((day, index) => (
        <span
          key={day.toISOString()}
          aria-hidden="true"
          className="flex flex-col items-center gap-1"
        >
          <span className="text-[10px] leading-none text-hm-ink-3">{INITIALS[day.getDay()]}</span>
          <span
            className="size-1.5 rounded-full"
            style={{
              backgroundColor: busy[index] ? `var(--hm-swatch-${swatch})` : 'var(--hm-line)',
            }}
          />
        </span>
      ))}
    </div>
  )
}
