import { SWATCHES } from '@homie/ui/lib/people'
import type { ManagerPerson } from '@/lib/use-settings'
import { FormSection } from '@/shared/ui/form-section'
import { EmojiPicker } from './components/emoji-picker'
import type { PersonDraft } from './person-editor'

/** The Badge tab: name, emoji and colour. Controlled by PersonEditor, which owns the draft. */
export function PersonIdentity({
  draft,
  people,
  personId,
  onChange,
}: {
  draft: PersonDraft
  people: ManagerPerson[]
  personId?: string
  onChange: (patch: Partial<PersonDraft>) => void
}) {
  const swatchInUse = people.some((other) => other.id !== personId && other.swatch === draft.swatch)

  return (
    <div className="flex flex-col gap-6">
      <label className="flex max-w-[420px] flex-col text-hm-caption font-semibold text-hm-ink-2">
        Name
        <input
          value={draft.name}
          onChange={(event) => onChange({ name: event.target.value })}
          placeholder="Name"
          className="mt-2 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2.5 text-hm-body font-normal text-hm-ink"
        />
      </label>

      <FormSection label="Emoji">
        <EmojiPicker value={draft.emoji} onChange={(emoji) => onChange({ emoji })} />
      </FormSection>

      <FormSection label="Colour">
        <div className="flex max-w-[480px] flex-wrap gap-2">
          {SWATCHES.map((option) => (
            <button
              key={option}
              type="button"
              aria-label={option}
              aria-pressed={draft.swatch === option}
              onClick={() => onChange({ swatch: option })}
              className={`size-9 rounded-full border-[3px] ring-offset-2 ring-offset-hm-surface ${draft.swatch === option ? 'border-hm-surface ring-2 ring-hm-accent' : 'border-transparent'}`}
              style={{ backgroundColor: `var(--hm-swatch-${option})` }}
            />
          ))}
        </div>
        <p className="mb-0 mt-3 text-hm-caption text-hm-ink-3">
          {swatchInUse ? 'Someone else uses this colour.' : 'Nobody else uses this colour.'}
        </p>
      </FormSection>

      {personId && (
        <label className="flex items-center gap-2 text-hm-caption text-hm-ink-2">
          <input
            type="radio"
            name="family"
            checked={draft.isFamily}
            onChange={() => onChange({ isFamily: true })}
          />
          Use this badge for unmatched household events
        </label>
      )}
    </div>
  )
}
