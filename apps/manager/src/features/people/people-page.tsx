import { useQuery } from '@tanstack/react-query'
import * as React from 'react'
import { Page } from '@/app/shell/page'
import { eventQueries } from '@/features/events'
import { HouseholdGate } from '@/features/household'
import { PersonList } from './components/person-list'
import { PersonEditor } from './person-editor'

export function PeoplePage() {
  const { data: events = [], isError: eventsFailed } = useQuery(eventQueries.list())
  const [selectedId, setSelectedId] = React.useState<string>()
  const [adding, setAdding] = React.useState(false)
  const [editorDirty, setEditorDirty] = React.useState(false)

  function canDiscardDraft() {
    return !editorDirty || window.confirm('Discard your unsaved badge changes?')
  }

  return (
    <Page title="People" description="Who lives here, their badge, and which events are theirs">
      <HouseholdGate>
        {(household) => {
          // The catch-all person sits at the bottom of the list, so order here
          // rather than in PersonList: the default selection has to be the row
          // the reader sees first.
          const ordered = [...household.people].sort(
            (a, b) => Number(a.isFamily) - Number(b.isFamily),
          )
          const selected = ordered.find((person) => person.id === selectedId)
          const active = adding ? undefined : (selected ?? ordered[0])

          return (
            <div className="grid items-start gap-7 lg:grid-cols-[minmax(260px,340px)_minmax(0,1fr)]">
              <PersonList
                people={ordered}
                events={events}
                showWeek={!eventsFailed}
                selectedId={active?.id}
                adding={adding}
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
                  people={household.people}
                  tagMappings={household.tagMappings}
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
          )
        }}
      </HouseholdGate>
    </Page>
  )
}
