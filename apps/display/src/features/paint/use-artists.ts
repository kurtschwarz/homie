import type { Person, Swatch } from '@homie/ui/lib/people'
import { useQuery } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'

type PersonRecord = { id: string; name: string; emoji: string; swatch: string; isFamily: boolean }

const QUERY = `
  query Artists {
    people { id name emoji swatch isFamily }
  }
`

/** People who can be credited with a drawing: everyone except the family pseudo-person. */
export function useArtists() {
  return useQuery({
    queryKey: ['artists'],
    queryFn: () => graphql<{ people?: PersonRecord[] }>(QUERY),
    select: (data): Person[] =>
      (data.people ?? [])
        .filter((person) => !person.isFamily)
        .map((person) => ({
          id: person.id,
          name: person.name,
          emoji: person.emoji,
          swatch: person.swatch as Swatch,
        })),
  })
}
