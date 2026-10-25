export type WeatherKitConnection = {
  teamId: string
  serviceId: string
  keyId: string
  hasPrivateKey: boolean
}

export type WeatherProvider = {
  id: string
  kind: 'WEATHERKIT'
  active: boolean
  connection: WeatherKitConnection
  refreshIntervalSeconds: number
  lastAttemptedAt: number | null
  lastError: string
}

export type WeatherLocation = {
  id: string
  label: string
  latitude: number
  longitude: number
  order: number
}

export type WeatherCondition = 'SUN' | 'CLOUD_SUN' | 'CLOUD' | 'RAIN' | 'MOON'

/** The current reading on the kitchen screen. Absent until the first refresh. */
export type WeatherReading = {
  condition: WeatherCondition
  conditionCode: string
  temperature: number
  fetchedAt: number
  stale: boolean
}

/** The WeatherKit form. `privateKey` is the PEM text of a .p8 and is write-only. */
export type WeatherProviderDraft = {
  teamId: string
  serviceId: string
  keyId: string
  privateKey: string
}

export type WeatherLocationDraft = {
  label: string
  latitude: string
  longitude: string
}
