import { MODES } from '../lib/modes'
import { formatTime } from '../lib/schedule'
import type { DisplayModeSetting } from '../types'

export function ModeRow({
  setting,
  options,
  disabled,
  onChange,
}: {
  setting: DisplayModeSetting
  options: number[]
  disabled: boolean
  onChange: (patch: Partial<DisplayModeSetting>) => void
}) {
  const mode = MODES[setting.mode]
  const name = mode.label
  return (
    <div
      className={`grid items-center gap-x-4 gap-y-2 sm:grid-cols-[28px_minmax(0,1fr)_120px_160px] ${disabled ? 'opacity-50' : ''}`}
    >
      <span
        aria-hidden
        className="size-7 rounded-full border border-hm-line"
        style={{ background: mode.palette.bg }}
      />
      <div className="min-w-0">
        <p className="m-0 text-hm-body font-semibold">{name}</p>
        <p className="m-0 text-hm-caption text-hm-ink-2">{mode.description}</p>
      </div>
      <select
        aria-label={`${name} starts at`}
        disabled={disabled}
        value={setting.startsAtMinutes}
        onChange={(event) => onChange({ startsAtMinutes: Number(event.target.value) })}
        className="h-11 rounded-hm-control border border-hm-line bg-hm-surface px-3 text-hm-body text-hm-ink"
      >
        {options.map((minutes) => (
          <option key={minutes} value={minutes}>
            {formatTime(minutes)}
          </option>
        ))}
      </select>
      {setting.mode === 'OFF' ? (
        <span className="text-hm-caption text-hm-ink-3">Panel off</span>
      ) : (
        <div className="flex items-center gap-2 text-hm-caption text-hm-ink-2">
          <input
            type="range"
            min={0}
            max={100}
            step={5}
            aria-label={`${name} brightness`}
            disabled={disabled}
            value={setting.brightness}
            onChange={(event) => onChange({ brightness: Number(event.target.value) })}
            className="min-w-0 flex-1 accent-hm-accent"
          />
          <span className="w-9 text-right tabular-nums">{setting.brightness}%</span>
        </div>
      )}
    </div>
  )
}
