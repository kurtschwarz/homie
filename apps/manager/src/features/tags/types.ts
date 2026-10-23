import type { ManagerPerson } from '@/features/people'

export type TagKind = 'PERSON' | 'INCLUDE'

export type TagMapping = {
  tag: string
  label: string
  kind: TagKind
  person: ManagerPerson | null
}

export type TagMappingInput = { tag: string; kind: TagKind; personId?: string }
