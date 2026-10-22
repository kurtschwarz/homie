import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { graphql } from '../../lib/graphql'

export type Photo = {
  id: string
  url: string
  width: number
  height: number
  takenAt: number
  place?: string | null
  focus?: { x: number; y: number; width: number; height: number } | null
}

export type PhotoShowing = {
  id: string
  photo: Photo
  startedAt: number
  endsAt: number
  endedAt?: number | null
  source: 'SCHEDULE' | 'TAP'
}

export type PhotoShowingState = { current: PhotoShowing | null; next: PhotoShowing | null }

const ROTATION_QUERY = `query PhotoShowing { photoShowing { current { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } next { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } } }`
const ROTATION_MUTATION = (source: 'SCHEDULE' | 'TAP') =>
  `mutation TrackPhotoShowing { trackPhotoShowing(source: ${source}) { current { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } next { id photo { id url width height takenAt place focus { x y width height } } startedAt endsAt source } } }`

const QUERY_KEY = ['photoShowing']

async function fetchPhotoShowing() {
  const { photoShowing } = await graphql<{ photoShowing: PhotoShowingState }>(ROTATION_QUERY)
  return photoShowing
}

/** Polls the API-owned slideshow timeline and advances expired slots. */
export function usePhotoShowing() {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: QUERY_KEY,
    queryFn: fetchPhotoShowing,
    refetchInterval: 1000,
  })

  const advance = useMutation({
    mutationFn: async (source: 'SCHEDULE' | 'TAP') => {
      const { trackPhotoShowing } = await graphql<{ trackPhotoShowing: PhotoShowingState }>(
        ROTATION_MUTATION(source),
      )
      return trackPhotoShowing
    },
    onSuccess: (result) => {
      queryClient.setQueryData(QUERY_KEY, result)
    },
  })

  const showing = query.data ?? null
  const { mutate, isPending } = advance

  // Keyed on dataUpdatedAt so a failed advance retries once per poll rather
  // than on every mutation state change.
  // biome-ignore lint/correctness/useExhaustiveDependencies: showing and isPending are read at poll time only
  useEffect(() => {
    if (!showing?.current || !showing.next || isPending) return
    if (showing.current.endsAt <= Math.floor(Date.now() / 1000)) {
      mutate('SCHEDULE')
    }
  }, [query.dataUpdatedAt, mutate])

  return {
    showing,
    advance: (source: 'SCHEDULE' | 'TAP') => {
      if (!isPending) mutate(source)
    },
  }
}
