/**
 * Who an event is for. Colour is a curated swatch chosen per person, not a
 * free hex, so each theme can supply a value already tuned for contrast
 * against `--hm-on-person`. See swatchVars.
 */
export const SWATCHES = [
  'clay',
  'amber',
  'moss',
  'pine',
  'teal',
  'sky',
  'iris',
  'orchid',
  'plum',
  'slate',
] as const
export type Swatch = (typeof SWATCHES)[number]
export const DEFAULT_SWATCH: Swatch = 'slate'

export interface Person {
  id: string
  /** Shown in legends, titles and detail lines. */
  name: string
  /** The person's emoji badge. Falls back to the first letter of their name. */
  emoji?: string
  swatch: Swatch
}

/**
 * Sets --p (colour) and --p-tint for a person; badge and card classes read them.
 * Static strings so Tailwind sees them. The tint is derived from --p with
 * color-mix() rather than stored separately, so a new swatch needs one value.
 */
export const swatchVars: Record<Swatch, string> = {
  clay: '[--p:var(--hm-swatch-clay)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  amber:
    '[--p:var(--hm-swatch-amber)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  moss: '[--p:var(--hm-swatch-moss)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  pine: '[--p:var(--hm-swatch-pine)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  teal: '[--p:var(--hm-swatch-teal)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  sky: '[--p:var(--hm-swatch-sky)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  iris: '[--p:var(--hm-swatch-iris)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  orchid:
    '[--p:var(--hm-swatch-orchid)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  plum: '[--p:var(--hm-swatch-plum)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
  slate:
    '[--p:var(--hm-swatch-slate)] [--p-tint:color-mix(in_srgb,var(--p)_var(--hm-tint-strength),transparent)]',
}

/** Falls back to the default swatch rather than leaving --p undefined for an unknown value. */
export const varsFor = (person: { swatch: string }) =>
  swatchVars[person.swatch as Swatch] ?? swatchVars[DEFAULT_SWATCH]

export function badgeFor(person: Person) {
  return person.emoji?.trim() || person.name.charAt(0).toUpperCase()
}

/** "Lana", "Lana & Elina", "Oskar, Lana & Elina" */
export function joinNames(people: Person[]) {
  if (people.length < 2) return people[0]?.name ?? ''
  return `${people
    .slice(0, -1)
    .map((p) => p.name)
    .join(', ')} & ${people[people.length - 1]!.name}`
}
