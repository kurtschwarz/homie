import { useMutation, useQueryClient } from '@tanstack/react-query'
import type {
  WeatherLocation,
  WeatherLocationDraft,
  WeatherProvider,
  WeatherProviderDraft,
} from '../types'
import { weatherApi } from './weather.api'
import { weatherKeys } from './weather.queries'

type SaveWeather = {
  provider: WeatherProvider | undefined
  location: WeatherLocation | undefined
  draft: WeatherProviderDraft
  locationDraft: WeatherLocationDraft
  refreshIntervalSeconds: number
}

/** Saves the provider, then the location. Refetches either way so a half-finished save shows. */
export function useSaveWeather() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({
      provider,
      location,
      draft,
      locationDraft,
      refreshIntervalSeconds,
    }: SaveWeather) => {
      if (provider) {
        await weatherApi.updateProvider({ id: provider.id, draft, refreshIntervalSeconds })
      } else {
        await weatherApi.createProvider({ draft, refreshIntervalSeconds })
      }
      if (location) await weatherApi.updateLocation({ id: location.id, draft: locationDraft })
      else await weatherApi.createLocation(locationDraft)
    },
    onSettled: () => queryClient.invalidateQueries({ queryKey: weatherKeys.all }),
  })
}

export function useRemoveWeatherProvider() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: weatherApi.removeProvider,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: weatherKeys.all }),
  })
}

export function useTestDraftWeatherProvider() {
  return useMutation({ mutationFn: weatherApi.testDraft })
}

export function useTestStoredWeatherProvider() {
  return useMutation({ mutationFn: weatherApi.testStored })
}
