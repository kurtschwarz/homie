'use client'

import { useNow } from './use-now'

export type HomieTheme = 'light' | 'dark' | 'night'
export type HomieThemeSetting = HomieTheme | 'auto'

/** A schedule entry: the theme (or "off", which keeps the night palette) that starts at a minute of the day. */
export type HomieThemeScheduleEntry = {
  mode: HomieTheme | 'off'
  startsAtMinutes: number
}

export type HomieThemeSchedule = HomieThemeScheduleEntry[]

/** Auto schedule: Day until 7 pm, Evening from 7 pm, Night from 9 pm until 6:30 am. */
export const DEFAULT_THEME_SCHEDULE: HomieThemeSchedule = [
  { mode: 'light', startsAtMinutes: 390 },
  { mode: 'dark', startsAtMinutes: 1140 },
  { mode: 'night', startsAtMinutes: 1260 },
]

function themeFor(mode: HomieTheme | 'off'): HomieTheme {
  return mode === 'off' ? 'night' : mode
}

export function themeForTime(
  d: Date,
  schedule: HomieThemeSchedule = DEFAULT_THEME_SCHEDULE,
): HomieTheme {
  if (schedule.length === 0) return themeForTime(d, DEFAULT_THEME_SCHEDULE)
  const sorted = [...schedule].sort((a, b) => a.startsAtMinutes - b.startsAtMinutes)
  const minutes = d.getHours() * 60 + d.getMinutes()
  let current = sorted[sorted.length - 1] as HomieThemeScheduleEntry
  for (const entry of sorted) {
    if (entry.startsAtMinutes <= minutes) current = entry
  }
  return themeFor(current.mode)
}

/**
 * Resolves a theme setting to the theme to show now. "auto" follows the clock through the schedule, unless an
 * override is live.
 */
export function useHomieTheme(
  setting: HomieThemeSetting = 'auto',
  schedule?: HomieThemeSchedule,
  override?: HomieTheme | null,
): HomieTheme {
  const now = useNow(60_000)
  if (setting !== 'auto') return setting
  return override ?? themeForTime(now, schedule)
}
