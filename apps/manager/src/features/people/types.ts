import type { Swatch } from '@homie/ui/lib/people'

export type ManagerPerson = {
  id: string
  name: string
  emoji: string
  swatch: Swatch
  isFamily: boolean
}

export type PersonInput = {
  id?: string
  name: string
  emoji: string
  swatch: string
  isFamily: boolean
}
