import { PersonBadge } from '@homie/ui/components/person-badge'
import { varsFor } from '@homie/ui/lib/people'
import { Plus } from 'lucide-react'
import { eventsFor, type ManagerEvent } from '@/features/events'
import type { ManagerPerson } from '../types'
import { PersonWeek } from './person-week'

/**
 * The people in the household, one row each, with an "Add person" card at the
 * end. It is controlled: it shows what it is given and reports clicks, and the
 * page decides what selecting or adding means.
 */
export function PersonList({
  people,
  events,
  showWeek,
  selectedId,
  adding,
  onSelect,
  onAdd,
}: {
  /** Shown in the order given. The list does no sorting of its own. */
  people: ManagerPerson[]
  /** The coming week's events, used for each row's day dots. */
  events: ManagerEvent[]
  /** Set to false to hide the day dots, e.g. when events failed to load. */
  showWeek: boolean
  /** The highlighted row. Leave unset when no one is selected. */
  selectedId?: string
  /** Highlights the "Add person" card instead of a row. */
  adding: boolean
  /** Called with the id of the person whose row was clicked. */
  onSelect: (id: string) => void
  /** Called when the "Add person" card is clicked. */
  onAdd: () => void
}) {
  return (
    <div className="flex flex-col gap-3">
      <p className="m-0 text-hm-caption text-hm-ink-3">{people.length} people</p>

      <ul className="m-0 flex list-none flex-col gap-2 p-0">
        {people.map((person) => {
          const selected = person.id === selectedId
          return (
            <li key={person.id}>
              <button
                type="button"
                aria-pressed={selected}
                onClick={() => onSelect(person.id)}
                className={`flex w-full cursor-pointer items-center gap-3 rounded-hm-16 border-2 p-3 text-left ${selected ? `${varsFor(person)} border-(--p) bg-[color-mix(in_srgb,var(--p)_8%,transparent)]` : 'border-transparent bg-hm-surface hover:bg-hm-shell/50'}`}
              >
                <PersonBadge person={person} size="lg" />
                <span className="flex min-w-0 flex-col gap-1.5">
                  <span className="truncate text-hm-body font-semibold text-hm-ink">
                    {person.name || 'Unnamed'}
                  </span>
                  {showWeek && (
                    <PersonWeek
                      events={eventsFor(events, person.id)}
                      swatch={person.swatch}
                      name={person.name}
                    />
                  )}
                </span>
              </button>
            </li>
          )
        })}
      </ul>

      <button
        type="button"
        aria-pressed={adding}
        onClick={onAdd}
        className={`flex w-full cursor-pointer items-center gap-3 rounded-hm-16 border-2 p-3 text-left ${adding ? 'border-hm-accent bg-hm-accent-soft' : 'border-hm-line bg-hm-surface hover:border-hm-accent'}`}
      >
        <span className="grid size-hm-badge place-items-center rounded-full bg-hm-accent-soft text-hm-accent">
          <Plus aria-hidden="true" size={20} />
        </span>
        <span className="flex flex-col">
          <span className="text-hm-body font-semibold text-hm-accent">Add person</span>
          <span className="text-hm-caption text-hm-ink-3">Someone new, or a group</span>
        </span>
      </button>

      <p className="m-0 mt-2 text-hm-caption text-hm-ink-3">
        Everyone gets a badge on the kitchen screen. Events show with a badge when they mention one
        of someone's tags.
      </p>
    </div>
  )
}
