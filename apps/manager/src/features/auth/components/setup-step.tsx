import { ChevronDown } from 'lucide-react'
import type { ReactNode } from 'react'

export function SetupStep({
  number,
  title,
  description,
  state,
  onOpen,
  children,
}: {
  number: number
  title: string
  description: string
  state: 'active' | 'upcoming' | 'done'
  onOpen: () => void
  children: ReactNode
}) {
  const active = state === 'active'
  const upcoming = state === 'upcoming'

  return (
    <section
      className={`rounded-hm-20 bg-hm-surface p-5 sm:p-6 ${active ? 'shadow-hm-panel' : ''}`}
    >
      <div className="flex items-start gap-4">
        <span
          className={`flex size-8 shrink-0 items-center justify-center rounded-hm-6 text-hm-label ${active ? 'bg-hm-accent-soft text-hm-accent' : 'bg-hm-shell text-hm-ink-3'}`}
          aria-hidden="true"
        >
          {number}
        </span>
        <div className="min-w-0 flex-1">
          {active ? (
            <h2 className="m-0 text-hm-title-sm font-semibold text-hm-ink">{title}</h2>
          ) : (
            <button
              type="button"
              disabled={upcoming}
              onClick={onOpen}
              className="flex w-full items-start justify-between gap-3 text-left disabled:cursor-default"
            >
              <span
                className={`text-hm-title-sm font-semibold ${upcoming ? 'text-hm-ink-2' : 'text-hm-ink'}`}
              >
                {title}
              </span>
              <ChevronDown aria-hidden="true" className="mt-1 size-5 shrink-0 text-hm-ink-3" />
            </button>
          )}
          <p className="mt-1 mb-0 text-hm-body text-hm-ink-2">{description}</p>
          {active && <div className="mt-5">{children}</div>}
        </div>
      </div>
    </section>
  )
}
