import { Button } from '@homie/ui/components/button'
import { useQuery } from '@tanstack/react-query'
import { Check, ChevronLeft, ChevronRight, Eye, EyeOff, Server } from 'lucide-react'
import * as React from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import {
  useCreateCalendarAccount,
  useDiscoverCalendars,
  useTestDraftCalendarAccount,
  useTestStoredCalendarAccount,
  useUpdateCalendarAccount,
} from './api/calendars.mutations'
import { calendarQueries } from './api/calendars.queries'
import { CALENDAR_PROVIDERS } from './lib/providers'
import type { CalendarAccountDraft } from './types'

export function AddCalendarAccountPage() {
  const { accountId } = useParams()
  const navigate = useNavigate()
  const existingQuery = useQuery(calendarQueries.list())
  const existing = existingQuery.data?.calendarAccounts.find((account) => account.id === accountId)
  const [providerId, setProviderId] = React.useState('fastmail')
  const provider =
    CALENDAR_PROVIDERS.find((candidate) => candidate.id === providerId) ?? CALENDAR_PROVIDERS[0]
  const [draft, setDraft] = React.useState<CalendarAccountDraft>({
    name: provider.name,
    url: provider.url,
    username: '',
    password: '',
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
  })
  const [showPassword, setShowPassword] = React.useState(false)
  const [tested, setTested] = React.useState(false)
  const [createdAccountId, setCreatedAccountId] = React.useState<string | null>(null)
  const seededAccount = React.useRef<string | null>(null)
  const create = useCreateCalendarAccount()
  const update = useUpdateCalendarAccount()
  const discover = useDiscoverCalendars()
  const testDraft = useTestDraftCalendarAccount()
  const testStored = useTestStoredCalendarAccount()

  React.useEffect(() => {
    if (!existing || seededAccount.current === existing.id) return
    seededAccount.current = existing.id
    const matchingProvider = CALENDAR_PROVIDERS.find(
      (candidate) => candidate.url && existing.connection.url.startsWith(candidate.url),
    )
    setProviderId(matchingProvider?.id ?? 'other')
    setDraft({
      name: existing.name,
      url: existing.connection.url,
      username: existing.connection.username,
      password: '',
      timezone: existing.connection.timezone,
    })
  }, [existing])

  function selectProvider(id: string) {
    const next = CALENDAR_PROVIDERS.find((candidate) => candidate.id === id)
    if (!next?.available || accountId) return
    setProviderId(id)
    setDraft((current) => ({ ...current, name: next.name, url: next.url }))
    setTested(false)
  }

  const endpointChanged = Boolean(
    existing &&
      (draft.url !== existing.connection.url || draft.username !== existing.connection.username),
  )

  async function testConnection() {
    setTested(false)
    if (accountId && !endpointChanged && !draft.password) await testStored.mutateAsync(accountId)
    else await testDraft.mutateAsync(draft)
    setTested(true)
  }

  async function save() {
    if (accountId) {
      await update.mutateAsync({ id: accountId, draft })
      await discover.mutateAsync(accountId)
    } else {
      let id = createdAccountId
      if (!id) {
        const account = await create.mutateAsync(draft)
        id = account.id
        setCreatedAccountId(id)
      } else {
        await update.mutateAsync({ id, draft })
      }
      await discover.mutateAsync(id)
    }
    navigate('/calendars')
  }

  const pending = create.isPending || update.isPending || discover.isPending
  const testPending = testDraft.isPending || testStored.isPending
  const error =
    create.error ?? update.error ?? discover.error ?? testDraft.error ?? testStored.error
  const valid = Boolean(
    draft.name.trim() &&
      draft.url.trim() &&
      draft.username.trim() &&
      (accountId ? !endpointChanged || draft.password : draft.password),
  )

  return (
    <Page title="Calendars" description="Where Homie gets its events from">
      <nav
        aria-label="Breadcrumb"
        className="flex items-center gap-2 text-hm-caption font-semibold"
      >
        <Link to="/calendars" className="flex items-center gap-1 text-hm-ink-2 no-underline">
          <ChevronLeft aria-hidden="true" className="size-4" /> Calendars
        </Link>
        <span aria-hidden="true">/</span>
        <span>{accountId ? 'Edit account' : 'Add an account'}</span>
      </nav>

      <div className="grid min-h-0 flex-1 gap-6 lg:grid-cols-[380px_minmax(0,1fr)]">
        <section className="flex flex-col gap-2">
          <h2 className="m-0 mb-1 text-hm-caption font-semibold">1 · Pick your provider</h2>
          {CALENDAR_PROVIDERS.map((candidate) => {
            const selected = candidate.id === providerId
            return (
              <button
                type="button"
                key={candidate.id}
                disabled={!candidate.available || Boolean(accountId)}
                onClick={() => selectProvider(candidate.id)}
                className={`flex min-h-16 items-center gap-3 rounded-hm-16 border px-3 py-2 text-left disabled:opacity-50 ${selected ? 'border-2 border-hm-accent bg-hm-accent/15' : 'border-hm-line bg-hm-surface'}`}
              >
                <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-hm-bg text-hm-body font-semibold">
                  {candidate.name.slice(0, 1)}
                </span>
                <span className="min-w-0 flex-1">
                  <strong className="block truncate text-hm-body">{candidate.name}</strong>
                  <span className="block truncate text-hm-caption text-hm-ink-2">
                    {candidate.detail}
                  </span>
                </span>
                {selected ? (
                  <Check aria-hidden="true" className="size-5 text-hm-accent" />
                ) : (
                  <ChevronRight aria-hidden="true" className="size-5 text-hm-ink-3" />
                )}
              </button>
            )
          })}
        </section>

        <section className="flex min-h-[540px] flex-col rounded-hm-16 border border-hm-line p-6">
          <header className="mb-5 flex items-center gap-3">
            <span className="grid size-11 place-items-center rounded-xl bg-hm-bg text-hm-title font-semibold">
              {provider.name.slice(0, 1)}
            </span>
            <div>
              <h2 className="m-0 text-hm-title font-semibold">2 · Sign in to {provider.name}</h2>
              <p className="m-0 text-hm-caption text-hm-ink-2">
                Enter the CalDAV details for this account.
              </p>
            </div>
          </header>

          {error && <ErrorMessage label="Connection failed" message={error.message} />}
          {tested && !error && (
            <p className="rounded-xl bg-hm-accent/15 px-4 py-3 text-hm-caption font-semibold text-[#087557]">
              Connection successful
            </p>
          )}

          <div className="flex flex-col gap-4">
            {accountId && (
              <Field
                label="Account name"
                value={draft.name}
                onChange={(name) => setDraft((current) => ({ ...current, name }))}
              />
            )}
            <Field
              label="Server"
              value={draft.url}
              icon={<Server aria-hidden="true" />}
              onChange={(url) => {
                setDraft((current) => ({ ...current, url }))
                setTested(false)
              }}
            />
            <Field
              label={provider.userLabel}
              value={draft.username}
              onChange={(username) => {
                setDraft((current) => ({ ...current, username }))
                setTested(false)
              }}
            />
            <Field
              label={
                accountId
                  ? `${provider.passwordLabel} (leave blank to keep it)`
                  : provider.passwordLabel
              }
              type={showPassword ? 'text' : 'password'}
              value={draft.password}
              onChange={(password) => {
                setDraft((current) => ({ ...current, password }))
                setTested(false)
              }}
              trailing={
                <button
                  type="button"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  className="border-0 bg-transparent p-1 text-hm-ink-3"
                  onClick={() => setShowPassword((shown) => !shown)}
                >
                  {showPassword ? <EyeOff aria-hidden="true" /> : <Eye aria-hidden="true" />}
                </button>
              }
            />
          </div>

          <div className="mt-5 rounded-hm-16 bg-hm-bg p-4">
            <h3 className="m-0 mb-3 text-hm-caption font-semibold">Getting connected</h3>
            <ol className="m-0 flex list-none flex-col gap-2 p-0">
              {provider.help.map((step, index) => (
                <li key={step} className="flex gap-3 text-hm-caption text-hm-ink-2">
                  <span className="grid size-5 shrink-0 place-items-center rounded-full bg-hm-accent/15 font-semibold text-[#087557]">
                    {index + 1}
                  </span>
                  {step}
                </li>
              ))}
            </ol>
          </div>

          <div className="mt-auto flex flex-wrap justify-end gap-3 pt-6">
            <Button
              variant="outline"
              disabled={!valid || testPending || pending}
              onClick={testConnection}
            >
              {testPending ? 'Testing…' : 'Test connection'}
            </Button>
            <Button variant="solid" disabled={!valid || pending || testPending} onClick={save}>
              {pending ? 'Connecting…' : accountId ? 'Save changes' : 'Connect'}
            </Button>
          </div>
        </section>
      </div>
    </Page>
  )
}

function Field({
  label,
  value,
  onChange,
  type = 'text',
  icon,
  trailing,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  type?: string
  icon?: React.ReactNode
  trailing?: React.ReactNode
}) {
  const id = React.useId()
  return (
    <label htmlFor={id} className="flex flex-col gap-2 text-hm-caption font-semibold text-hm-ink-2">
      {label}
      <span className="flex h-11 items-center gap-2 rounded-hm-control border border-hm-line bg-hm-surface px-3 focus-within:outline-3 focus-within:outline-offset-2 focus-within:outline-hm-accent">
        {icon && <span className="[&_svg]:size-5">{icon}</span>}
        <input
          id={id}
          type={type}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          className="min-w-0 flex-1 border-0 bg-transparent text-hm-body text-hm-ink outline-none"
        />
        {trailing}
      </span>
    </label>
  )
}
