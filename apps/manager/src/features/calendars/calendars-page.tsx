import { Button } from '@homie/ui/components/button'
import { useAppData } from '../../app-data'
import { SettingsGate } from '../../components/settings-gate'
import { Page } from '../../shell/page'
import { TagMappingSection } from './tag-mapping-section'

export function CalendarsPage() {
  const { state, saveTagMapping, deleteTagMapping, refreshEvents } = useAppData()
  const syncing = state.kind === 'ready' && state.saving

  return (
    <Page
      title="Calendars"
      description="Choose which calendar tags appear on Homie and who they belong to."
      actions={
        <Button variant="solid" onClick={refreshEvents} disabled={syncing}>
          {syncing ? 'Syncing…' : 'Sync calendar'}
        </Button>
      }
    >
      <SettingsGate>
        {(settings) => (
          <TagMappingSection
            tagMappings={settings.tagMappings}
            calendarTags={settings.calendarTags}
            people={settings.people}
            onSave={saveTagMapping}
            onDelete={deleteTagMapping}
            refreshing={syncing}
          />
        )}
      </SettingsGate>
    </Page>
  )
}
