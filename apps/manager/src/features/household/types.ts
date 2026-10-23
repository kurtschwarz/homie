import type { Swatch } from '@homie/ui/lib/people'

export type ManagerPerson = {
  id: string
  name: string
  emoji: string
  swatch: Swatch
  isFamily: boolean
}

export type TagKind = 'PERSON' | 'INCLUDE'

export type TagMapping = {
  tag: string
  label: string
  kind: TagKind
  person: ManagerPerson | null
}

export type Household = {
  people: ManagerPerson[]
  tagMappings: TagMapping[]
  calendarTags: string[]
}

export type PersonInput = {
  id?: string
  name: string
  emoji: string
  swatch: string
  isFamily: boolean
}

export type TagMappingInput = { tag: string; kind: TagKind; personId?: string }
