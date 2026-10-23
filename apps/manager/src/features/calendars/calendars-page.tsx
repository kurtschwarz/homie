import { Button } from '@homie/ui/components/button'
import { Page } from '@/app/shell/page'
import { useAppData } from '@/app-data'
import { SettingsGate } from '@/components/settings-gate'
import { TagMappingSection } from './tag-mapping-section'

export function CalendarsPage() {
  const { saveTagMapping, deleteTagMapping, syncCalendar, saving } = useAppData()

  return (
    <Page
      title="Calendars"
      description="Choose which calendar tags appear on Homie and who they belong to."
      actions={
        <Button variant="solid" onClick={() => void syncCalendar()} disabled={saving}>
          {saving ? 'Syncing…' : 'Sync calendar'}
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
            refreshing={saving}
          />
        )}
      </SettingsGate>
    </Page>
  )
}
