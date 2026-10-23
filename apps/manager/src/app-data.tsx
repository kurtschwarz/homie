import * as React from 'react'
import type { User } from '@/lib/use-session'
import { useSettings } from '@/lib/use-settings'

type AppData = ReturnType<typeof useSettings> & {
  user: User | null
  logout: () => Promise<void>
}

const AppDataContext = React.createContext<AppData | null>(null)

/**
 * Settings and the logged-in user, held once for the whole shell. The query
 * itself would be shared without this - the cache dedupes by key - but the
 * save lane would not: a `useSettings` call per page means a `saving` flag and
 * a `saveError` per page, and the banner would miss a failure started
 * elsewhere. Holding the hook once keeps one lane for every page to read.
 *
 * user and logout come from AuthGate, which resolved them before this ever
 * mounted. user is null when the API isn't gating at all. Health is not here:
 * it is read-only, so consumers call useHealth directly and the cache does the
 * sharing.
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

  const value = React.useMemo(() => ({ ...settings, user, logout }), [settings, user, logout])
  return <AppDataContext.Provider value={value}>{children}</AppDataContext.Provider>
}

export function useAppData() {
  const value = React.useContext(AppDataContext)
  if (!value) throw new Error('App data context is missing')
  return value
}
