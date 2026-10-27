import { HomieThemeProvider } from '@homie/ui/components/theme-provider'
import type { ReactNode } from 'react'
import { useDisplayOverride, useDisplaySchedule } from './use-display-theme'

/** Applies the theme from the fetched schedule, falling back to the built-in one while the API is unreachable. */
export function DisplayTheme({ children }: { children: ReactNode }) {
  const schedule = useDisplaySchedule()
  const override = useDisplayOverride()
  return (
    <HomieThemeProvider
      theme="auto"
      surface="kiosk"
      schedule={schedule.data}
      override={override.data}
    >
      {children}
    </HomieThemeProvider>
  )
}
