import { SmilePlus } from 'lucide-react'
import * as React from 'react'

const EMOJIS = ['🦊', '🐢', '🦋', '☕', '🏠', '🐻', '🐸', '🦄', '🐙', '🦖', '🐝', '🌻', '⚽', '🚀']

/**
 * The curated badge emoji, plus a final tile for anything else. An emoji
 * already stored but outside the list keeps its place on that tile, so the
 * grid never quietly rewrites someone's badge.
 */
export function EmojiPicker({
  value,
  onChange,
}: {
  /**
   * The selected emoji, or an empty string for none. A value outside the
   * curated list is shown on the custom tile rather than dropped.
   */
  value: string
  /**
   * Called with the chosen emoji. Typing in the custom field calls this on
   * every keystroke with the trimmed text, so it can be an empty string
   * when the field is cleared.
   */
  onChange: (emoji: string) => void
}) {
  const isCustom = Boolean(value) && !EMOJIS.includes(value)
  const [editingCustom, setEditingCustom] = React.useState(isCustom)
  const [custom, setCustom] = React.useState(isCustom ? value : '')

  return (
    <div>
      <div className="flex max-w-[560px] flex-wrap gap-2">
        {EMOJIS.map((emoji) => (
          <button
            key={emoji}
            type="button"
            aria-label={`Use ${emoji}`}
            aria-pressed={value === emoji}
            onClick={() => {
              setEditingCustom(false)
              onChange(emoji)
            }}
            className={`grid size-10 place-items-center rounded-xl border-2 bg-hm-bg text-xl ${value === emoji ? 'border-hm-accent' : 'border-transparent hover:border-hm-line'}`}
          >
            {emoji}
          </button>
        ))}
        <button
          type="button"
          aria-label="Use a custom emoji"
          aria-pressed={editingCustom || isCustom}
          onClick={() => {
            setEditingCustom(true)
            if (custom.trim()) onChange(custom.trim())
          }}
          className={`grid size-10 place-items-center rounded-xl border-2 border-dashed text-xl text-hm-ink-3 ${editingCustom || isCustom ? 'border-hm-accent' : 'border-hm-line hover:border-hm-ink-3'}`}
        >
          {isCustom ? value : <SmilePlus aria-hidden="true" size={20} />}
        </button>
      </div>
      {editingCustom ? (
        <input
          value={custom}
          onChange={(event) => {
            setCustom(event.target.value)
            onChange(event.target.value.trim())
          }}
          placeholder="Paste an emoji"
          aria-label="Custom emoji"
          className="mt-3 w-40 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink"
        />
      ) : (
        <p className="mb-0 mt-3 text-hm-caption text-hm-ink-3">
          Not in the list? Use the last button to pick any emoji.
        </p>
      )}
    </div>
  )
}
