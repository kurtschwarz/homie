export type WeatherProviderPreset = {
  id: string
  name: string
  detail: string
  available: boolean
}

export const WEATHER_PROVIDERS: WeatherProviderPreset[] = [
  {
    id: 'weatherkit',
    name: 'Apple WeatherKit',
    detail: 'Accurate forecasts and rain alerts from Apple. Needs an Apple Developer key.',
    available: true,
  },
  {
    id: 'environment-canada',
    name: 'Environment Canada',
    detail: 'Free, official Canadian forecasts and weather warnings.',
    available: false,
  },
  {
    id: 'open-meteo',
    name: 'Open-Meteo',
    detail: 'Free and open, no account or key needed.',
    available: false,
  },
  {
    id: 'openweather',
    name: 'OpenWeather',
    detail: 'Global forecasts with a free API key.',
    available: false,
  },
]

export const REFRESH_OPTIONS = [
  { seconds: 900, label: '15 min' },
  { seconds: 1800, label: '30 min' },
  { seconds: 3600, label: '1 hr' },
]

export const CONDITION_LABELS = {
  SUN: 'Sunny',
  CLOUD_SUN: 'Partly cloudy',
  CLOUD: 'Cloudy',
  RAIN: 'Rain or snow',
  MOON: 'Clear night',
} as const
