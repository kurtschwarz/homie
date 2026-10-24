import { graphql } from '@/shared/lib/graphql'
import type { Calendar, CalendarAccount, CalendarAccountDraft } from '../types'

const ACCOUNT_FIELDS = `
  id name provider refreshIntervalSeconds windowPastSeconds windowFutureSeconds
  connection { ... on CalDAVConnection { url username timezone } }
`

const CALENDAR_FIELDS = `
  id accountId providerRef name readOnly visible lastSyncedAt lastError eventCount
`

const CALENDARS_QUERY = `
  query ManageCalendars {
    calendarAccounts { ${ACCOUNT_FIELDS} }
    calendars { ${CALENDAR_FIELDS} }
  }
`

function connection(draft: CalendarAccountDraft) {
  return {
    caldav: {
      url: draft.url,
      username: draft.username,
      password: draft.password,
      timezone: draft.timezone,
    },
  }
}

export const calendarsApi = {
  list: () =>
    graphql<{ calendarAccounts: CalendarAccount[]; calendars: Calendar[] }>(CALENDARS_QUERY),

  testDraft: (draft: CalendarAccountDraft) =>
    graphql<{ testCalendarAccount: boolean }>(
      `mutation TestCalendarAccount($input: TestCalendarAccountInput!) {
        testCalendarAccount(input: $input)
      }`,
      { input: { connection: connection(draft) } },
    ),

  testStored: (id: string) =>
    graphql<{ testCalendarAccount: boolean }>(
      `mutation TestStoredCalendarAccount($input: TestCalendarAccountInput!) {
        testCalendarAccount(input: $input)
      }`,
      { input: { id } },
    ),

  create: async (draft: CalendarAccountDraft): Promise<CalendarAccount> => {
    const data = await graphql<{ createCalendarAccount: CalendarAccount }>(
      `mutation CreateCalendarAccount($input: CreateCalendarAccountInput!) {
        createCalendarAccount(input: $input) { ${ACCOUNT_FIELDS} }
      }`,
      { input: { name: draft.name, connection: connection(draft) } },
    )
    return data.createCalendarAccount
  },

  update: async ({ id, draft }: { id: string; draft: CalendarAccountDraft }) => {
    const caldav: Record<string, string> = {
      url: draft.url,
      username: draft.username,
      timezone: draft.timezone,
    }
    if (draft.password) caldav.password = draft.password
    const data = await graphql<{ updateCalendarAccount: CalendarAccount }>(
      `mutation UpdateCalendarAccount($input: UpdateCalendarAccountInput!) {
        updateCalendarAccount(input: $input) { ${ACCOUNT_FIELDS} }
      }`,
      { input: { id, name: draft.name, connection: { caldav } } },
    )
    return data.updateCalendarAccount
  },

  remove: (id: string) =>
    graphql(`mutation DeleteCalendarAccount($id: ID!) { deleteCalendarAccount(id: $id) }`, {
      id,
    }),

  discover: (accountId: string) =>
    graphql<{ discoverCalendars: Calendar[] }>(
      `mutation DiscoverCalendars($accountId: ID!) {
        discoverCalendars(accountId: $accountId) { ${CALENDAR_FIELDS} }
      }`,
      { accountId },
    ),

  setVisibility: ({ id, visible }: { id: string; visible: boolean }) =>
    graphql<{ updateCalendarVisibility: Calendar }>(
      `mutation UpdateCalendarVisibility($id: ID!, $visible: Boolean!) {
        updateCalendarVisibility(id: $id, visible: $visible) { ${CALENDAR_FIELDS} }
      }`,
      { id, visible },
    ),

  refresh: (id: string) =>
    graphql(`mutation RefreshCalendar($id: ID!) { refreshCalendar(id: $id) }`, { id }),
}
