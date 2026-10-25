import { graphql } from '@/shared/lib/graphql'
import type {
  WeatherLocation,
  WeatherLocationDraft,
  WeatherProvider,
  WeatherProviderDraft,
  WeatherReading,
} from '../types'

const PROVIDER_FIELDS = `
  id kind active refreshIntervalSeconds lastAttemptedAt lastError
  connection { ... on WeatherKitConnection { teamId serviceId keyId hasPrivateKey } }
`

const LOCATION_FIELDS = `id label latitude longitude order`

const WEATHER_QUERY = `
  query ManageWeather {
    weatherProviders { ${PROVIDER_FIELDS} }
    weatherLocations { ${LOCATION_FIELDS} }
  }
`

function connection(draft: WeatherProviderDraft) {
  return {
    weatherkit: {
      teamId: draft.teamId,
      serviceId: draft.serviceId,
      keyId: draft.keyId,
      privateKey: draft.privateKey,
    },
  }
}

function coordinates(draft: WeatherLocationDraft) {
  return {
    label: draft.label.trim(),
    latitude: Number(draft.latitude),
    longitude: Number(draft.longitude),
  }
}

export const weatherApi = {
  list: () =>
    graphql<{ weatherProviders: WeatherProvider[]; weatherLocations: WeatherLocation[] }>(
      WEATHER_QUERY,
    ),

  /** Null while there is no reading yet, which the API reports as an error. */
  reading: async (): Promise<WeatherReading | null> => {
    try {
      const data = await graphql<{ weather: WeatherReading }>(
        `query ManageWeatherReading {
          weather { condition conditionCode temperature fetchedAt stale }
        }`,
      )
      return data.weather
    } catch {
      return null
    }
  },

  testDraft: (draft: WeatherProviderDraft) =>
    graphql<{ testWeatherProvider: boolean }>(
      `mutation TestWeatherProvider($input: TestWeatherProviderInput!) {
        testWeatherProvider(input: $input)
      }`,
      { input: { connection: connection(draft) } },
    ),

  testStored: (id: string) =>
    graphql<{ testWeatherProvider: boolean }>(
      `mutation TestStoredWeatherProvider($input: TestWeatherProviderInput!) {
        testWeatherProvider(input: $input)
      }`,
      { input: { id } },
    ),

  createProvider: async ({
    draft,
    refreshIntervalSeconds,
  }: {
    draft: WeatherProviderDraft
    refreshIntervalSeconds: number
  }) => {
    const data = await graphql<{ createWeatherProvider: WeatherProvider }>(
      `mutation CreateWeatherProvider($input: CreateWeatherProviderInput!) {
        createWeatherProvider(input: $input) { ${PROVIDER_FIELDS} }
      }`,
      { input: { connection: connection(draft), refreshIntervalSeconds } },
    )
    return data.createWeatherProvider
  },

  updateProvider: async ({
    id,
    draft,
    refreshIntervalSeconds,
  }: {
    id: string
    draft: WeatherProviderDraft
    refreshIntervalSeconds: number
  }) => {
    const weatherkit: Record<string, string> = {
      teamId: draft.teamId,
      serviceId: draft.serviceId,
      keyId: draft.keyId,
    }
    if (draft.privateKey) weatherkit.privateKey = draft.privateKey
    const data = await graphql<{ updateWeatherProvider: WeatherProvider }>(
      `mutation UpdateWeatherProvider($input: UpdateWeatherProviderInput!) {
        updateWeatherProvider(input: $input) { ${PROVIDER_FIELDS} }
      }`,
      { input: { id, connection: { weatherkit }, refreshIntervalSeconds } },
    )
    return data.updateWeatherProvider
  },

  removeProvider: (id: string) =>
    graphql(`mutation DeleteWeatherProvider($id: ID!) { deleteWeatherProvider(id: $id) }`, { id }),

  createLocation: async (draft: WeatherLocationDraft) => {
    const data = await graphql<{ createWeatherLocation: WeatherLocation }>(
      `mutation CreateWeatherLocation($input: CreateWeatherLocationInput!) {
        createWeatherLocation(input: $input) { ${LOCATION_FIELDS} }
      }`,
      { input: coordinates(draft) },
    )
    return data.createWeatherLocation
  },

  updateLocation: async ({ id, draft }: { id: string; draft: WeatherLocationDraft }) => {
    const data = await graphql<{ updateWeatherLocation: WeatherLocation }>(
      `mutation UpdateWeatherLocation($input: UpdateWeatherLocationInput!) {
        updateWeatherLocation(input: $input) { ${LOCATION_FIELDS} }
      }`,
      { input: { id, ...coordinates(draft) } },
    )
    return data.updateWeatherLocation
  },
}
