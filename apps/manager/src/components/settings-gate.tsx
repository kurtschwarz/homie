import type { ReactNode } from 'react'
import { useAppData } from '../app-data'
import type { Settings } from '../lib/use-settings'
import { ErrorMessage } from './error-message'

/**
 * Renders its child only once settings have loaded. A failed save keeps the
 * last good settings on screen and shows the error above them, so an
 * unsaved edit is never thrown away by a flash of the error state.
 */
export function SettingsGate({ children }: { children: (settings: Settings) => ReactNode }) {
  const { state } = useAppData()

  if (state.kind === 'loading')
    return <p className="text-hm-body text-hm-ink-2">Loading settings…</p>
  if (state.kind === 'error') return <ErrorMessage message={state.message} />

  return (
    <>
      {state.error && <ErrorMessage message={state.error} />}
      {children(state.settings)}
    </>
  )
}
