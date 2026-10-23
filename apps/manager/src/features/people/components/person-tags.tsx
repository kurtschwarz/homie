import { Button } from '@homie/ui/components/button'
import { useQuery } from '@tanstack/react-query'
import * as React from 'react'
import { tagQueries, useDeleteTagMapping, useSaveTagMapping } from '@/features/tags'
import { ErrorMessage } from '@/shared/ui/error-message'

/**
 * The tags that give events one person's badge, with a field to add more.
 *
 * Unlike the Badge tab, nothing here waits for a Save button: adding or
 * removing a tag is its own mutation and takes effect straight away.
 *
 * A tag already mapped elsewhere can be moved here after confirmation.
 *
 * The API matches a tag against an event's calendar categories only. A word
 * that appears in an event's title is not matched yet, even though the field
 * accepts free text.
 */
export function PersonTags({
  personId,
  personName,
}: {
  /** The person the tags belong to. New tags are saved against this id. */
  personId: string
  /** Used in the explanatory copy, e.g. "show with Lana’s badge too". */
  personName: string
}) {
  const [newTag, setNewTag] = React.useState('')
  const mappings = useQuery(tagQueries.list())
  const saveTag = useSaveTagMapping()
  const deleteTag = useDeleteTagMapping()

  if (mappings.isPending) {
    return <p className="m-0 text-hm-body text-hm-ink-2">Loading tags…</p>
  }
  if (!mappings.data) {
    return (
      <ErrorMessage label="Tags unavailable" message={mappings.error?.message ?? 'Unknown error'} />
    )
  }

  // Keep the PERSON mappings that point at this person and ignore the rest.
  const tagMappings = mappings.data
  const tags = tagMappings.filter(
    (mapping) => mapping.kind === 'PERSON' && mapping.person?.id === personId,
  )

  function addTag(event: React.FormEvent) {
    event.preventDefault()
    const tag = newTag.trim()
    if (!tag) return
    const existing = tagMappings.find((mapping) => mapping.tag === tag.toLowerCase())
    if (existing && existing.person?.id !== personId) {
      const owner = existing.person?.name ?? 'Show on Homie'
      if (!window.confirm(`Move “${existing.label}” from ${owner} to ${personName}?`)) return
    }
    saveTag.mutate({ tag, kind: 'PERSON', personId }, { onSuccess: () => setNewTag('') })
  }

  return (
    <section>
      <h2 className="m-0 text-hm-caption font-semibold text-hm-ink">Tags and words</h2>
      <p className="mb-3 mt-2 max-w-[360px] text-hm-body text-hm-ink-2">
        Shared calendar events that mention one of these show with {personName}’s badge too.
      </p>
      {saveTag.error && <ErrorMessage label="Couldn't add tag" message={saveTag.error.message} />}
      {deleteTag.error && (
        <ErrorMessage label="Couldn't remove tag" message={deleteTag.error.message} />
      )}
      <div className="mb-3 flex min-h-9 flex-wrap gap-2">
        {tags.map((mapping) => (
          <span
            key={mapping.tag}
            className="inline-flex items-center gap-2 rounded-full bg-hm-accent/10 px-3 py-2 text-hm-caption font-semibold text-hm-ink"
          >
            {mapping.label}
            <button
              type="button"
              aria-label={`Remove ${mapping.label}`}
              disabled={deleteTag.isPending}
              onClick={() => deleteTag.mutate(mapping.tag)}
              className="border-0 bg-transparent p-0 text-hm-ink-3"
            >
              ×
            </button>
          </span>
        ))}
      </div>
      <form onSubmit={addTag} className="flex max-w-[420px] gap-2">
        <input
          value={newTag}
          onChange={(event) => setNewTag(event.target.value)}
          placeholder="#tag or a word"
          aria-label="Tag or word"
          className="min-w-0 flex-1 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink"
        />
        <Button type="submit" variant="text" disabled={!newTag.trim() || saveTag.isPending}>
          Add
        </Button>
      </form>
    </section>
  )
}
