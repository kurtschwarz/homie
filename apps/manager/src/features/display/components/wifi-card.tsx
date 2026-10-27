import { useQuery } from '@tanstack/react-query'
import { Wifi, WifiOff } from 'lucide-react'
import { displayQueries } from '../api/display.queries'

function strength(percent: number) {
  if (percent >= 75) return 'Excellent'
  if (percent >= 50) return 'Good'
  if (percent >= 25) return 'Fair'
  return 'Weak'
}

export function WifiCard() {
  const { data, isPending, error } = useQuery({
    ...displayQueries.wifi(),
    refetchInterval: 30_000,
  })

  return (
    <section
      aria-labelledby="wifi-heading"
      className="flex flex-col gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-5"
    >
      <h2 id="wifi-heading" className="m-0 text-hm-caption font-semibold">
        Wi-Fi
      </h2>
      {isPending ? (
        <p className="m-0 text-hm-body text-hm-ink-2">Checking…</p>
      ) : error ? (
        <p className="m-0 text-hm-body text-hm-ink-2">Wi-Fi status unavailable: {error.message}</p>
      ) : data.connected ? (
        <div className="flex items-center gap-3">
          <Wifi aria-hidden className="size-6 text-hm-accent" />
          <div className="min-w-0">
            <p className="m-0 truncate text-hm-body font-semibold">{data.ssid}</p>
            <p className="m-0 text-hm-caption text-hm-ink-2">
              {strength(data.signalPercent)} signal · {data.signalPercent}%
            </p>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-3">
          <WifiOff aria-hidden className="size-6 text-hm-ink-3" />
          <div>
            <p className="m-0 text-hm-body font-semibold">Nothing to report</p>
            <p className="m-0 text-hm-caption text-hm-ink-2">
              Wi-Fi is only reported when the API runs on the Homie itself.
            </p>
          </div>
        </div>
      )}
    </section>
  )
}
