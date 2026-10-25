import { useQuery } from '@tanstack/react-query'
import * as React from 'react'
import { eventQueries } from '@/features/events'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import { useReorderPeople } from './api/people.mutations'
import { peopleQueries } from './api/people.queries'
import { PersonList } from './components/person-list'
import { PersonEditor } from './person-editor'

export function PeoplePage() {
  const { data: events = [], isError: eventsFailed } = useQuery(eventQueries.list())
  const people = useQuery(peopleQueries.list())
  const reorder = useReorderPeople()
  const [selectedId, setSelectedId] = React.useState<string>()
  const [adding, setAdding] = React.useState(false)
  const [editorDirty, setEditorDirty] = React.useState(false)

  function canDiscardDraft() {
    return !editorDirty || window.confirm('Discard your unsaved badge changes?')
  }

  // The catch-all person sits at the bottom of the list, so order here
  // rather than in PersonList: the default selection has to be the row
  // the reader sees first.
  const ordered = people.data
    ? [...people.data].sort((a, b) => Number(a.isFamily) - Number(b.isFamily))
    : []
  const selected = ordered.find((person) => person.id === selectedId)
  const active = adding ? undefined : (selected ?? ordered[0])

  return (
    <Page title="People" description="Who lives here, their badge, and which events are theirs">
      {people.isPending ? (
        <p className="text-hm-body text-hm-ink-2">Loading people…</p>
      ) : !people.data ? (
        <ErrorMessage
          label="People unavailable"
          message={people.error?.message ?? 'Unknown error'}
        />
      ) : (
        <>
          {people.error && <ErrorMessage message={people.error.message} />}
          {reorder.error && (
            <ErrorMessage label="Couldn't save the order" message={reorder.error.message} />
          )}
          <div className="grid items-start gap-7 lg:grid-cols-[minmax(260px,340px)_minmax(0,1fr)]">
            <PersonList
              people={ordered}
              events={events}
              showWeek={!eventsFailed}
              selectedId={active?.id}
              adding={adding}
              onReorder={(ids) => {
                const family = ordered
                  .filter((person) => person.isFamily)
                  .map((person) => person.id)
                reorder.reorder([...ids, ...family])
              }}
              onSelect={(id) => {
                if (!canDiscardDraft()) return
                setAdding(false)
                setSelectedId(id)
              }}
              onAdd={() => {
                if (!canDiscardDraft()) return
                setAdding(true)
                setSelectedId(undefined)
              }}
            />

            {active || adding ? (
              <PersonEditor
                key={active?.id ?? 'new'}
                person={active}
                people={people.data}
                events={events}
                eventsFailed={eventsFailed}
                onDirtyChange={setEditorDirty}
                onCreated={(id) => {
                  setAdding(false)
                  setSelectedId(id)
                }}
                onDeleted={() => setSelectedId(undefined)}
                onCancelAdd={() => setAdding(false)}
              />
            ) : (
              <p className="m-0 text-hm-body text-hm-ink-2">
                No one lives here yet. Choose Add person to create the first one.
              </p>
            )}
          </div>
        </>
      )}
    </Page>
  )
}
