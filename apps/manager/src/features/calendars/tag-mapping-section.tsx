import { Button } from '@homie/ui/components/button'
import * as React from 'react'
import type { ManagerPerson } from '@/features/people'
import { type TagMapping, useDeleteTagMapping, useSaveTagMapping } from '@/features/tags'
import { ErrorMessage } from '@/shared/ui/error-message'

export function TagMappingSection({
  tagMappings,
  calendarTags,
  people,
}: {
  tagMappings: TagMapping[]
  calendarTags: string[]
  people: ManagerPerson[]
}) {
  const mappingByTag = new Map(tagMappings.map((m) => [m.tag, m]))
  const tags = [...new Set([...calendarTags, ...tagMappings.map((m) => m.tag)])].sort()
  const [newTag, setNewTag] = React.useState('')
  const saveTag = useSaveTagMapping()
  const deleteTag = useDeleteTagMapping()

  // calendarTags only covers the synced window, so a tag used outside it is
  // invisible here and cannot be mapped — including, on a fresh install,
  // every tag. Adding by hand is the only way to break that deadlock.
  const addTag = (e: React.FormEvent) => {
    e.preventDefault()
    const tag = newTag.trim()
    if (!tag) return
    saveTag.mutate({ tag, kind: 'INCLUDE' }, { onSuccess: () => setNewTag('') })
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
      {saveTag.error && <ErrorMessage label="Couldn't save tag" message={saveTag.error.message} />}
      {deleteTag.error && (
        <ErrorMessage label="Couldn't remove tag" message={deleteTag.error.message} />
      )}
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
                    if (mapping) deleteTag.mutate(tag)
                    return
                  }
                  if (value === 'include') {
                    saveTag.mutate({ tag: mapping?.label ?? tag, kind: 'INCLUDE' })
                    return
                  }
                  saveTag.mutate({ tag: mapping?.label ?? tag, kind: 'PERSON', personId: value })
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
                <Button
                  type="button"
                  variant="text"
                  disabled={deleteTag.isPending}
                  onClick={() => deleteTag.mutate(tag)}
                >
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
        <Button type="submit" variant="solid" disabled={saveTag.isPending}>
          Add tag
        </Button>
      </form>
    </section>
  )
}
