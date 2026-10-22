import { useQuery } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'
import { pollEvery } from '../../lib/query-client'

type Weather = {
  condition: 'SUN' | 'CLOUD_SUN' | 'CLOUD' | 'RAIN' | 'MOON'
  conditionCode: string
  temperature: number
  fetchedAt: number
  stale: boolean
}

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

async function fetchWeather() {
  const data = await graphql<{ weather?: Weather }>(QUERY)
  if (!data.weather) throw new Error('Weather query failed')
  return data.weather
}

/** Fetches current conditions through GraphQL and refreshes on a ten-minute cadence. */
export function useWeather() {
  return useQuery({
    queryKey: ['weather'],
    queryFn: fetchWeather,
    refetchInterval: pollEvery(10 * 60_000),
  })
}
