export type DisplayModeKey = 'LIGHT' | 'DARK' | 'NIGHT' | 'OFF'

export type DisplayModeSetting = {
  mode: DisplayModeKey
  startsAtMinutes: number
  brightness: number
}

export type DisplaySchedule = {
  modes: DisplayModeSetting[]
  powerSavingEnabled: boolean
  wakeForAlerts: boolean
  ddcBus: string
}

export type DisplayStatus = {
  mode: DisplayModeKey
  brightness: number
  powered: boolean
  overrideActive: boolean
  overrideEndsAt: number | null
}

export type WifiStatus = {
  connected: boolean
  ssid: string
  signalPercent: number
}
