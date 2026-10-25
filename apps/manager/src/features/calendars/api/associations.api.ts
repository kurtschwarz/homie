import { graphql } from '@/shared/lib/graphql'
import type { AssociationScope, CalendarAssociation } from '../types'

const ASSOCIATION_FIELDS = `calendarId person { id } scope terms`

type AssociationPayload = {
  calendarId: string
  person: { id: string }
  scope: AssociationScope
  terms: string[]
}

function toAssociation(payload: AssociationPayload): CalendarAssociation {
  return {
    calendarId: payload.calendarId,
    personId: payload.person.id,
    scope: payload.scope,
    terms: payload.terms,
  }
}

export type CreateAssociationInput = {
  calendarId: string
  personId: string
  scope?: AssociationScope
  terms?: string[]
}

export type UpdateAssociationInput = {
  calendarId: string
  personId: string
  scope?: AssociationScope
  terms?: string[]
}

export const associationsApi = {
  forPerson: async (personId: string): Promise<CalendarAssociation[]> => {
    const data = await graphql<{ calendarAssociations: AssociationPayload[] }>(
      `query PersonAssociations($personId: ID!) {
        calendarAssociations(personId: $personId) { ${ASSOCIATION_FIELDS} }
      }`,
      { personId },
    )
    return data.calendarAssociations.map(toAssociation)
  },

  create: async (input: CreateAssociationInput): Promise<CalendarAssociation> => {
    const data = await graphql<{ createCalendarAssociation: AssociationPayload }>(
      `mutation CreateCalendarAssociation($input: CreateCalendarAssociationInput!) {
        createCalendarAssociation(input: $input) { ${ASSOCIATION_FIELDS} }
      }`,
      { input },
    )
    return toAssociation(data.createCalendarAssociation)
  },

  update: async (input: UpdateAssociationInput): Promise<CalendarAssociation> => {
    const data = await graphql<{ updateCalendarAssociation: AssociationPayload }>(
      `mutation UpdateCalendarAssociation($input: UpdateCalendarAssociationInput!) {
        updateCalendarAssociation(input: $input) { ${ASSOCIATION_FIELDS} }
      }`,
      { input },
    )
    return toAssociation(data.updateCalendarAssociation)
  },

  remove: (input: { calendarId: string; personId: string }) =>
    graphql(
      `mutation DeleteCalendarAssociation($calendarId: ID!, $personId: ID!) {
        deleteCalendarAssociation(calendarId: $calendarId, personId: $personId)
      }`,
      input,
    ),
}
