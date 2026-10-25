import { WEATHER_PROVIDERS } from '../lib/providers'

export function ProviderPicker({ connected }: { connected: boolean }) {
  return (
    <section aria-labelledby="forecast-provider" className="flex flex-col gap-2">
      <h2 id="forecast-provider" className="m-0 text-hm-caption font-semibold">
        Forecast provider
      </h2>
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {WEATHER_PROVIDERS.map((provider) => {
          const selected = provider.available
          return (
            <div
              key={provider.id}
              className={`flex flex-col gap-2 rounded-hm-16 border p-4 ${selected ? 'border-2 border-hm-accent bg-hm-surface' : 'border-hm-line bg-hm-surface opacity-60'}`}
            >
              <strong className="text-hm-body">{provider.name}</strong>
              <span className="text-hm-caption text-hm-ink-2">{provider.detail}</span>
              {!provider.available ? (
                <span className="self-start rounded-full bg-hm-bg px-3 py-1 text-hm-caption font-semibold text-hm-ink-2">
                  Coming later
                </span>
              ) : connected ? (
                <span className="self-start rounded-full bg-hm-accent/15 px-3 py-1 text-hm-caption font-semibold text-[#087557]">
                  Connected
                </span>
              ) : null}
            </div>
          )
        })}
      </div>
    </section>
  )
}
