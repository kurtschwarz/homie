import { useQuery } from '@tanstack/react-query'
import { displayQueries } from '../api/display.queries'

export function DdcCard({ bus, onSelect }: { bus: string; onSelect: (bus: string) => void }) {
  const { data, error, isFetching, refetch } = useQuery(displayQueries.ddcBuses())
  const scanned = data !== undefined
  const known = data?.some((device) => device.bus === bus) ?? false

  return (
    <section
      aria-labelledby="ddc-heading"
      className="flex flex-col gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-5"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col gap-1">
          <h2 id="ddc-heading" className="m-0 text-hm-caption font-semibold">
            Panel control
          </h2>
          <p className="m-0 text-hm-caption text-hm-ink-2">
            The monitor Homie dims and switches off over DDC/CI.
          </p>
        </div>
        <button
          type="button"
          disabled={isFetching}
          onClick={() => refetch()}
          className="h-9 shrink-0 cursor-pointer rounded-hm-control border border-hm-line bg-hm-surface px-3 text-hm-caption font-semibold text-hm-ink disabled:cursor-not-allowed disabled:opacity-60"
        >
          {isFetching ? 'Scanning…' : scanned ? 'Scan again' : 'Scan for displays'}
        </button>
      </div>
      {error && (
        <p role="alert" className="m-0 text-hm-caption text-hm-ink-2">
          Scan failed: {error.message}
        </p>
      )}
      {!scanned && !isFetching && (
        <p className="m-0 text-hm-body text-hm-ink-2">
          {bus
            ? `Using I2C bus ${bus}.`
            : 'No display selected, so brightness and power are not controlled.'}
        </p>
      )}
      {isFetching && <p className="m-0 text-hm-body text-hm-ink-2">Looking for displays…</p>}
      {scanned && !isFetching && (
        <fieldset className="m-0 flex flex-col gap-2 border-0 p-0">
          <legend className="sr-only">Display to control</legend>
          {data.length === 0 && (
            <p className="m-0 text-hm-body text-hm-ink-2">No DDC-capable displays found.</p>
          )}
          {data.map((device) => (
            <label
              key={device.bus}
              className="flex cursor-pointer items-center gap-3 rounded-hm-control border border-hm-line px-3 py-2"
            >
              <input
                type="radio"
                name="ddc-bus"
                checked={bus === device.bus}
                onChange={() => onSelect(device.bus)}
                className="accent-hm-accent"
              />
              <span className="min-w-0">
                <span className="block truncate text-hm-body font-semibold">{device.monitor}</span>
                <span className="block text-hm-caption text-hm-ink-2">
                  {device.connector} · I2C bus {device.bus}
                </span>
              </span>
            </label>
          ))}
          {bus && !known && (
            <p className="m-0 text-hm-caption text-hm-ink-2">
              The saved bus {bus} was not found in this scan.
            </p>
          )}
          <label className="flex cursor-pointer items-center gap-3 rounded-hm-control border border-hm-line px-3 py-2">
            <input
              type="radio"
              name="ddc-bus"
              checked={bus === ''}
              onChange={() => onSelect('')}
              className="accent-hm-accent"
            />
            <span className="text-hm-body">None (turn off panel control)</span>
          </label>
        </fieldset>
      )}
    </section>
  )
}
