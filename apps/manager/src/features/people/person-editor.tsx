import { Button } from '@homie/ui/components/button'
import { PersonBadge } from '@homie/ui/components/person-badge'
import { DEFAULT_SWATCH, type Swatch } from '@homie/ui/lib/people'
import { formatTime, shortDayLabel } from '@homie/ui/lib/time'
import { Check } from 'lucide-react'
import * as React from 'react'
import { eventsFor, type ManagerEvent } from '@/features/events'
import {
  type ManagerPerson,
  type TagMapping,
  useDeletePerson,
  useSavePerson,
} from '@/features/household'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Tabs } from '@/shared/ui/tabs'
import { PersonTags } from './components/person-tags'
import { PersonIdentity } from './person-identity'

export type PersonDraft = {
  name: string
  emoji: string
  swatch: Swatch
  isFamily: boolean
}

type Tab = 'badge' | 'calendars' | 'tags'

const TABS: { id: Tab; label: string }[] = [
  { id: 'badge', label: 'Badge' },
  { id: 'calendars', label: 'Calendars' },
  { id: 'tags', label: 'Tags and words' },
]

function draftFor(person?: ManagerPerson): PersonDraft {
  return {
    name: person?.name ?? '',
    emoji: person?.emoji ?? '',
    swatch: person?.swatch ?? DEFAULT_SWATCH,
    isFamily: person?.isFamily ?? false,
  }
}

/**
 * The card for one person: who they are at the top, then a tab per thing you
 * can change about them. Badge edits are held locally until Save, so a
 * half-typed name never reaches the API; tags save as they are added because
 * each is its own mutation.
 *
 * Mounted under a key of the person's id, so selecting someone else reseeds
 * the draft rather than carrying the previous person's edits across.
 */
export function PersonEditor({
  person,
  people,
  tagMappings,
  events,
  eventsFailed,
  onDirtyChange,
  onCreated,
  onDeleted,
  onCancelAdd,
}: {
  person?: ManagerPerson
  people: ManagerPerson[]
  tagMappings: TagMapping[]
  events: ManagerEvent[]
  eventsFailed: boolean
  onDirtyChange: (dirty: boolean) => void
  onCreated: (id: string) => void
  onDeleted: () => void
  onCancelAdd: () => void
}) {
  const saved = draftFor(person)
  const [draft, setDraft] = React.useState(saved)
  const [revision, setRevision] = React.useState(0)
  const [tab, setTab] = React.useState<Tab>('badge')
  const [confirmingRemove, setConfirmingRemove] = React.useState(false)
  const savePerson = useSavePerson()
  const deletePerson = useDeletePerson()

  const dirty =
    draft.name !== saved.name ||
    draft.emoji !== saved.emoji ||
    draft.swatch !== saved.swatch ||
    draft.isFamily !== saved.isFamily
  const named = draft.name.trim().length > 0
  const personEvents = person ? eventsFor(events, person.id) : []

  React.useEffect(() => onDirtyChange(dirty), [dirty, onDirtyChange])
  React.useEffect(() => () => onDirtyChange(false), [onDirtyChange])

  function save() {
    if (!named) return
    savePerson.mutate(
      { ...draft, id: person?.id, name: draft.name.trim() },
      { onSuccess: (id) => !person && onCreated(id) },
    )
  }

  function cancel() {
    if (!person) {
      onCancelAdd()
      return
    }
    setDraft(saved)
    // EmojiPicker keeps the custom input locally. Remount the form so Cancel
    // resets that UI state as well as the controlled draft values.
    setRevision((current) => current + 1)
  }

  return (
    <div className="flex flex-col gap-6 rounded-hm-16 border border-hm-line bg-hm-surface p-6">
      <header className="flex items-center gap-4">
        <PersonBadge
          person={{
            id: person?.id ?? 'new',
            name: draft.name || '?',
            emoji: draft.emoji,
            swatch: draft.swatch,
          }}
          size="lg"
        />
        <div className="flex min-w-0 flex-col gap-1">
          <h2 className="m-0 truncate text-hm-title font-semibold text-hm-ink">
            {draft.name.trim() || 'New person'}
          </h2>
          <p className="m-0 text-hm-caption text-hm-ink-2">
            {!person
              ? 'Not saved yet'
              : eventsFailed
                ? 'Could not load this week’s events'
                : `${personEvents.length} ${personEvents.length === 1 ? 'event' : 'events'} with this badge in the next 7 days`}
          </p>
        </div>
      </header>

      <Tabs
        id="person-settings"
        label="Person settings"
        value={tab}
        onChange={setTab}
        tabs={TABS.map((entry) => ({ ...entry, disabled: !person && entry.id !== 'badge' }))}
      />

      {savePerson.error && (
        <ErrorMessage label="Couldn't save person" message={savePerson.error.message} />
      )}
      {deletePerson.error && (
        <ErrorMessage label="Couldn't remove person" message={deletePerson.error.message} />
      )}

      {tab === 'badge' && (
        <div
          id="person-settings-badge-panel"
          role="tabpanel"
          aria-labelledby="person-settings-badge-tab"
          className="flex flex-col gap-6"
        >
          <PersonIdentity
            key={revision}
            draft={draft}
            people={people}
            personId={person?.id}
            onChange={(patch) => setDraft((previous) => ({ ...previous, ...patch }))}
          />

          <div className="flex flex-wrap items-center gap-3 border-t border-hm-line pt-5">
            {person &&
              (confirmingRemove ? (
                <div className="flex flex-wrap items-center gap-2 text-hm-caption text-hm-ink-2">
                  <span>Remove {person.name} and their tags?</span>
                  <Button
                    type="button"
                    variant="text"
                    disabled={deletePerson.isPending}
                    onClick={() =>
                      deletePerson.mutate(person.id, {
                        onSuccess: onDeleted,
                      })
                    }
                  >
                    Confirm
                  </Button>
                  <Button type="button" variant="text" onClick={() => setConfirmingRemove(false)}>
                    Keep
                  </Button>
                </div>
              ) : (
                <Button
                  type="button"
                  variant="text"
                  className="rounded-hm-control bg-hm-bg px-hm-20 font-semibold text-hm-ink"
                  onClick={() => setConfirmingRemove(true)}
                >
                  Remove {person.name}
                </Button>
              ))}

            <div className="ml-auto flex flex-wrap items-center gap-3">
              {dirty && (
                <span className="flex items-center gap-2 text-hm-caption text-hm-ink-2">
                  <span
                    aria-hidden="true"
                    className="size-2 rounded-full"
                    style={{ backgroundColor: 'var(--hm-swatch-amber)' }}
                  />
                  Unsaved changes
                </span>
              )}
              <Button
                type="button"
                variant="outline"
                disabled={savePerson.isPending || (!dirty && Boolean(person))}
                onClick={cancel}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="solid"
                disabled={!dirty || !named || savePerson.isPending}
                onClick={save}
              >
                <Check aria-hidden="true" />
                {savePerson.isPending ? 'Saving…' : 'Save changes'}
              </Button>
            </div>
          </div>
        </div>
      )}

      {tab === 'calendars' && (
        <div
          id="person-settings-calendars-panel"
          role="tabpanel"
          aria-labelledby="person-settings-calendars-tab"
          className="flex flex-col gap-2"
        >
          <p className="m-0 text-hm-body text-hm-ink-2">
            Homie syncs one calendar today, so a calendar cannot belong to a person yet. Until it
            can, claim someone’s events under <strong>Tags and words</strong>.
          </p>
        </div>
      )}

      {tab === 'tags' && person && (
        <div
          id="person-settings-tags-panel"
          role="tabpanel"
          aria-labelledby="person-settings-tags-tab"
        >
          <PersonTags personId={person.id} personName={person.name} tagMappings={tagMappings} />
        </div>
      )}

      {person && !eventsFailed && (
        <section className="flex flex-col gap-3 border-t border-hm-line pt-5">
          <div className="flex flex-wrap items-baseline justify-between gap-2">
            <h3 className="m-0 text-hm-caption font-semibold text-hm-ink">On the kitchen screen</h3>
            <span className="text-hm-caption text-hm-ink-3">
              {personEvents.length} in the next 7 days
            </span>
          </div>
          {personEvents.length === 0 ? (
            <p className="m-0 text-hm-body text-hm-ink-2">
              Nothing yet. Events arrive here when they mention one of their tags.
            </p>
          ) : (
            <ul className="m-0 flex list-none flex-wrap gap-2 p-0">
              {personEvents.slice(0, 6).map((event) => (
                <li
                  key={event.id}
                  className="flex min-w-[200px] flex-1 items-center gap-3 rounded-hm-16 bg-hm-bg px-3 py-2.5"
                >
                  <PersonBadge
                    person={{
                      id: person.id,
                      name: draft.name,
                      emoji: draft.emoji,
                      swatch: draft.swatch,
                    }}
                    size="chip"
                  />
                  <span className="flex min-w-0 flex-col">
                    <span className="text-hm-caption text-hm-ink-3">
                      {shortDayLabel(event.start)}
                      {!event.allDay && ` · ${formatTime(event.start)}`}
                    </span>
                    <span className="truncate text-hm-body font-semibold text-hm-ink">
                      {event.title}
                    </span>
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  )
}
