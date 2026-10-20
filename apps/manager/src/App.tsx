import * as React from "react";
import { Button } from "@homie/ui/components/button";
import { PersonBadge } from "@homie/ui/components/person-badge";
import { DEFAULT_SWATCH, SWATCHES, type Swatch } from "@homie/ui/lib/people";
import { useHealth } from "./useHealth";
import { useSettings, type ManagerPerson, type PersonInput, type TagKind, type TagMapping } from "./useSettings";

export function App() {
  const health = useHealth();
  const settings = useSettings();
  const [token, setToken] = React.useState(localStorage.getItem("homie-auth-token") ?? "");

  return (
    <main className="mx-auto flex max-w-5xl flex-col gap-hm-32 p-hm-screen text-hm-ink">
      <header className="flex flex-col gap-hm-8">
        <h1 className="m-0 text-hm-title-lg">Homie Manager</h1>
        <p className="m-0 text-hm-body text-hm-ink-2">
          {health.kind === "connecting" && "connecting to api…"}
          {health.kind === "error" && `api unreachable: ${health.message} — retrying…`}
          {health.kind === "ok" && (
            <>
              <span className="text-hm-accent">{health.health.status}</span>
              {` · v${health.health.version} · up ${health.health.uptimeSeconds}s`}
            </>
          )}
        </p>
      </header>

      <form className="flex flex-wrap items-center gap-hm-8" onSubmit={(event) => {
        event.preventDefault();
        localStorage.setItem("homie-auth-token", token.trim());
        window.location.reload();
      }}>
        <label htmlFor="api-token" className="text-hm-body">API token</label>
        <input id="api-token" type="password" autoComplete="current-password" value={token}
          onChange={(event) => setToken(event.target.value)} placeholder="Paste token from API log"
          className="min-w-0 flex-1 rounded-hm-control border-2 border-hm-line bg-hm-surface px-hm-12 py-hm-8 text-hm-body text-hm-ink" />
        <Button type="submit">Save token</Button>
      </form>

      {settings.state.kind === "loading" && <p className="m-0 text-hm-body text-hm-ink-2">Loading settings…</p>}
      {settings.state.kind === "error" && (
        <p className="m-0 text-hm-body text-hm-accent">Settings unavailable: {settings.state.message}</p>
      )}
      {settings.state.kind === "ready" && (
        <>
          {settings.state.error && <p className="m-0 text-hm-body text-hm-accent">{settings.state.error}</p>}
          <PeopleSection
            people={settings.state.settings.people}
            onSave={settings.savePerson}
            onDelete={settings.deletePerson}
          />
          <TagMappingSection
            tagMappings={settings.state.settings.tagMappings}
            calendarTags={settings.state.settings.calendarTags}
            people={settings.state.settings.people}
            onSave={settings.saveTagMapping}
            onDelete={settings.deleteTagMapping}
            onRefresh={settings.refreshEvents}
            refreshing={settings.state.saving}
          />
        </>
      )}
    </main>
  );
}

function PeopleSection({
  people,
  onSave,
  onDelete,
}: {
  people: ManagerPerson[];
  onSave: (input: PersonInput) => void;
  onDelete: (id: string) => void;
}) {
  return (
    <section className="flex flex-col gap-hm-16">
      <h2 className="m-0 text-hm-title-sm">People</h2>
      {people.length === 0 && (
        <p className="m-0 text-hm-body text-hm-ink-2">
          No one is set up yet. Add the first person below and mark them as family — events tagged{" "}
          <strong>Homie</strong> attribute to whoever holds that flag.
        </p>
      )}
      <div className="flex flex-col gap-hm-12">
        {people.map((person) => (
          <PersonRow
            key={person.id}
            person={person}
            familyChecked={person.isFamily}
            onSave={onSave}
            onSetFamily={() => onSave({ id: person.id, name: person.name, emoji: person.emoji, swatch: person.swatch, isFamily: true })}
            onDelete={() => onDelete(person.id)}
          />
        ))}
        <PersonRow onSave={onSave} />
      </div>
    </section>
  );
}

function PersonRow({
  person,
  familyChecked = false,
  onSave,
  onSetFamily,
  onDelete,
}: {
  person?: ManagerPerson;
  familyChecked?: boolean;
  onSave: (input: PersonInput) => void;
  onSetFamily?: () => void;
  onDelete?: () => void;
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
    if (!person) {
      setName("");
      setEmoji("");
      setSwatch(DEFAULT_SWATCH);
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-wrap items-center gap-hm-16 rounded-hm-16 border-2 border-hm-line p-hm-16">
      <PersonBadge person={{ id: person?.id ?? "preview", name: name || "?", emoji, swatch }} size="md" />
      <input
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Name"
        aria-label="Name"
        className="min-w-0 flex-1 rounded-hm-control border-2 border-hm-line bg-hm-surface px-hm-12 py-hm-8 text-hm-body text-hm-ink"
      />
      <input
        value={emoji}
        onChange={(e) => setEmoji(e.target.value)}
        placeholder="Emoji"
        aria-label="Emoji"
        className="w-[4.5rem] rounded-hm-control border-2 border-hm-line bg-hm-surface px-hm-12 py-hm-8 text-hm-body text-hm-ink"
      />
      <div className="flex gap-hm-6" role="group" aria-label="Swatch">
        {SWATCHES.map((s) => (
          <button
            key={s}
            type="button"
            aria-label={s}
            aria-pressed={swatch === s}
            onClick={() => setSwatch(s)}
            className="size-hm-26 rounded-full border-2"
            style={{ backgroundColor: `var(--hm-swatch-${s})`, borderColor: swatch === s ? "var(--hm-accent)" : "transparent" }}
          />
        ))}
      </div>
      {person && onSetFamily && (
        <label className="flex items-center gap-hm-6 text-hm-caption text-hm-ink-2">
          <input type="radio" name="family" checked={familyChecked} onChange={onSetFamily} />
          Family
        </label>
      )}
      <Button type="submit" variant="text" disabled={!dirty}>
        {person ? "Save" : "Add person"}
      </Button>
      {person && onDelete && (
        confirmingDelete ? (
          <span className="flex items-center gap-hm-6 text-hm-caption text-hm-ink-2">
            Delete {person.name} and their tag mappings?
            <Button type="button" variant="text" onClick={onDelete}>
              Confirm
            </Button>
            <Button type="button" variant="text" onClick={() => setConfirmingDelete(false)}>
              Cancel
            </Button>
          </span>
        ) : (
          <Button type="button" variant="text" onClick={() => setConfirmingDelete(true)}>
            Delete
          </Button>
        )
      )}
    </form>
  );
}

function TagMappingSection({
  tagMappings,
  calendarTags,
  people,
  onSave,
  onDelete,
  onRefresh,
  refreshing,
}: {
  tagMappings: TagMapping[];
  calendarTags: string[];
  people: ManagerPerson[];
  onSave: (input: { tag: string; kind: TagKind; personId?: string }) => void;
  onDelete: (tag: string) => void;
  onRefresh: () => void;
  refreshing: boolean;
}) {
  const mappingByTag = new Map(tagMappings.map((m) => [m.tag, m]));
  const tags = [...new Set([...calendarTags, ...tagMappings.map((m) => m.tag)])].sort();
  const [newTag, setNewTag] = React.useState("");

  // calendarTags only covers the synced window, so a tag used outside it is
  // invisible here and cannot be mapped — including, on a fresh install,
  // every tag. Adding by hand is the only way to break that deadlock.
  const addTag = (e: React.FormEvent) => {
    e.preventDefault();
    const tag = newTag.trim();
    if (tag === "") return;
    onSave({ tag, kind: "INCLUDE" });
    setNewTag("");
  };

  return (
    <section className="flex flex-col gap-hm-16">
      <div className="flex items-center gap-hm-16">
        <h2 className="m-0 flex-1 text-hm-title-sm">Tag mapping</h2>
        <Button type="button" variant="text" onClick={onRefresh} disabled={refreshing}>
          {refreshing ? "Syncing…" : "Sync calendar"}
        </Button>
      </div>
      <p className="m-0 text-hm-body text-hm-ink-2">
        Map a calendar tag to a person, or to <strong>Show on Homie</strong> to bring the event in without
        attributing it to anyone in particular. Only tags on events inside the synced window are listed.
      </p>
      <div className="flex flex-col gap-hm-8">
        {tags.map((tag) => {
          const mapping = mappingByTag.get(tag);
          const stale = !calendarTags.includes(tag);
          return (
            <div key={tag} className="flex items-center gap-hm-16 rounded-hm-16 border-2 border-hm-line p-hm-12">
              <span className="min-w-0 flex-1 truncate text-hm-body text-hm-ink">
                {mapping?.label ?? tag}
                {stale && <span className="ml-hm-6 text-hm-caption text-hm-ink-3">(not seen recently)</span>}
              </span>
              <select
                aria-label={`Mapping for ${tag}`}
                value={mapping ? (mapping.kind === "INCLUDE" ? "include" : mapping.person?.id ?? "") : "unmapped"}
                onChange={(e) => {
                  const value = e.target.value;
                  if (value === "unmapped") {
                    if (mapping) onDelete(tag);
                    return;
                  }
                  if (value === "include") {
                    onSave({ tag: mapping?.label ?? tag, kind: "INCLUDE" });
                    return;
                  }
                  onSave({ tag: mapping?.label ?? tag, kind: "PERSON", personId: value });
                }}
                className="rounded-hm-control border-2 border-hm-line bg-hm-surface px-hm-12 py-hm-8 text-hm-body text-hm-ink"
              >
                <option value="unmapped">Unmapped</option>
                <option value="include">Show on Homie</option>
                {people.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
              {mapping && (
                <Button type="button" variant="text" onClick={() => onDelete(tag)}>
                  Remove
                </Button>
              )}
            </div>
          );
        })}
        {tags.length === 0 && (
          <p className="m-0 text-hm-body text-hm-ink-2">
            No tags seen in the calendar yet. Add one below if you know its name.
          </p>
        )}
      </div>
      <form onSubmit={addTag} className="flex flex-wrap items-center gap-hm-16 rounded-hm-16 border-2 border-hm-line p-hm-16">
        <input
          value={newTag}
          onChange={(e) => setNewTag(e.target.value)}
          placeholder="Tag name, as spelled in the calendar"
          aria-label="Tag name"
          className="min-w-0 flex-1 rounded-hm-control border-2 border-hm-line bg-hm-surface px-hm-12 py-hm-8 text-hm-body text-hm-ink"
        />
        <Button type="submit">Add tag</Button>
      </form>
    </section>
  );
}
