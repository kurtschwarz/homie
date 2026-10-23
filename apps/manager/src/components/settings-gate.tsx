import type { ReactNode } from 'react'
import { useAppData } from '../app-data'
import type { Settings } from '../lib/use-settings'
import { ErrorMessage } from './error-message'

/**
 * Renders its child only once settings have loaded. A failed save - or a
 * failed refetch after one - keeps the last good settings on screen and shows
 * the error above them, so an unsaved edit is never thrown away by a flash of
 * the error state. Only a first load with nothing cached gives the error the
 * whole page.
 */
export function SettingsGate({ children }: { children: (settings: Settings) => ReactNode }) {
  const { settings, isPending, error, saveError } = useAppData()

  if (isPending) return <p className="text-hm-body text-hm-ink-2">Loading settings…</p>
  if (!settings) return <ErrorMessage message={error?.message ?? 'Settings unavailable'} />

  return (
    <>
      {error && <ErrorMessage message={error.message} />}
      {saveError && <ErrorMessage message={saveError.message} />}
      {children(settings)}
    </>
  )
}
