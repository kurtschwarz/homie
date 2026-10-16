/**
 * Who an event is for. Homie has three people slots plus Family and Grown-ups,
 * each with a colour and tint tuned for all three themes.
 */
export type PersonKey = "a" | "b" | "c" | "family" | "grownups"

export interface Person {
  key: PersonKey
  /** Shown in legends, titles and detail lines. */
  name: string
  /** The person's emoji badge. Falls back to the first letter of their name. */
  emoji?: string
}

/** Sets --p (colour) and --p-tint for a person; badge and card classes read them. Static strings so Tailwind sees them. */
export const personVars: Record<PersonKey, string> = {
  a: "[--p:var(--hm-person-a)] [--p-tint:var(--hm-person-a-tint)]",
  b: "[--p:var(--hm-person-b)] [--p-tint:var(--hm-person-b-tint)]",
  c: "[--p:var(--hm-person-c)] [--p-tint:var(--hm-person-c-tint)]",
  family: "[--p:var(--hm-person-family)] [--p-tint:var(--hm-person-family-tint)]",
  grownups: "[--p:var(--hm-person-grownups)] [--p-tint:var(--hm-person-grownups-tint)]",
}

export function badgeFor(person: Person) {
  return person.emoji?.trim() || person.name.charAt(0).toUpperCase()
}

/** "Lana", "Lana & Elina", "Oskar, Lana & Elina" */
export function joinNames(people: Person[]) {
  if (people.length < 2) return people[0]?.name ?? ""
  return `${people.slice(0, -1).map((p) => p.name).join(", ")} & ${people[people.length - 1]!.name}`
}
