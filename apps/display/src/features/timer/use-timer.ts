import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { graphql } from '../../lib/graphql'

export type ActiveTimer = {
  title: string
  emoji: string
  durationSeconds: number
  remainingSeconds: number
  endsAt: number
  paused: boolean
}

const QUERY_KEY = ['timer']
const TIMER_QUERY = `query ActiveTimer { timer { title emoji durationSeconds remainingSeconds endsAt paused } }`

async function fetchTimer() {
  const data = await graphql<{ timer: ActiveTimer | null }>(TIMER_QUERY)
  return data.timer
}

export function useTimer() {
  const queryClient = useQueryClient()

  const query = useQuery({
    queryKey: QUERY_KEY,
    queryFn: fetchTimer,
    refetchInterval: 1000,
  })

  const run = useMutation({
    mutationFn: (mutation: string) => graphql(mutation),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: QUERY_KEY }),
  })

  const timer = query.data ?? null

  return {
    timer,
    addMinute: () => run.mutateAsync('mutation { addTimerMinute { title } }').then(() => {}),
    togglePause: () =>
      run
        .mutateAsync(
          timer?.paused
            ? 'mutation { resumeTimer { title } }'
            : 'mutation { pauseTimer { title } }',
        )
        .then(() => {}),
    cancel: () => run.mutateAsync('mutation { cancelTimer }').then(() => {}),
  }
}
