import { useEffect, useState } from 'react'
import { graphql } from '../../lib/graphql'

type Weather = {
  condition: 'SUN' | 'CLOUD_SUN' | 'CLOUD' | 'RAIN' | 'MOON'
  conditionCode: string
  temperature: number
  fetchedAt: number
  stale: boolean
}

type WeatherState = { kind: 'loading' } | { kind: 'ready'; weather: Weather } | { kind: 'error' }

const QUERY = `
  query Weather {
    weather {
      condition
      conditionCode
      temperature
      fetchedAt
      stale
    }
  }
`

/** Fetches current conditions through GraphQL and refreshes on a ten-minute cadence. */
export function useWeather(): WeatherState {
  const [state, setState] = useState<WeatherState>({ kind: 'loading' })

  useEffect(() => {
    let cancelled = false
    let timeout: ReturnType<typeof setTimeout> | undefined

    async function fetchWeather() {
      let retryDelay = 10 * 60_000
      try {
        const data = await graphql<{ weather?: Weather }>(QUERY)
        if (!data.weather) throw new Error('Weather query failed')
        if (!cancelled) setState({ kind: 'ready', weather: data.weather })
      } catch {
        retryDelay = 30_000
        if (!cancelled) setState({ kind: 'error' })
      } finally {
        if (!cancelled) timeout = setTimeout(fetchWeather, retryDelay)
      }
    }

    void fetchWeather()
    return () => {
      cancelled = true
      if (timeout) clearTimeout(timeout)
    }
  }, [])

  return state
}
