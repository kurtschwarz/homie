import { Button } from '@homie/ui/components/button'
import { useQuery } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { Link } from 'react-router-dom'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import {
  useCalendarVisibility,
  useDeleteCalendarAccount,
  useSyncCalendarAccount,
} from './api/calendars.mutations'
import { calendarQueries } from './api/calendars.queries'
import { AccountCard } from './components/account-card'

export function CalendarsPage() {
  const result = useQuery(calendarQueries.list())
  const sync = useSyncCalendarAccount()
  const visibility = useCalendarVisibility()
  const remove = useDeleteCalendarAccount()

  const accounts = result.data?.calendarAccounts ?? []
  const calendars = result.data?.calendars ?? []
  const visibleCount = calendars.filter((calendar) => calendar.visible).length

  return (
    <Page
      title="Calendars"
      description="Where Homie gets its events from"
      actions={
        <Button
          variant="solid"
          nativeButton={false}
          render={<Link to="add" className="no-underline" />}
        >
          <Plus aria-hidden="true" />
          Add account
        </Button>
      }
    >
      {result.isPending ? (
        <p className="m-0 text-hm-body text-hm-ink-2">Loading calendar accounts…</p>
      ) : result.error ? (
        <ErrorMessage label="Calendars unavailable" message={result.error.message} />
      ) : (
        <>
          <p className="m-0 text-hm-caption text-hm-ink-2">
            <strong className="text-hm-ink">
              {accounts.length} {accounts.length === 1 ? 'account' : 'accounts'}
            </strong>
            {' · '}
            {calendars.length} {calendars.length === 1 ? 'calendar' : 'calendars'}
            {' · '}
            {visibleCount} on the kitchen screen
          </p>

          {(sync.error ?? visibility.error ?? remove.error) && (
            <ErrorMessage
              label="Calendar update failed"
              message={(sync.error ?? visibility.error ?? remove.error)?.message ?? 'Unknown error'}
            />
          )}

          <div className="flex flex-col gap-4">
            {accounts.map((account) => {
              const accountCalendars = calendars.filter(
                (calendar) => calendar.accountId === account.id,
              )
              return (
                <AccountCard
                  key={account.id}
                  account={account}
                  calendars={accountCalendars}
                  syncing={sync.isPending && sync.variables?.accountId === account.id}
                  syncPending={sync.isPending}
                  visibilityPending={visibility.isPending}
                  onSync={() => sync.mutate({ accountId: account.id, calendars: accountCalendars })}
                  onDelete={() => remove.mutate(account.id)}
                  onVisibleChange={(calendar, visible) =>
                    visibility.mutate({ id: calendar.id, visible })
                  }
                />
              )
            })}
            {accounts.length === 0 && (
              <div className="rounded-hm-16 border border-dashed border-hm-line px-6 py-12 text-center">
                <h2 className="m-0 text-hm-title font-semibold">Add your first calendar</h2>
                <p className="mb-5 mt-2 text-hm-body text-hm-ink-2">
                  Connect a CalDAV account to bring events into Homie.
                </p>
                <Button
                  variant="outline"
                  nativeButton={false}
                  render={<Link to="add" className="no-underline" />}
                >
                  <Plus aria-hidden="true" /> Add a calendar account
                </Button>
              </div>
            )}
          </div>

          {accounts.length > 0 && (
            <Button
              variant="outline"
              block
              nativeButton={false}
              render={<Link to="add" className="no-underline" />}
            >
              <Plus aria-hidden="true" /> Add a calendar account
            </Button>
          )}
        </>
      )}
    </Page>
  )
}
