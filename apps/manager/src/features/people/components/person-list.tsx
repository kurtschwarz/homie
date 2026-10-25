import {
  type Announcements,
  closestCenter,
  DndContext,
  type DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { restrictToParentElement, restrictToVerticalAxis } from '@dnd-kit/modifiers'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { PersonBadge } from '@homie/ui/components/person-badge'
import { varsFor } from '@homie/ui/lib/people'
import { GripVertical, Plus } from 'lucide-react'
import { eventsFor, type ManagerEvent } from '@/features/events'
import type { ManagerPerson } from '../types'
import { PersonWeek } from './person-week'

const HANDLE_WIDTH = 'w-6 shrink-0'

/**
 * The people in the household, one row each, with an "Add person" card at the
 * end. It is controlled: it shows what it is given and reports clicks and
 * drags, and the page decides what selecting, adding or reordering means.
 */
export function PersonList({
  people,
  events,
  showWeek,
  selectedId,
  adding,
  onSelect,
  onAdd,
  onReorder,
}: {
  /**
   * Shown in the order given. The list does no sorting of its own. People
   * flagged isFamily are shown but not draggable, so keep them last.
   */
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
  /** Called with the new order of the draggable people, as ids, after a drop. */
  onReorder: (ids: string[]) => void
}) {
  const sortable = people.filter((person) => !person.isFamily)
  const pinned = people.filter((person) => person.isFamily)
  const ids = sortable.map((person) => person.id)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )

  function nameOf(id: string | number) {
    return people.find((person) => person.id === id)?.name || 'Unnamed'
  }

  function positionOf(id: string | number) {
    return `position ${ids.indexOf(String(id)) + 1} of ${ids.length}`
  }

  const announcements: Announcements = {
    onDragStart: ({ active }) => `Picked up ${nameOf(active.id)}, ${positionOf(active.id)}.`,
    onDragOver: ({ active, over }) =>
      over ? `${nameOf(active.id)} is over ${nameOf(over.id)}, ${positionOf(over.id)}.` : undefined,
    onDragEnd: ({ active, over }) =>
      over
        ? `Dropped ${nameOf(active.id)}, ${positionOf(over.id)}.`
        : `Dropped ${nameOf(active.id)}.`,
    onDragCancel: ({ active }) => `Cancelled. ${nameOf(active.id)} stays where it was.`,
  }

  function handleDragEnd({ active, over }: DragEndEvent) {
    if (!over || active.id === over.id) return
    const from = ids.indexOf(String(active.id))
    const to = ids.indexOf(String(over.id))
    if (from < 0 || to < 0) return
    onReorder(arrayMove(ids, from, to))
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between gap-2 text-hm-caption text-hm-ink-3">
        <p className="m-0 font-semibold">{people.length} people</p>
        {sortable.length > 1 && (
          <p className="m-0 flex items-center gap-1">
            <GripVertical aria-hidden="true" size={14} />
            Drag to reorder
          </p>
        )}
      </div>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        modifiers={[restrictToVerticalAxis, restrictToParentElement]}
        accessibility={{ announcements }}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={ids} strategy={verticalListSortingStrategy}>
          <ul className="m-0 flex list-none flex-col gap-2 p-0">
            {sortable.map((person) => (
              <SortablePersonRow
                key={person.id}
                person={person}
                events={events}
                showWeek={showWeek}
                selected={person.id === selectedId}
                onSelect={onSelect}
              />
            ))}
            {pinned.map((person) => (
              <li key={person.id} className="flex items-center gap-1">
                <span aria-hidden="true" className={HANDLE_WIDTH} />
                <PersonRow
                  person={person}
                  events={events}
                  showWeek={showWeek}
                  selected={person.id === selectedId}
                  onSelect={onSelect}
                />
              </li>
            ))}
          </ul>
        </SortableContext>
      </DndContext>

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

type RowProps = {
  person: ManagerPerson
  events: ManagerEvent[]
  showWeek: boolean
  selected: boolean
  onSelect: (id: string) => void
}

function SortablePersonRow(props: RowProps) {
  const { person } = props
  // The drop reorders the list itself, which moves every row's slot. Animating
  // that on top of clearing the drag transform counts the same shift twice, so
  // rows overshoot their new slot and slide back into it.
  const {
    attributes,
    listeners,
    setNodeRef,
    setActivatorNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: person.id, animateLayoutChanges: () => false })

  return (
    <li
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), transition }}
      className={`relative flex items-center gap-1 ${isDragging ? 'z-10 opacity-90 shadow-hm-raised' : ''}`}
    >
      <button
        type="button"
        ref={setActivatorNodeRef}
        aria-label={`Reorder ${person.name || 'Unnamed'}`}
        className={`${HANDLE_WIDTH} flex cursor-grab touch-none items-center justify-center self-stretch border-0 bg-transparent p-0 text-hm-ink-3 hover:text-hm-ink active:cursor-grabbing`}
        {...attributes}
        {...listeners}
      >
        <GripVertical aria-hidden="true" size={18} />
      </button>
      <PersonRow {...props} />
    </li>
  )
}

function PersonRow({ person, events, showWeek, selected, onSelect }: RowProps) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={() => onSelect(person.id)}
      className={`flex min-w-0 flex-1 cursor-pointer items-center gap-3 rounded-hm-16 border-2 p-3 text-left ${selected ? `${varsFor(person)} border-(--p) bg-[color-mix(in_srgb,var(--p)_8%,transparent)]` : 'border-transparent bg-hm-surface hover:bg-hm-shell/50'}`}
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
  )
}
