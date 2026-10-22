import { Button } from '@homie/ui/components/button'
import * as React from 'react'
import type { ManagerPerson, TagKind, TagMapping } from '../../lib/use-settings'

export function TagMappingSection({
  tagMappings,
  calendarTags,
  people,
  onSave,
  onDelete,
  refreshing,
}: {
  tagMappings: TagMapping[]
  calendarTags: string[]
  people: ManagerPerson[]
  onSave: (input: { tag: string; kind: TagKind; personId?: string }) => void
  onDelete: (tag: string) => void
  refreshing: boolean
}) {
  const mappingByTag = new Map(tagMappings.map((m) => [m.tag, m]))
  const tags = [...new Set([...calendarTags, ...tagMappings.map((m) => m.tag)])].sort()
  const [newTag, setNewTag] = React.useState('')

  // calendarTags only covers the synced window, so a tag used outside it is
  // invisible here and cannot be mapped — including, on a fresh install,
  // every tag. Adding by hand is the only way to break that deadlock.
  const addTag = (e: React.FormEvent) => {
    e.preventDefault()
    const tag = newTag.trim()
    if (!tag) return
    onSave({ tag, kind: 'INCLUDE' })
    setNewTag('')
  }

  return (
    <section className="flex flex-col gap-4">
      <div className="flex flex-col gap-2">
        <p className="m-0 text-hm-body text-hm-ink-2">
          Map a calendar tag to a person, or to <strong>Show on Homie</strong> to bring the event in
          without attributing it to anyone in particular. Only tags on events inside the synced
          window are listed.
        </p>
      </div>
      <div className="flex flex-col gap-2">
        {tags.map((tag) => {
          const mapping = mappingByTag.get(tag)
          const stale = !calendarTags.includes(tag)
          return (
            <div
              key={tag}
              className="flex flex-wrap items-center gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-3"
            >
              <span className="min-w-0 flex-1 truncate text-hm-body text-hm-ink">
                {mapping?.label ?? tag}
                {stale && (
                  <span className="ml-2 text-hm-caption text-hm-ink-3">(not seen recently)</span>
                )}
              </span>
              <select
                aria-label={`Mapping for ${tag}`}
                value={
                  mapping
                    ? mapping.kind === 'INCLUDE'
                      ? 'include'
                      : (mapping.person?.id ?? '')
                    : 'unmapped'
                }
                onChange={(e) => {
                  const value = e.target.value
                  if (value === 'unmapped') {
                    if (mapping) onDelete(tag)
                    return
                  }
                  if (value === 'include') {
                    onSave({ tag: mapping?.label ?? tag, kind: 'INCLUDE' })
                    return
                  }
                  onSave({ tag: mapping?.label ?? tag, kind: 'PERSON', personId: value })
                }}
                className="rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink"
              >
                <option value="unmapped">Unmapped</option>
                <option value="include">Show on Homie</option>
                {people.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
              {mapping && (
                <Button type="button" variant="text" onClick={() => onDelete(tag)}>
                  Remove
                </Button>
              )}
            </div>
          )
        })}
        {tags.length === 0 && (
          <p className="m-0 text-hm-body text-hm-ink-2">
            No tags seen in the calendar yet. Add one below if you know its name.
          </p>
        )}
      </div>
      <form
        onSubmit={addTag}
        className="flex flex-wrap items-center gap-3 rounded-hm-16 border border-hm-line p-4"
      >
        <input
          value={newTag}
          onChange={(e) => setNewTag(e.target.value)}
          placeholder="Tag name, as spelled in the calendar"
          aria-label="Tag name"
          className="min-w-0 flex-1 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink"
        />
        <Button type="submit" variant="solid">
          Add tag
        </Button>
      </form>
      {refreshing && (
        <span className="sr-only" aria-live="polite">
          Syncing calendar…
        </span>
      )}
    </section>
  )
}
