import { CONDITION_LABELS } from '../lib/providers'
import type { WeatherLocation, WeatherReading } from '../types'

function updatedLabel(fetchedAt: number) {
  const minutes = Math.max(0, Math.floor((Date.now() / 1000 - fetchedAt) / 60))
  if (minutes < 1) return 'Updated just now'
  if (minutes < 60) return `Updated ${minutes} min ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `Updated ${hours} hr ago`
  return `Updated ${Math.floor(hours / 24)} d ago`
}

export function ScreenPreview({
  reading,
  location,
}: {
  reading: WeatherReading | null | undefined
  location: WeatherLocation | undefined
}) {
  return (
    <section
      aria-labelledby="weather-preview-heading"
      className="flex flex-col gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-5"
    >
      <h2 id="weather-preview-heading" className="m-0 text-hm-caption font-semibold">
        On the kitchen screen
      </h2>
      <div className="rounded-hm-16 bg-hm-ink px-5 py-4 text-white">
        {reading ? (
          <>
            <p className="m-0 text-[44px] font-bold leading-none">{reading.temperature}°</p>
            <p className="m-0 mt-2 text-hm-caption">
              {CONDITION_LABELS[reading.condition]}
              {location ? ` · ${location.label}` : ''}
            </p>
          </>
        ) : (
          <p className="m-0 text-hm-body">No reading yet</p>
        )}
      </div>
      <p className="m-0 text-hm-caption text-hm-ink-2">
        {reading
          ? `${updatedLabel(reading.fetchedAt)}${reading.stale ? ' · out of date' : ''}`
          : 'The first reading appears after the next refresh.'}
      </p>
    </section>
  )
}
