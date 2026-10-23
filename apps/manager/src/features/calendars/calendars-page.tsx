import { Button } from '@homie/ui/components/button'
import { useQuery } from '@tanstack/react-query'
import { peopleQueries } from '@/features/people'
import { tagQueries } from '@/features/tags'
import { ErrorMessage } from '@/shared/ui/error-message'
import { Page } from '@/shared/ui/page'
import { useSyncCalendar } from './api/calendars.mutations'
import { calendarQueries } from './api/calendars.queries'
import { TagMappingSection } from './tag-mapping-section'

export function CalendarsPage() {
  const syncCalendar = useSyncCalendar()
  const people = useQuery(peopleQueries.list())
  const tags = useQuery(tagQueries.list())
  const calendarTags = useQuery(calendarQueries.tags())

  const pending = people.isPending || tags.isPending || calendarTags.isPending
  const loadError = people.error ?? tags.error ?? calendarTags.error

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
      {people.data && tags.data && calendarTags.data ? (
        <>
          {loadError && <ErrorMessage message={loadError.message} />}
          <TagMappingSection
            tagMappings={tags.data}
            calendarTags={calendarTags.data}
            people={people.data}
          />
        </>
      ) : pending ? (
        <p className="text-hm-body text-hm-ink-2">Loading calendar settings…</p>
      ) : (
        <ErrorMessage message={loadError?.message ?? 'Calendar settings unavailable'} />
      )}
    </Page>
  )
}
