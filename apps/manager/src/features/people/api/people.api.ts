import { graphql } from '@/shared/lib/graphql'
import type { ManagerPerson, PersonInput } from '../types'

const PEOPLE_QUERY = `
  query ManagerPeople {
    people { id name emoji swatch isFamily }
  }
`

export const peopleApi = {
  list: async (): Promise<ManagerPerson[]> => {
    const data = await graphql<{ people: ManagerPerson[] }>(PEOPLE_QUERY)
    return data.people
  },

  save: async (input: PersonInput): Promise<string> => {
    const data = await graphql<{ savePerson: { id: string } }>(
      `mutation SavePerson($input: PersonInput!) { savePerson(input: $input) { id } }`,
      { input },
    )
    return data.savePerson.id
  },

  remove: (id: string) =>
    graphql(`mutation DeletePerson($id: ID!) { deletePerson(id: $id) }`, { id }),
}
