import type { DisplayModeKey, DisplayModeSetting, DisplaySchedule } from '../types'
import { MODE_ORDER, MODES } from './modes'

export const DAY = 24 * 60
const STEP = 30

export type Segment = { mode: DisplayModeKey; from: number; to: number }

export function formatTime(minutes: number) {
  const hour24 = Math.floor(minutes / 60) % 24
  const minute = minutes % 60
  const hour = hour24 % 12 === 0 ? 12 : hour24 % 12
  return `${hour}:${String(minute).padStart(2, '0')} ${hour24 < 12 ? 'AM' : 'PM'}`
}

export function orderModes(modes: DisplayModeSetting[]) {
  return MODE_ORDER.flatMap((mode) => modes.filter((setting) => setting.mode === mode))
}

/** Start times a row may take: strictly between its neighbours, so the day stays in order. */
export function startOptions(modes: DisplayModeSetting[], index: number) {
  const previous = modes[index - 1]
  const next = modes[index + 1]
  const current = modes[index]?.startsAtMinutes
  const min = previous ? previous.startsAtMinutes + STEP : 0
  const max = next ? next.startsAtMinutes - STEP : DAY - STEP
  const options: number[] = []
  for (let minutes = min; minutes <= max; minutes += STEP) options.push(minutes)
  if (current !== undefined && !options.includes(current)) options.push(current)
  return options.sort((a, b) => a - b)
}

/** Power saving off drops OFF from the day, so the last mode then runs through midnight. */
function activeModes(schedule: DisplaySchedule) {
  const modes = orderModes(schedule.modes)
  return schedule.powerSavingEnabled ? modes : modes.filter((setting) => setting.mode !== 'OFF')
}

/** Bands across 0-1440. The first and last bands are the same mode because the day wraps. */
export function segments(schedule: DisplaySchedule): Segment[] {
  const modes = activeModes(schedule)
  const first = modes[0]
  const last = modes[modes.length - 1]
  if (!first || !last) return []
  const result: Segment[] = []
  if (first.startsAtMinutes > 0) {
    result.push({ mode: last.mode, from: 0, to: first.startsAtMinutes })
  }
  modes.forEach((setting, index) => {
    const to = modes[index + 1]?.startsAtMinutes ?? DAY
    result.push({ mode: setting.mode, from: setting.startsAtMinutes, to })
  })
  return result
}

export function modeAt(schedule: DisplaySchedule, minutes: number): DisplayModeKey {
  const found = segments(schedule).find(
    (segment) => minutes >= segment.from && minutes < segment.to,
  )
  return found?.mode ?? 'LIGHT'
}

export function brightnessOf(schedule: DisplaySchedule, mode: DisplayModeKey) {
  return schedule.modes.find((setting) => setting.mode === mode)?.brightness ?? 100
}

function startOf(schedule: DisplaySchedule, mode: DisplayModeKey) {
  return schedule.modes.find((setting) => setting.mode === mode)?.startsAtMinutes ?? 0
}

export function overnightSummary(schedule: DisplaySchedule) {
  const light = formatTime(startOf(schedule, 'LIGHT'))
  if (!schedule.powerSavingEnabled) {
    const dark = formatTime(startOf(schedule, 'NIGHT'))
    return `The display never switches off. ${MODES.NIGHT.label} stays on from ${dark} until ${MODES.LIGHT.label} starts at ${light}.`
  }
  const off = formatTime(startOf(schedule, 'OFF'))
  const wake = schedule.wakeForAlerts
    ? ' It wakes for alerts and when you tap it.'
    : ' It wakes when you tap it.'
  return `The display is off from ${off} until ${MODES.LIGHT.label} starts at ${light}.${wake}`
}
