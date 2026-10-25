import { Button } from '@homie/ui/components/button'
import * as React from 'react'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Field } from '@/shared/ui/field'
import { FormSection } from '@/shared/ui/form-section'
import {
  useRemoveWeatherProvider,
  useSaveWeather,
  useTestDraftWeatherProvider,
  useTestStoredWeatherProvider,
} from '../api/weather.mutations'
import { REFRESH_OPTIONS } from '../lib/providers'
import type {
  WeatherLocation,
  WeatherLocationDraft,
  WeatherProvider,
  WeatherProviderDraft,
} from '../types'
import { PrivateKeyField } from './private-key-field'

export function WeatherKitForm({
  provider,
  location,
}: {
  provider: WeatherProvider | undefined
  location: WeatherLocation | undefined
}) {
  const [draft, setDraft] = React.useState<WeatherProviderDraft>({
    teamId: provider?.connection.teamId ?? '',
    serviceId: provider?.connection.serviceId ?? '',
    keyId: provider?.connection.keyId ?? '',
    privateKey: '',
  })
  const [fileName, setFileName] = React.useState('')
  const [locationDraft, setLocationDraft] = React.useState<WeatherLocationDraft>({
    label: location?.label ?? '',
    latitude: location ? String(location.latitude) : '',
    longitude: location ? String(location.longitude) : '',
  })
  const [refresh, setRefresh] = React.useState(provider?.refreshIntervalSeconds ?? 1800)
  const [confirmRemove, setConfirmRemove] = React.useState(false)
  const [tested, setTested] = React.useState(false)
  const save = useSaveWeather()
  const remove = useRemoveWeatherProvider()
  const testDraft = useTestDraftWeatherProvider()
  const testStored = useTestStoredWeatherProvider()

  const identityChanged = Boolean(
    provider &&
      (draft.teamId !== provider.connection.teamId ||
        draft.serviceId !== provider.connection.serviceId ||
        draft.keyId !== provider.connection.keyId),
  )
  const needsKey = !provider?.connection.hasPrivateKey || identityChanged
  const latitude = Number(locationDraft.latitude)
  const longitude = Number(locationDraft.longitude)
  const locationValid =
    locationDraft.label.trim() !== '' &&
    locationDraft.latitude.trim() !== '' &&
    locationDraft.longitude.trim() !== '' &&
    Math.abs(latitude) <= 90 &&
    Math.abs(longitude) <= 180
  const connectionValid = Boolean(
    draft.teamId.trim() &&
      draft.serviceId.trim() &&
      draft.keyId.trim() &&
      (!needsKey || draft.privateKey),
  )

  function edit(patch: Partial<WeatherProviderDraft>) {
    setDraft((current) => ({ ...current, ...patch }))
    setTested(false)
  }

  function editLocation(patch: Partial<WeatherLocationDraft>) {
    setLocationDraft((current) => ({ ...current, ...patch }))
  }

  async function testConnection() {
    setTested(false)
    if (provider && !needsKey && !draft.privateKey) await testStored.mutateAsync(provider.id)
    else await testDraft.mutateAsync(draft)
    setTested(true)
  }

  async function saveAll() {
    await save.mutateAsync({
      provider,
      location,
      draft,
      locationDraft,
      refreshIntervalSeconds: refresh,
    })
    setDraft((current) => ({ ...current, privateKey: '' }))
    setFileName('')
  }

  const testPending = testDraft.isPending || testStored.isPending
  const error = save.error ?? remove.error ?? testDraft.error ?? testStored.error

  return (
    <section
      aria-labelledby="weatherkit-heading"
      className="flex flex-col gap-5 rounded-hm-16 border border-hm-line bg-hm-surface p-6"
    >
      <header>
        <h2 id="weatherkit-heading" className="m-0 text-hm-title font-semibold">
          WeatherKit
        </h2>
        <p className="m-0 text-hm-caption text-hm-ink-2">
          From your Apple Developer account, under Certificates, IDs &amp; Profiles
        </p>
      </header>

      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          id="weatherkit-team"
          label="Team ID"
          value={draft.teamId}
          onChange={(value) => edit({ teamId: value })}
        />
        <Field
          id="weatherkit-key-id"
          label="Key ID"
          value={draft.keyId}
          onChange={(value) => edit({ keyId: value })}
        />
      </div>
      <Field
        id="weatherkit-service"
        label="Service ID"
        placeholder="app.homie.weather"
        value={draft.serviceId}
        onChange={(value) => edit({ serviceId: value })}
      />
      <PrivateKeyField
        stored={Boolean(provider?.connection.hasPrivateKey)}
        fileName={fileName}
        onChange={(pem, name) => {
          edit({ privateKey: pem })
          setFileName(name)
        }}
      />

      <hr className="m-0 w-full border-0 border-t border-hm-line" />

      <div className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_auto]">
        <Field
          id="weather-location"
          label="Location"
          placeholder="Kitchener, ON"
          value={locationDraft.label}
          onChange={(value) => editLocation({ label: value })}
        />
        <FormSection label="Refresh every">
          <div className="flex gap-1 rounded-hm-control bg-hm-bg p-1">
            {REFRESH_OPTIONS.map((option) => (
              <button
                key={option.seconds}
                type="button"
                aria-pressed={refresh === option.seconds}
                onClick={() => setRefresh(option.seconds)}
                className={`h-9 cursor-pointer rounded-hm-control border-0 px-3 text-hm-caption font-semibold ${refresh === option.seconds ? 'bg-hm-surface text-hm-ink shadow-hm-raised' : 'bg-transparent text-hm-ink-2'}`}
              >
                {option.label}
              </button>
            ))}
          </div>
        </FormSection>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field
          id="weather-latitude"
          label="Latitude"
          placeholder="43.4422"
          value={locationDraft.latitude}
          onChange={(value) => editLocation({ latitude: value })}
        />
        <Field
          id="weather-longitude"
          label="Longitude"
          placeholder="-80.4699"
          value={locationDraft.longitude}
          onChange={(value) => editLocation({ longitude: value })}
        />
      </div>

      {error && <ErrorMessage label="Weather update failed" message={error.message} />}
      {tested && !error && (
        <p role="status" className="m-0 rounded-hm-16 bg-hm-accent/15 px-4 py-3 text-hm-body">
          <strong>WeatherKit accepted these credentials.</strong> Save to start fetching weather.
        </p>
      )}
      {provider?.lastError && (
        <ErrorMessage label="Last refresh failed" message={provider.lastError} />
      )}

      <div className="flex flex-wrap items-center gap-3">
        <Button
          variant="outline"
          className="h-11 border px-4 text-hm-caption"
          disabled={!connectionValid || testPending}
          onClick={() => void testConnection().catch(() => {})}
        >
          {testPending ? 'Testing…' : 'Test connection'}
        </Button>
        <Button
          variant="solid"
          disabled={!connectionValid || !locationValid || save.isPending}
          onClick={() => void saveAll().catch(() => {})}
        >
          {save.isPending ? 'Saving…' : 'Save'}
        </Button>
        {provider &&
          (confirmRemove ? (
            <button
              type="button"
              className="ml-auto cursor-pointer rounded-lg border-0 bg-red-50 px-3 py-2 text-hm-caption font-semibold text-red-700"
              onClick={() => remove.mutate(provider.id)}
            >
              Confirm remove
            </button>
          ) : (
            <button
              type="button"
              className="ml-auto cursor-pointer rounded-lg border-0 bg-transparent px-3 py-2 text-hm-caption font-semibold text-red-700 hover:bg-red-50"
              onClick={() => setConfirmRemove(true)}
            >
              Remove provider
            </button>
          ))}
      </div>
    </section>
  )
}
