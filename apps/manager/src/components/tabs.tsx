/**
 * A row of tabs. Controlled and panel-less: it only chooses which tab is
 * selected, and the caller renders the matching panel.
 *
 * `T` is the union of your tab ids, so `value` and `onChange` are typed to it
 * rather than to a plain string.
 *
 * Arrow keys move through enabled tabs and select the focused tab. Home and
 * End choose the first and last enabled tabs. The caller must give each panel
 * the id `${id}-${tabId}-panel` and label it with `${id}-${tabId}-tab`.
 */
export function Tabs<T extends string>({
  tabs,
  value,
  onChange,
  label,
  id,
}: {
  /**
   * The tabs, in display order. A `disabled` tab is dimmed and cannot be
   * selected, but still takes its place in the row.
   */
  tabs: { id: T; label: string; disabled?: boolean }[]
  /** The id of the selected tab. */
  value: T
  /** Called with the id of the tab the user chose. */
  onChange: (id: T) => void
  /** Names the tablist for assistive technology. */
  label: string
  /** Stable prefix used to associate each tab with its panel. */
  id: string
}) {
  const buttons = useRef<(HTMLButtonElement | null)[]>([])

  function choose(index: number) {
    const tab = tabs[index]
    if (!tab || tab.disabled) return
    onChange(tab.id)
    buttons.current[index]?.focus()
  }

  function move(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return
    event.preventDefault()
    if (event.key === 'Home') {
      choose(tabs.findIndex((tab) => !tab.disabled))
      return
    }
    if (event.key === 'End') {
      let last = tabs.length - 1
      while (last >= 0 && tabs[last]?.disabled) last--
      choose(last)
      return
    }
    const direction = event.key === 'ArrowRight' ? 1 : -1
    let next = index
    do {
      next = (next + direction + tabs.length) % tabs.length
    } while (tabs[next]?.disabled && next !== index)
    choose(next)
  }

  return (
    <div
      role="tablist"
      aria-label={label}
      className="flex gap-1 self-start rounded-hm-control bg-hm-bg p-1"
    >
      {tabs.map((tab, index) => (
        <button
          key={tab.id}
          ref={(button) => {
            buttons.current[index] = button
          }}
          id={`${id}-${tab.id}-tab`}
          type="button"
          role="tab"
          aria-selected={value === tab.id}
          aria-controls={`${id}-${tab.id}-panel`}
          tabIndex={value === tab.id ? 0 : -1}
          disabled={tab.disabled}
          onClick={() => onChange(tab.id)}
          onKeyDown={(event) => move(event, index)}
          className={`cursor-pointer rounded-hm-control px-4 py-2 text-hm-body disabled:cursor-not-allowed disabled:opacity-40 ${value === tab.id ? 'bg-hm-surface text-hm-ink shadow-hm-raised' : 'text-hm-ink-2'}`}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}

import { type KeyboardEvent, useRef } from 'react'
