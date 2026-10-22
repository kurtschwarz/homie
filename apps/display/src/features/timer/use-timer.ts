import { useCallback, useEffect, useState } from 'react'
import { graphql } from '../../lib/graphql'

export type ActiveTimer = {
  title: string
  emoji: string
  durationSeconds: number
  remainingSeconds: number
  endsAt: number
  paused: boolean
}

const TIMER_QUERY = `query ActiveTimer { timer { title emoji durationSeconds remainingSeconds endsAt paused } }`

export function useTimer() {
  const [timer, setTimer] = useState<ActiveTimer | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await graphql<{ timer: ActiveTimer | null }>(TIMER_QUERY)
      setTimer(data.timer)
    } catch {
      // Keep the last known timer during a temporary network failure.
    }
  }, [])

  useEffect(() => {
    void refresh()
    const interval = window.setInterval(refresh, 1000)
    return () => window.clearInterval(interval)
  }, [refresh])

  const run = useCallback(
    async (mutation: string) => {
      await graphql(mutation)
      await refresh()
    },
    [refresh],
  )

  return {
    timer,
    addMinute: () => run('mutation { addTimerMinute { title } }'),
    togglePause: () =>
      run(
        timer?.paused ? 'mutation { resumeTimer { title } }' : 'mutation { pauseTimer { title } }',
      ),
    cancel: () => run('mutation { cancelTimer }'),
  }
}
