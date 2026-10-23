import type { Swatch } from '@homie/ui/lib/people'
import { useCallback, useEffect, useRef, useState } from 'react'
import { graphql } from './graphql'

export type ManagerEvent = {
  id: string
  title: string
  start: Date
  end: Date
  allDay: boolean
  people: { id: string; name: string; emoji: string; swatch: Swatch }[]
}

type EventRecord = {
  id: string
  title: string
  start: number
  end: number
  allDay: boolean
  people: { id: string; name: string; emoji: string; swatch: string }[]
}

const QUERY = `
  query ManagerEvents {
    events(days: 7) {
      id
      title
      start
      end
      allDay
      people { id name emoji swatch }
    }
  }
`

/**
 * The coming week, already attributed to people by the API. The manager only
 * reads this to show someone what their badge is actually doing — it never
 * writes here — so a failed fetch reports `failed` and the page drops the
 * week strip rather than claiming nobody has any events.
 *
 * `events(days:)` counts forward from now, so this is the next seven days
 * rather than Monday to Sunday: earlier days of the current week are not
 * available to ask for. The hook refreshes every five minutes and exposes
 * `refresh` so a settings mutation can update attribution immediately.
 */
export function useEvents() {
  const [events, setEvents] = useState<ManagerEvent[]>([])
  const [failed, setFailed] = useState(false)
  const request = useRef(0)

  const refresh = useCallback(async () => {
    const current = ++request.current
    try {
      const data = await graphql<{ events: EventRecord[] }>(QUERY)
      if (current !== request.current) return
      setEvents(
        data.events.map((record) => ({
          id: record.id,
          title: record.title,
          start: new Date(record.start * 1000),
          end: new Date(record.end * 1000),
          allDay: record.allDay,
          people: record.people.map((person) => ({
            ...person,
            swatch: person.swatch as Swatch,
          })),
        })),
      )
      setFailed(false)
    } catch {
      if (current === request.current) setFailed(true)
    }
  }, [])

  useEffect(() => {
    void refresh()
    const timer = setInterval(refresh, 5 * 60_000)
    return () => {
      request.current++
      clearInterval(timer)
    }
  }, [refresh])

  return { events, failed, refresh }
}

/**
 * The seven calendar days starting with the day containing `now`, each at
 * local midnight. Days are stepped with `setDate` rather than by adding 24
 * hours, so a daylight-saving change cannot skip or repeat a date.
 *
 * @param now - The moment that counts as today. Defaults to the current time;
 * pass a fixed value to get a stable result.
 */
export function weekDays(now = new Date()) {
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  return Array.from({ length: 7 }, (_, offset) => {
    const day = new Date(start)
    day.setDate(start.getDate() + offset)
    return day
  })
}

/**
 * The events attributed to one person. An event shared by several people is
 * included for each of them.
 *
 * @param events - Events as returned by `useEvents`.
 * @param personId - The id of the person to keep events for.
 */
export function eventsFor(events: ManagerEvent[], personId: string) {
  return events.filter((event) => event.people.some((person) => person.id === personId))
}

/**
 * The events that start on a given local calendar day. Only the start counts:
 * an event that runs past midnight or spans several days appears on the day
 * it begins and not on the days after.
 *
 * @param events - Events to filter.
 * @param day - Any moment within the day to match; the time of day is ignored.
 */
export function onDay(events: ManagerEvent[], day: Date) {
  return events.filter(
    (event) =>
      event.start.getFullYear() === day.getFullYear() &&
      event.start.getMonth() === day.getMonth() &&
      event.start.getDate() === day.getDate(),
  )
}
