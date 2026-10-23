import * as React from 'react'

import { useAppData } from '../../app-data'
import { SettingsGate } from '../../components/settings-gate'
import { useEvents } from '../../lib/use-events'
import { Page } from '../../shell/page'
import { PersonList } from './components/person-list'
import { PersonEditor } from './person-editor'

export function PeoplePage() {
  const { savePerson, deletePerson, saveTagMapping, deleteTagMapping } = useAppData()
  // Every mutation below invalidates the events query, so there is nothing to
  // refresh by hand after a save: the week strip catches up on its own.
  const { data: events = [], isError: eventsFailed } = useEvents()
  const [selectedId, setSelectedId] = React.useState<string>()
  const [adding, setAdding] = React.useState(false)
  const [editorDirty, setEditorDirty] = React.useState(false)

  function canDiscardDraft() {
    return !editorDirty || window.confirm('Discard your unsaved badge changes?')
  }

  return (
    <Page title="People" description="Who lives here, their badge, and which events are theirs">
      <SettingsGate>
        {(settings) => {
          // The catch-all person sits at the bottom of the list, so order here
          // rather than in PersonList: the default selection has to be the row
          // the reader sees first.
          const ordered = [...settings.people].sort(
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
                  people={settings.people}
                  tagMappings={settings.tagMappings}
                  events={events}
                  eventsFailed={eventsFailed}
                  onDirtyChange={setEditorDirty}
                  onSave={savePerson}
                  onCreated={(id) => {
                    setAdding(false)
                    setSelectedId(id)
                  }}
                  onDelete={async () => {
                    if (!active) return
                    const deleted = await deletePerson(active.id)
                    if (!deleted) return
                    setSelectedId(undefined)
                  }}
                  onCancelAdd={() => setAdding(false)}
                  onSaveTag={saveTagMapping}
                  onDeleteTag={deleteTagMapping}
                />
              ) : (
                <p className="m-0 text-hm-body text-hm-ink-2">
                  No one lives here yet. Choose Add person to create the first one.
                </p>
              )}
            </div>
          )
        }}
      </SettingsGate>
    </Page>
  )
}
