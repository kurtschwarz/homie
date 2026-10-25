import { useQuery } from '@tanstack/react-query'
import { Plus, X } from 'lucide-react'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ErrorMessage } from '@/shared/ui/error-message'
import {
  useCreateCalendarAssociation,
  useDeleteCalendarAssociation,
  useUpdateCalendarAssociation,
} from '../api/associations.mutations'
import { associationQueries } from '../api/associations.queries'
import { calendarQueries } from '../api/calendars.queries'
import type { AssociationScope, Calendar, CalendarAssociation } from '../types'

const SCOPES: { scope: AssociationScope; label: string }[] = [
  { scope: 'MATCHED', label: 'Tagged' },
  { scope: 'ALL', label: 'All' },
]

function sameTerm(a: string, b: string) {
  return a.trim().toLowerCase() === b.trim().toLowerCase()
}

/**
 * Which calendars one person appears on, and how: by tag or word, or by
 * claiming the whole calendar. Every change saves straight away.
 *
 * Switching to All keeps the tags, so switching back restores the list.
 */
export function PersonCalendars({
  personId,
  personName,
}: {
  personId: string
  personName: string
}) {
  const calendars = useQuery(calendarQueries.list())
  const associations = useQuery(associationQueries.forPerson(personId))
  const create = useCreateCalendarAssociation()
  const update = useUpdateCalendarAssociation()
  const remove = useDeleteCalendarAssociation()
  const pending = create.isPending || update.isPending || remove.isPending
  const error = create.error ?? update.error ?? remove.error

  if (calendars.isPending || associations.isPending) {
    return <p className="m-0 text-hm-body text-hm-ink-2">Loading calendars…</p>
  }
  if (!calendars.data || !associations.data) {
    return (
      <ErrorMessage
        label="Calendars unavailable"
        message={(calendars.error ?? associations.error)?.message ?? 'Unknown error'}
      />
    )
  }

  const { calendarAccounts, calendars: list } = calendars.data
  const byCalendar = new Map(associations.data.map((entry) => [entry.calendarId, entry]))
  const accountName = new Map(calendarAccounts.map((account) => [account.id, account.name]))
  const showAccount = calendarAccounts.length > 1

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="m-0 text-hm-body font-semibold text-hm-ink">Calendars</h3>
        <Link to="/calendars" className="text-hm-body font-semibold text-hm-accent no-underline">
          Manage
        </Link>
      </div>
      <p className="m-0 text-hm-body text-hm-ink-2">
        <strong className="text-hm-ink">Tagged</strong> picks out events that mention a tag.{' '}
        <strong className="text-hm-ink">All</strong> claims the whole calendar.
      </p>

      {error && <ErrorMessage label="Couldn't update calendars" message={error.message} />}

      {list.length === 0 ? (
        <p className="m-0 text-hm-body text-hm-ink-2">
          No calendars yet. Add an account under Calendars to connect one.
        </p>
      ) : (
        <ul className="m-0 flex list-none flex-col gap-2 p-0">
          {list.map((calendar) => (
            <li key={calendar.id}>
              <CalendarOption
                calendar={calendar}
                accountName={showAccount ? accountName.get(calendar.accountId) : undefined}
                personName={personName}
                association={byCalendar.get(calendar.id)}
                disabled={pending}
                onEnable={() => create.mutate({ calendarId: calendar.id, personId })}
                onDisable={() => remove.mutate({ calendarId: calendar.id, personId })}
                onChange={(patch) => update.mutate({ calendarId: calendar.id, personId, ...patch })}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function CalendarOption({
  calendar,
  accountName,
  personName,
  association,
  disabled,
  onEnable,
  onDisable,
  onChange,
}: {
  calendar: Calendar
  accountName?: string
  personName: string
  association?: CalendarAssociation
  disabled: boolean
  onEnable: () => void
  onDisable: () => void
  onChange: (patch: { scope?: AssociationScope; terms?: string[] }) => void
}) {
  const on = Boolean(association)
  const checkboxId = `calendar-${calendar.id}`

  return (
    <div
      className={`flex flex-col gap-3 rounded-hm-16 px-4 py-3 ${on ? 'bg-hm-accent/10' : 'bg-hm-bg'}`}
    >
      <div className="flex items-center gap-3">
        <input
          id={checkboxId}
          type="checkbox"
          checked={on}
          disabled={disabled}
          onChange={(event) => (event.target.checked ? onEnable() : onDisable())}
          className="size-6 shrink-0 cursor-pointer accent-hm-accent disabled:cursor-not-allowed"
        />
        <label
          htmlFor={checkboxId}
          className="min-w-0 flex-1 cursor-pointer truncate text-hm-body font-semibold text-hm-ink"
        >
          {calendar.name}
          {accountName && (
            <span className="ml-2 text-hm-caption font-normal text-hm-ink-3">{accountName}</span>
          )}
        </label>
        {association ? (
          <ScopeToggle
            calendarName={calendar.name}
            scope={association.scope}
            disabled={disabled}
            onChange={(scope) => onChange({ scope })}
          />
        ) : (
          <span className="text-hm-caption text-hm-ink-3">Off</span>
        )}
      </div>

      {association &&
        (association.scope === 'ALL' ? (
          <p className="m-0 pl-9 text-hm-body text-hm-ink-2">
            Every event on {calendar.name} is {personName}’s.
          </p>
        ) : (
          <TermList
            terms={association.terms}
            disabled={disabled}
            onChange={(terms) => onChange({ terms })}
          />
        ))}
    </div>
  )
}

function ScopeToggle({
  calendarName,
  scope,
  disabled,
  onChange,
}: {
  calendarName: string
  scope: AssociationScope
  disabled: boolean
  onChange: (scope: AssociationScope) => void
}) {
  return (
    <fieldset className="m-0 flex min-w-0 rounded-full border-0 bg-hm-surface p-1">
      <legend className="sr-only">How {calendarName} is matched</legend>
      {SCOPES.map((entry) => (
        <button
          key={entry.scope}
          type="button"
          aria-pressed={scope === entry.scope}
          disabled={disabled}
          onClick={() => scope !== entry.scope && onChange(entry.scope)}
          className={`cursor-pointer rounded-full border-0 px-4 py-1.5 text-hm-body font-semibold disabled:cursor-not-allowed ${
            scope === entry.scope ? 'bg-hm-accent text-white' : 'bg-transparent text-hm-ink-2'
          }`}
        >
          {entry.label}
        </button>
      ))}
    </fieldset>
  )
}

function TermList({
  terms,
  disabled,
  onChange,
}: {
  terms: string[]
  disabled: boolean
  onChange: (terms: string[]) => void
}) {
  const [adding, setAdding] = React.useState(false)
  const [value, setValue] = React.useState('')

  function commit() {
    const label = value.trim()
    setAdding(false)
    setValue('')
    if (!label || terms.some((term) => sameTerm(term, label))) return
    onChange([...terms, label])
  }

  return (
    <div className="flex flex-wrap items-center gap-2 pl-9">
      <span className="text-hm-body text-hm-ink-2">Events mentioning</span>
      {terms.map((term) => (
        <span
          key={term}
          className="inline-flex items-center gap-2 rounded-full bg-hm-accent/15 px-3 py-1 text-hm-body text-hm-ink"
        >
          <span className={term.startsWith('#') ? 'font-semibold' : ''}>{term}</span>
          <button
            type="button"
            aria-label={`Remove ${term}`}
            disabled={disabled}
            onClick={() => onChange(terms.filter((entry) => entry !== term))}
            className="flex cursor-pointer items-center border-0 bg-transparent p-0 text-hm-ink-3 disabled:cursor-not-allowed"
          >
            <X aria-hidden="true" className="size-4" />
          </button>
        </span>
      ))}
      {adding ? (
        <input
          // biome-ignore lint/a11y/noAutofocus: the field only appears after the user asks to add a tag
          autoFocus
          value={value}
          aria-label="Tag or word"
          placeholder="#tag or a word"
          onChange={(event) => setValue(event.target.value)}
          onBlur={commit}
          onKeyDown={(event) => {
            if (event.key === 'Enter') commit()
            if (event.key === 'Escape') {
              setAdding(false)
              setValue('')
            }
          }}
          className="w-40 rounded-full border border-hm-line bg-hm-surface px-3 py-1 text-hm-body text-hm-ink"
        />
      ) : (
        <button
          type="button"
          disabled={disabled}
          onClick={() => setAdding(true)}
          className="inline-flex cursor-pointer items-center gap-1 rounded-full border border-dashed border-hm-line bg-transparent px-3 py-1 text-hm-body font-semibold text-hm-ink-2 disabled:cursor-not-allowed"
        >
          <Plus aria-hidden="true" className="size-4" />
          Add tag
        </button>
      )}
    </div>
  )
}
