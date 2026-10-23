import * as React from 'react'
import { useSettings } from '@/lib/use-settings'

type AppData = ReturnType<typeof useSettings>

const AppDataContext = React.createContext<AppData | null>(null)

/**
 * Settings, held once for the whole shell. The query itself would be shared
 * without this - the cache dedupes by key - but the save lane would not: a
 * `useSettings` call per page means a `saving` flag and a `saveError` per
 * page, and the banner would miss a failure started elsewhere. Holding the
 * hook once keeps one lane for every page to read.
 */
export function AppDataProvider({ children }: { children: React.ReactNode }) {
  const settings = useSettings()

  const value = React.useMemo(() => ({ ...settings }), [settings])
  return <AppDataContext.Provider value={value}>{children}</AppDataContext.Provider>
}

export function useAppData() {
  const value = React.useContext(AppDataContext)
  if (!value) throw new Error('App data context is missing')
  return value
}
