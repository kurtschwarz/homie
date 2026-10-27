import type { HomieTheme, HomieThemeSchedule } from '@homie/ui/hooks/use-homie-theme'
import { useQuery } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'
import { pollEvery } from '../../lib/query-client'

type Mode = 'LIGHT' | 'DARK' | 'NIGHT' | 'OFF'

const SCHEDULE_QUERY = `
  query DisplaySchedule {
    displaySchedule {
      modes { mode startsAtMinutes }
    }
  }
`

const STATUS_QUERY = `
  query DisplayStatus {
    displayStatus {
      mode
      overrideActive
    }
  }
`

function toTheme(mode: Mode): HomieTheme {
  return mode === 'LIGHT' ? 'light' : mode === 'DARK' ? 'dark' : 'night'
}

/** Fetches the theme schedule every few minutes; the clock still picks the current theme client-side. */
export function useDisplaySchedule() {
  return useQuery({
    queryKey: ['display-schedule'],
    queryFn: async (): Promise<HomieThemeSchedule> => {
      const data = await graphql<{
        displaySchedule: { modes: { mode: Mode; startsAtMinutes: number }[] }
      }>(SCHEDULE_QUERY)
      return data.displaySchedule.modes.map((entry) => ({
        mode: entry.mode === 'OFF' ? 'off' : toTheme(entry.mode),
        startsAtMinutes: entry.startsAtMinutes,
      }))
    },
    refetchInterval: pollEvery(5 * 60_000),
  })
}

/** Polls for a live preview on a short interval; resolves to the previewed theme or null. */
export function useDisplayOverride() {
  return useQuery({
    queryKey: ['display-override'],
    queryFn: async (): Promise<HomieTheme | null> => {
      const data = await graphql<{ displayStatus: { mode: Mode; overrideActive: boolean } }>(
        STATUS_QUERY,
      )
      return data.displayStatus.overrideActive ? toTheme(data.displayStatus.mode) : null
    },
    refetchInterval: pollEvery(2_000),
  })
}
