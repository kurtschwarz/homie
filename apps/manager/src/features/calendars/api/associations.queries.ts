import { queryOptions } from '@tanstack/react-query'
import { associationsApi } from './associations.api'

export const associationKeys = {
  all: ['associations'] as const,
  person: (personId: string) => [...associationKeys.all, 'person', personId] as const,
}

export const associationQueries = {
  /** Every calendar one person is associated with, and how. */
  forPerson: (personId: string) =>
    queryOptions({
      queryKey: associationKeys.person(personId),
      queryFn: () => associationsApi.forPerson(personId),
    }),
}
