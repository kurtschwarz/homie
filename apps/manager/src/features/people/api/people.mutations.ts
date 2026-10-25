import { useMutation, useQueryClient } from '@tanstack/react-query'
import { eventKeys } from '@/features/events'
import type { ManagerPerson } from '../types'
import { peopleApi } from './people.api'
import { peopleKeys } from './people.queries'

function usePeopleInvalidation() {
  const queryClient = useQueryClient()
  // A person's badge is on their events, so saving or removing someone
  // refreshes those too.
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: peopleKeys.all }),
      queryClient.invalidateQueries({ queryKey: eventKeys.all }),
    ])
}

export function useSavePerson() {
  return useMutation({
    mutationFn: peopleApi.save,
    onSuccess: usePeopleInvalidation(),
  })
}

export function useReorderPeople() {
  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: peopleApi.reorder,
    onSettled: () => queryClient.invalidateQueries({ queryKey: peopleKeys.all }),
  })

  // The cache is written synchronously, in the same tick as the drop. mutate's
  // own onMutate runs a microtask later, after dnd-kit has already cleared its
  // drag transforms, so the list would flash back to the old order first.
  function reorder(ids: string[]) {
    const previous = queryClient.getQueryData<ManagerPerson[]>(peopleKeys.list())
    if (previous) {
      const byId = new Map(previous.map((person) => [person.id, person]))
      queryClient.setQueryData(
        peopleKeys.list(),
        ids.flatMap((id) => byId.get(id) ?? []),
      )
    }
    void queryClient.cancelQueries({ queryKey: peopleKeys.all })
    mutation.mutate(ids, {
      onError: () => {
        if (previous) queryClient.setQueryData(peopleKeys.list(), previous)
      },
    })
  }

  return { reorder, error: mutation.error }
}

export function useDeletePerson() {
  return useMutation({
    mutationFn: peopleApi.remove,
    onSuccess: usePeopleInvalidation(),
  })
}
