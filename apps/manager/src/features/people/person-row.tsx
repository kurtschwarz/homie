import * as React from "react";
import { Button } from "@homie/ui/components/button";
import { PersonBadge } from "@homie/ui/components/person-badge";
import { DEFAULT_SWATCH, SWATCHES, type Swatch } from "@homie/ui/lib/people";
import type { ManagerPerson, PersonInput } from "../../lib/use-settings";

/** One person's row. Without a `person` it is the add form instead, and clears itself on submit. */
export function PersonRow({ person, familyChecked = false, onSave, onSetFamily, onDelete }: {
  person?: ManagerPerson; familyChecked?: boolean; onSave: (input: PersonInput) => void; onSetFamily?: () => void; onDelete?: () => void;
}) {
  const [name, setName] = React.useState(person?.name ?? "");
  const [emoji, setEmoji] = React.useState(person?.emoji ?? "");
  const [swatch, setSwatch] = React.useState<Swatch>(person?.swatch ?? DEFAULT_SWATCH);
  const [confirmingDelete, setConfirmingDelete] = React.useState(false);
  const dirty = person ? name !== person.name || emoji !== person.emoji || swatch !== person.swatch : name.trim().length > 0;

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    onSave({ id: person?.id, name: name.trim(), emoji, swatch, isFamily: person?.isFamily ?? false });
    if (!person) { setName(""); setEmoji(""); setSwatch(DEFAULT_SWATCH); }
  }

  return <form onSubmit={submit} className="flex flex-wrap items-center gap-3 rounded-hm-16 border border-hm-line bg-hm-surface p-4">
    <PersonBadge person={{ id: person?.id ?? "preview", name: name || "?", emoji, swatch }} size="md" />
    <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" aria-label="Name" className="min-w-36 flex-1 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink" />
    <input value={emoji} onChange={(e) => setEmoji(e.target.value)} placeholder="Emoji" aria-label="Emoji" className="w-20 rounded-hm-control border border-hm-line bg-hm-surface px-3 py-2 text-hm-body text-hm-ink" />
    <div className="flex gap-1" role="group" aria-label="Swatch">
      {SWATCHES.map((s) => <button key={s} type="button" aria-label={s} aria-pressed={swatch === s} onClick={() => setSwatch(s)} className="size-6 rounded-full border-2" style={{ backgroundColor: `var(--hm-swatch-${s})`, borderColor: swatch === s ? "var(--hm-accent)" : "transparent" }} />)}
    </div>
    {person && onSetFamily && <label className="flex items-center gap-1 text-hm-caption text-hm-ink-2"><input type="radio" name="family" checked={familyChecked} onChange={onSetFamily} />Family</label>}
    <Button type="submit" variant="text" disabled={!dirty}>{person ? "Save" : "Add person"}</Button>
    {person && onDelete && (confirmingDelete ? <span className="flex flex-wrap items-center gap-2 text-hm-caption text-hm-ink-2">Delete {person.name} and their tag mappings?<Button type="button" variant="text" onClick={onDelete}>Confirm</Button><Button type="button" variant="text" onClick={() => setConfirmingDelete(false)}>Cancel</Button></span> : <Button type="button" variant="text" onClick={() => setConfirmingDelete(true)}>Delete</Button>)}
  </form>;
}
