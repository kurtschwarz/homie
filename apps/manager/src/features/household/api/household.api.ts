import { graphql } from '@/shared/lib/graphql'
import type { Household, PersonInput, TagMappingInput } from '../types'

const HOUSEHOLD_QUERY = `
  query Settings {
    people { id name emoji swatch isFamily }
    tagMappings { tag label kind person { id name emoji swatch isFamily } }
    calendarTags
  }
`

export const householdApi = {
  get: () => graphql<Household>(HOUSEHOLD_QUERY),

  savePerson: async (input: PersonInput): Promise<string> => {
    const data = await graphql<{ savePerson: { id: string } }>(
      `mutation SavePerson($input: PersonInput!) { savePerson(input: $input) { id } }`,
      { input },
    )
    return data.savePerson.id
  },

  deletePerson: (id: string) =>
    graphql(`mutation DeletePerson($id: ID!) { deletePerson(id: $id) }`, { id }),

  saveTagMapping: (input: TagMappingInput) =>
    graphql<{ saveTagMapping: { tag: string } }>(
      `mutation SaveTagMapping($input: TagMappingInput!) { saveTagMapping(input: $input) { tag } }`,
      { input },
    ),

  deleteTagMapping: (tag: string) =>
    graphql(`mutation DeleteTagMapping($tag: String!) { deleteTagMapping(tag: $tag) }`, { tag }),

  syncCalendar: () => graphql(`mutation RefreshEvents { refreshEvents }`),
}
