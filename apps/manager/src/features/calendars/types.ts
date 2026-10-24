export type CalendarConnection = {
  url: string
  username: string
  timezone: string
}

export type CalendarAccount = {
  id: string
  name: string
  provider: 'CALDAV'
  connection: CalendarConnection
  refreshIntervalSeconds: number
  windowPastSeconds: number
  windowFutureSeconds: number
}

export type Calendar = {
  id: string
  accountId: string
  providerRef: string
  name: string
  readOnly: boolean
  visible: boolean
  lastSyncedAt: number | null
  lastError: string
  eventCount: number
}

export type CalendarAccountDraft = {
  name: string
  url: string
  username: string
  password: string
  timezone: string
}
