import { Button } from '@homie/ui/components/button'
import { MoreHorizontal, RefreshCw } from 'lucide-react'
import * as React from 'react'
import { Link } from 'react-router-dom'
import type { Calendar, CalendarAccount } from '../types'
import { CalendarRow } from './calendar-row'

function syncLabel(calendars: Calendar[]) {
  const failures = calendars.filter((calendar) => calendar.lastError)
  if (failures.length) return { label: 'Sync needs attention', error: true }
  const timestamps = calendars.flatMap((calendar) =>
    calendar.lastSyncedAt === null ? [] : [calendar.lastSyncedAt],
  )
  if (!timestamps.length) return { label: 'Not synced yet', error: false }
  if (timestamps.length !== calendars.length) return { label: 'Sync incomplete', error: false }
  const minutes = Math.max(0, Math.floor((Date.now() / 1000 - Math.min(...timestamps)) / 60))
  if (minutes < 1) return { label: 'Synced just now', error: false }
  if (minutes < 60) return { label: `Synced ${minutes} min ago`, error: false }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return { label: `Synced ${hours} hr ago`, error: false }
  return { label: `Synced ${Math.floor(hours / 24)} d ago`, error: false }
}

export function AccountCard({
  account,
  calendars,
  syncing,
  syncPending,
  visibilityPending,
  onSync,
  onDelete,
  onVisibleChange,
}: {
  account: CalendarAccount
  calendars: Calendar[]
  syncing: boolean
  syncPending: boolean
  visibilityPending: boolean
  onSync: () => void
  onDelete: () => void
  onVisibleChange: (calendar: Calendar, visible: boolean) => void
}) {
  const [menuOpen, setMenuOpen] = React.useState(false)
  const [confirmDelete, setConfirmDelete] = React.useState(false)
  const status = syncLabel(calendars)

  return (
    <section className="overflow-hidden rounded-hm-16 border border-hm-line bg-hm-surface">
      <header className="flex flex-wrap items-center gap-3 p-4">
        <span className="grid size-10 place-items-center rounded-xl bg-hm-bg text-hm-title font-semibold">
          {account.name.slice(0, 1).toUpperCase()}
        </span>
        <div className="min-w-[160px] flex-1">
          <h2 className="m-0 text-hm-body font-semibold text-hm-ink">{account.name}</h2>
          <p className="m-0 truncate text-hm-caption text-hm-ink-2">
            {account.connection.username} · CalDAV
          </p>
        </div>
        <span
          className={`rounded-full px-3 py-2 text-hm-caption font-semibold ${status.error ? 'bg-red-50 text-red-700' : 'bg-hm-accent/15 text-[#087557]'}`}
        >
          {status.label}
        </span>
        <Button
          variant="outline"
          className="h-10 border px-3 text-hm-caption"
          disabled={syncPending || calendars.length === 0}
          onClick={onSync}
        >
          <RefreshCw aria-hidden="true" className={syncing ? 'animate-spin' : ''} />
          {syncing ? 'Syncing…' : 'Sync now'}
        </Button>
        <div className="relative">
          <Button
            variant="outline"
            className="h-10 w-10 border p-0"
            aria-label={`Actions for ${account.name}`}
            aria-expanded={menuOpen}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <MoreHorizontal aria-hidden="true" />
          </Button>
          {menuOpen && (
            <div className="absolute right-0 top-11 z-10 flex w-40 flex-col rounded-xl border border-hm-line bg-hm-surface p-1 shadow-hm-raised">
              <Link
                to={`${account.id}/edit`}
                className="rounded-lg px-3 py-2 text-hm-caption font-semibold text-hm-ink no-underline hover:bg-hm-bg"
              >
                Edit account
              </Link>
              {confirmDelete ? (
                <button
                  type="button"
                  className="rounded-lg border-0 bg-red-50 px-3 py-2 text-left text-hm-caption font-semibold text-red-700"
                  onClick={onDelete}
                >
                  Confirm delete
                </button>
              ) : (
                <button
                  type="button"
                  className="rounded-lg border-0 bg-transparent px-3 py-2 text-left text-hm-caption font-semibold text-red-700 hover:bg-red-50"
                  onClick={() => setConfirmDelete(true)}
                >
                  Delete account
                </button>
              )}
            </div>
          )}
        </div>
      </header>
      {calendars.length ? (
        calendars.map((calendar) => (
          <CalendarRow
            key={calendar.id}
            calendar={calendar}
            pending={visibilityPending}
            onVisibleChange={(visible) => onVisibleChange(calendar, visible)}
          />
        ))
      ) : (
        <p className="m-0 border-t border-hm-line px-4 py-4 text-hm-body text-hm-ink-2">
          No calendars discovered yet. Edit the account to try discovery again.
        </p>
      )}
    </section>
  )
}
