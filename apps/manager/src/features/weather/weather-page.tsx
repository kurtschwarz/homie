import { useQuery } from '@tanstack/react-query'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import { weatherQueries } from './api/weather.queries'
import { ProviderPicker } from './components/provider-picker'
import { ScreenPreview } from './components/screen-preview'
import { WeatherKitForm } from './components/weatherkit-form'

export function WeatherPage() {
  const result = useQuery(weatherQueries.list())
  const reading = useQuery({ ...weatherQueries.reading(), refetchInterval: 60_000 })

  const provider = result.data?.weatherProviders[0]
  const location = result.data?.weatherLocations[0]
  const connected = Boolean(provider && provider.lastAttemptedAt !== null && !provider.lastError)

  return (
    <Page title="Weather" description="Where the forecast on the kitchen screen comes from">
      {result.isPending ? (
        <p className="m-0 text-hm-body text-hm-ink-2">Loading weather settings…</p>
      ) : result.error ? (
        <ErrorMessage label="Weather unavailable" message={result.error.message} />
      ) : (
        <>
          <ProviderPicker connected={connected} />
          <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
            <WeatherKitForm
              key={`${provider?.id ?? 'new'}:${location?.id ?? 'none'}`}
              provider={provider}
              location={location}
            />
            <ScreenPreview reading={reading.data} location={location} />
          </div>
        </>
      )}
    </Page>
  )
}
