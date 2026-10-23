import type { Swatch } from '@homie/ui/lib/people'
import { graphql } from '@/shared/lib/graphql'
import type { ManagerEvent } from '../types'

/** The wire shape: seconds since the epoch, and swatch as a bare string. */
type EventRecord = {
  id: string
  title: string
  start: number
  end: number
  allDay: boolean
  people: { id: string; name: string; emoji: string; swatch: string }[]
}

const EVENTS_QUERY = `
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

function toManagerEvent(record: EventRecord): ManagerEvent {
  return {
    id: record.id,
    title: record.title,
    start: new Date(record.start * 1000),
    end: new Date(record.end * 1000),
    allDay: record.allDay,
    people: record.people.map((person) => ({
      ...person,
      swatch: person.swatch as Swatch,
    })),
  }
}

export const eventsApi = {
  /**
   * `events(days:)` counts forward from now, so this is the next seven days
   * rather than Monday to Sunday: earlier days of the current week are not
   * available to ask for.
   */
  list: async (): Promise<ManagerEvent[]> => {
    const data = await graphql<{ events: EventRecord[] }>(EVENTS_QUERY)
    return data.events.map(toManagerEvent)
  },
}
