import * as React from 'react'
import { type HealthState, useHealth } from './lib/use-health'
import type { User } from './lib/use-session'
import { useSettings } from './lib/use-settings'

type AppData = ReturnType<typeof useSettings> & {
  health: HealthState
  user: User | null
  logout: () => Promise<void>
}

const AppDataContext = React.createContext<AppData | null>(null)

/**
 * Settings, health and the logged-in user, held once for the whole shell
 * so the sidebar badge and every page read the same copy - and so health
 * is polled once rather than per consumer. user and logout come from
 * AuthGate, which resolved them before this ever mounted. user is null
 * when the API isn't gating at all.
 */
export function AppDataProvider({
  user,
  logout,
  children,
}: {
  user: User | null
  logout: () => Promise<void>
  children: React.ReactNode
}) {
  const settings = useSettings()
  const health = useHealth()

  const value = React.useMemo(
    () => ({ ...settings, health, user, logout }),
    [settings, health, user, logout],
  )
  return <AppDataContext.Provider value={value}>{children}</AppDataContext.Provider>
}

export function useAppData() {
  const value = React.useContext(AppDataContext)
  if (!value) throw new Error('App data context is missing')
  return value
}
