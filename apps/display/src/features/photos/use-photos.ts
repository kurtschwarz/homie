import type { Photo } from '@homie/ui/components/photo-panel'
import { useQuery } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'
import { pollEvery } from '../../lib/query-client'

type PhotoRecord = {
  url: string
  takenAt: number
  place?: string | null
  focus?: Photo['focus'] | null
}

const QUERY = `
  query Photos {
    photos {
      id
      url
      width
      height
      takenAt
      place
      focus { x y width height }
    }
  }
`

function select(data: { photos?: PhotoRecord[] }): Photo[] {
  if (!data.photos) throw new Error('Photos query failed')
  return data.photos.map((photo) => ({
    src: photo.url,
    alt: photo.place ?? '',
    takenAt: new Date(photo.takenAt * 1000),
    ...(photo.place ? { place: photo.place } : {}),
    ...(photo.focus ? { focus: photo.focus } : {}),
  }))
}

async function fetchPhotos() {
  return graphql<{ photos?: PhotoRecord[] }>(QUERY)
}

/** Fetches the current photo library and refreshes periodically. */
export function usePhotos() {
  return useQuery({
    queryKey: ['photos'],
    queryFn: fetchPhotos,
    select,
    refetchInterval: pollEvery(5 * 60_000),
  })
}
