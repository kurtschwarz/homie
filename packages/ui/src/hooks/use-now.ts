'use client'

import * as React from 'react'

/** The current time, refreshed every `intervalMs` (5 s by default, enough for a minute clock). */
export function useNow(intervalMs = 5_000) {
  const [now, setNow] = React.useState(() => new Date())
  React.useEffect(() => {
    const id = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])
  return now
}
