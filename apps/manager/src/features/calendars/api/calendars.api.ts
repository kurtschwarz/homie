import { graphql } from '@/shared/lib/graphql'

const CALENDAR_TAGS_QUERY = `
  query CalendarTags {
    calendarTags
  }
`

export const calendarsApi = {
  listTags: async (): Promise<string[]> => {
    const data = await graphql<{ calendarTags: string[] }>(CALENDAR_TAGS_QUERY)
    return data.calendarTags
  },

  sync: () => graphql(`mutation RefreshEvents { refreshEvents }`),
}
