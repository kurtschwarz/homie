"use client"

import * as React from "react"

import { joinNames, type Person } from "../lib/people"
import { cn } from "../lib/utils"
import { PersonBadges } from "./person-badge"

/** Steps relative to the surface's `headline` size (112 → 96 → 84 → 72 → 60 on kiosk). */
const STEPS = [1, 0.857, 0.75, 0.643, 0.536] as const

/** Picks the largest step whose title fits in two lines of its column (three lines at the last step). */
function useHeadlineStep(ref: React.RefObject<HTMLElement | null>, title: string) {
  const [step, setStep] = React.useState(0)
  React.useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const fit = () => {
      const cs = getComputedStyle(el)
      const base = parseFloat(cs.getPropertyValue("--hm-fs-headline")) || 56
      const lh = parseFloat(cs.getPropertyValue("--hm-lh-headline")) || 1.08
      const prev = el.style.fontSize
      let chosen = STEPS.length - 1
      for (let i = 0; i < STEPS.length; i++) {
        const size = base * STEPS[i]!
        el.style.fontSize = `${size}px`
        if (el.scrollHeight <= size * lh * 2 + 2) {
          chosen = i
          break
        }
      }
      el.style.fontSize = prev
      setStep(chosen)
    }
    fit()
    const ro = new ResizeObserver(fit)
    if (el.parentElement) ro.observe(el.parentElement)
    return () => ro.disconnect()
  }, [ref, title])
  return step
}

/**
 * The home screen's headline: the one event happening now or next, as large as it can go.
 * Label: "Now · until 4:15 pm", "Next · in 2 hr 17 min" or "Next · Tomorrow".
 * Detail: time range and place ("3:30–4:15 pm · School"); names are prepended from `people`.
 * With no `title`, shows the empty state.
 */
export function NextUp({
  label,
  title,
  people = [],
  detail,
  className,
  ...props
}: React.ComponentProps<"div"> & { label?: string; title?: string; people?: Person[]; detail?: string }) {
  const ref = React.useRef<HTMLHeadingElement>(null)
  const step = useHeadlineStep(ref, title ?? "")

  if (!title) {
    return (
      <div data-slot="next-up" data-empty="" className={cn("flex flex-col gap-hm-16", className)} {...props}>
        <h1 className="m-0 text-hm-display">Nothing coming up</h1>
        <div className="text-hm-lead text-hm-ink-2">The rest of the week is clear</div>
      </div>
    )
  }

  return (
    <div data-slot="next-up" className={cn("flex flex-col gap-hm-18", className)} {...props}>
      {label && <div className="text-hm-eyebrow-lg uppercase text-hm-accent">{label}</div>}
      <h1
        ref={ref}
        data-step={step + 1}
        style={{ fontSize: `calc(var(--hm-fs-headline) * ${STEPS[step]})`, WebkitLineClamp: step === STEPS.length - 1 ? 3 : 2 }}
        className="m-0 overflow-hidden text-hm-headline text-balance [display:-webkit-box] [-webkit-box-orient:vertical] transition-[font-size] duration-300"
      >
        {title}
      </h1>
      {(people.length > 0 || detail) && (
        <div className="flex min-w-0 items-center gap-hm-16">
          {people.length > 0 && <PersonBadges people={people} size="lg" />}
          <span className="text-hm-lead text-hm-ink-2">
            {people.length > 0 && <span className="font-medium text-hm-ink">{joinNames(people)}</span>}
            {people.length > 0 && detail ? " · " : null}
            {detail}
          </span>
        </div>
      )}
    </div>
  )
}
