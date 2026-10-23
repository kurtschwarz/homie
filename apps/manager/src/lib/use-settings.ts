import type { Swatch } from '@homie/ui/lib/people'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useCallback } from 'react'
import { graphql } from '@/shared/lib/graphql'

export type ManagerPerson = {
  id: string
  name: string
  emoji: string
  swatch: Swatch
  isFamily: boolean
}

export type TagKind = 'PERSON' | 'INCLUDE'

export type TagMapping = {
  tag: string
  label: string
  kind: TagKind
  person: ManagerPerson | null
}

export type Settings = {
  people: ManagerPerson[]
  tagMappings: TagMapping[]
  calendarTags: string[]
}

const SETTINGS_QUERY = `
  query Settings {
    people { id name emoji swatch isFamily }
    tagMappings { tag label kind person { id name emoji swatch isFamily } }
    calendarTags
  }
`

export type PersonInput = {
  id?: string
  name: string
  emoji: string
  swatch: string
  isFamily: boolean
}
export type TagMappingInput = { tag: string; kind: TagKind; personId?: string }

type SaveRequest = { query: string; variables?: Record<string, unknown> }

/**
 * One hook for people and tag mappings: they are edited together (assigning
 * a tag to a person needs the people list to be current), and every
 * mutation refetches both. Refetching rather than patching the cache is
 * the right call for a single-user settings page where correctness beats a
 * frame of latency.
 *
 * All five mutations share one `useMutation`, so there is a single `saving`
 * flag and a single `saveError` to show, however the save was started. That
 * matches how the pages read it — one banner above the settings, one disabled
 * state — and is why `AppDataProvider` holds this hook once for the shell
 * instead of each page calling it.
 */
export function useSettings() {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: ['settings'],
    queryFn: () => graphql<Settings>(SETTINGS_QUERY),
  })

  const save = useMutation({
    mutationFn: (request: SaveRequest) =>
      graphql<Record<string, unknown>>(request.query, request.variables),
    // Returning the invalidations makes `mutateAsync` resolve only once the
    // refetches have landed, so a caller that saves and then reads is never a
    // render behind. Events are attributed to people by tag, so every one of
    // these mutations can change which events belong to whom.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: ['settings'] }),
        queryClient.invalidateQueries({ queryKey: ['events'] }),
      ]),
  })

  /**
   * Runs one mutation and resolves to `undefined` if it failed rather than
   * throwing. The error is already on `saveError` for the page to show, and
   * the editors read a falsy result as "not saved" — a rejection would leave
   * them stuck on "Saving…".
   */
  const run = useCallback(
    async <T>(query: string, variables?: Record<string, unknown>) => {
      try {
        return (await save.mutateAsync({ query, variables })) as T
      } catch {
        return undefined
      }
    },
    [save.mutateAsync],
  )

  const savePerson = useCallback(
    async (input: PersonInput) => {
      const data = await run<{ savePerson: { id: string } }>(
        `mutation SavePerson($input: PersonInput!) { savePerson(input: $input) { id } }`,
        { input },
      )
      return data?.savePerson.id
    },
    [run],
  )

  const deletePerson = useCallback(
    (id: string) => run(`mutation DeletePerson($id: ID!) { deletePerson(id: $id) }`, { id }),
    [run],
  )

  const saveTagMapping = useCallback(
    (input: TagMappingInput) =>
      run(
        `mutation SaveTagMapping($input: TagMappingInput!) { saveTagMapping(input: $input) { tag } }`,
        { input },
      ),
    [run],
  )

  const deleteTagMapping = useCallback(
    (tag: string) =>
      run(`mutation DeleteTagMapping($tag: String!) { deleteTagMapping(tag: $tag) }`, { tag }),
    [run],
  )

  // Going through the same lane means the refetch afterwards picks up any tag
  // the sync just discovered, which is the whole point of syncing from here.
  const syncCalendar = useCallback(() => run(`mutation RefreshEvents { refreshEvents }`), [run])

  return {
    settings: query.data,
    isPending: query.isPending,
    error: query.error,
    saving: save.isPending,
    saveError: save.error,
    savePerson,
    deletePerson,
    saveTagMapping,
    deleteTagMapping,
    syncCalendar,
  }
}
