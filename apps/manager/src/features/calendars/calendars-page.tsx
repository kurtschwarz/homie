import { Button } from '@homie/ui/components/button'
import { Page } from '@/app/shell/page'
import { HouseholdGate, useSyncCalendar } from '@/features/household'
import { ErrorMessage } from '@/shared/ui/error-message'
import { TagMappingSection } from './tag-mapping-section'

export function CalendarsPage() {
  const syncCalendar = useSyncCalendar()

  return (
    <Page
      title="Calendars"
      description="Choose which calendar tags appear on Homie and who they belong to."
      actions={
        <Button
          variant="solid"
          onClick={() => syncCalendar.mutate()}
          disabled={syncCalendar.isPending}
        >
          {syncCalendar.isPending ? 'Syncing…' : 'Sync calendar'}
        </Button>
      }
    >
      {syncCalendar.error && (
        <ErrorMessage label="Couldn't sync calendar" message={syncCalendar.error.message} />
      )}
      <HouseholdGate>
        {(household) => (
          <TagMappingSection
            tagMappings={household.tagMappings}
            calendarTags={household.calendarTags}
            people={household.people}
          />
        )}
      </HouseholdGate>
    </Page>
  )
}
