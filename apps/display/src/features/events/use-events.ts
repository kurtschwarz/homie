import type { Swatch } from '@homie/ui/lib/people'
import { useQuery } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'
import { pollEvery } from '../../lib/query-client'
import type { CalendarEvent } from './calendar-event'

type EventRecord = {
  id: string
  title: string
  location: string | null
  start: number
  end: number
  allDay: boolean
  people: { id: string; name: string; emoji: string; swatch: string }[]
}

const QUERY = `
  query Events {
    events(days: 21) {
      id
      title
      location
      start
      end
      allDay
      people { id name emoji swatch }
    }
  }
`

function toCalendarEvent(record: EventRecord): CalendarEvent {
  return {
    id: record.id,
    title: record.title,
    location: record.location ?? undefined,
    start: new Date(record.start * 1000),
    end: new Date(record.end * 1000),
    allDay: record.allDay,
    people: record.people.map((p) => ({
      id: p.id,
      name: p.name,
      emoji: p.emoji,
      swatch: p.swatch as Swatch,
    })),
  }
}

function select(data: { events?: EventRecord[] }): CalendarEvent[] {
  if (!data.events) throw new Error('Events query failed')
  return data.events.map(toCalendarEvent)
}

async function fetchEvents() {
  return graphql<{ events?: EventRecord[] }>(QUERY)
}

/**
 * Fetches the family schedule through GraphQL and refreshes every 5 minutes —
 * shorter than the server's 15-minute sync, because the derived view (what's
 * Now vs Next) decays even without new data. A failed refetch keeps the last
 * good schedule in the cache, which TanStack retains automatically.
 */
export function useEvents() {
  return useQuery({
    queryKey: ['events'],
    queryFn: fetchEvents,
    select,
    refetchInterval: pollEvery(5 * 60_000),
  })
}
