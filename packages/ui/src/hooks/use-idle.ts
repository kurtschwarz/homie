"use client"

import * as React from "react"

/**
 * Calls `onIdle` after `timeoutMs` without a touch while `active` is true.
 * Homie uses it to drift back to the home screen after 60 s on Week or Weather.
 */
export function useIdle(active: boolean, onIdle: () => void, timeoutMs = 60_000) {
  const cb = React.useRef(onIdle)
  cb.current = onIdle
  React.useEffect(() => {
    if (!active) return
    let id = setTimeout(() => cb.current(), timeoutMs)
    const reset = () => {
      clearTimeout(id)
      id = setTimeout(() => cb.current(), timeoutMs)
    }
    window.addEventListener("pointerdown", reset, { passive: true })
    window.addEventListener("keydown", reset)
    return () => {
      clearTimeout(id)
      window.removeEventListener("pointerdown", reset)
      window.removeEventListener("keydown", reset)
    }
  }, [active, timeoutMs])
}
