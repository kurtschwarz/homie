import { ToggleSwitch } from '@/shared/ui/toggle-switch'
import type { Calendar } from '../types'

export function CalendarRow({
  calendar,
  pending,
  onVisibleChange,
}: {
  calendar: Calendar
  pending: boolean
  onVisibleChange: (visible: boolean) => void
}) {
  return (
    <div
      className={`grid min-h-13 grid-cols-[minmax(0,1fr)_auto] items-center gap-4 border-t border-hm-line px-4 py-3 sm:grid-cols-[220px_minmax(200px,1fr)_auto] ${calendar.visible ? '' : 'text-hm-ink-3'}`}
    >
      <div className="flex min-w-0 items-center gap-3">
        <span className="min-w-0">
          <span className="block truncate text-hm-body font-semibold">{calendar.name}</span>
          <span className="block text-hm-caption sm:hidden">{calendar.eventCount} this week</span>
        </span>
      </div>
      <span className="hidden whitespace-nowrap text-hm-caption sm:block">
        {calendar.eventCount} this week
      </span>
      <div className="flex items-center gap-3 justify-self-end">
        <span className="hidden text-hm-caption sm:inline">
          {calendar.visible ? 'On the kitchen screen' : 'Hidden'}
        </span>
        <ToggleSwitch
          checked={calendar.visible}
          disabled={pending}
          label={`${calendar.visible ? 'Hide' : 'Show'} ${calendar.name} on the kitchen screen`}
          onCheckedChange={onVisibleChange}
        />
      </div>
    </div>
  )
}
