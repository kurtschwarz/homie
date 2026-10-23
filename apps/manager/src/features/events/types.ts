import type { Swatch } from '@homie/ui/lib/people'

export type ManagerEvent = {
  id: string
  title: string
  start: Date
  end: Date
  allDay: boolean
  people: { id: string; name: string; emoji: string; swatch: Swatch }[]
}
