import type * as React from 'react'

import { badgeFor, type Person, type Swatch, varsFor } from '../lib/people'
import { cn } from '../lib/utils'

const SIZES = {
  lg: 'size-hm-badge-lg text-[length:calc(var(--hm-badge-lg)*0.58)]',
  md: 'size-hm-badge text-[length:calc(var(--hm-badge)*0.56)]',
  sm: 'size-hm-badge-sm text-[length:calc(var(--hm-badge-sm)*0.55)] shadow-[inset_0_0_0_max(1px,calc(0.125rem*var(--hm-density)))_var(--p)]',
  chip: 'size-hm-36 text-[length:calc(1.25rem*var(--hm-density))]',
  legend:
    'size-hm-40 text-[length:calc(1.375rem*var(--hm-density))] shadow-[inset_0_0_0_calc(0.1875rem*var(--hm-density))_var(--p)]',
}

/**
 * Round emoji badge in a person's colour. Colour always travels with the emoji and the name (the title),
 * never alone. Sizes follow the surface: lg (`badge-lg`, beside the headline), md (`badge`, event rows),
 * sm (`badge-sm`, week cards), chip and legend. `ring` adds the inset ring used in home list rows.
 */
export function PersonBadge({
  person,
  size = 'md',
  ring = false,
  className,
  ...props
}: React.ComponentProps<'span'> & { person: Person; size?: keyof typeof SIZES; ring?: boolean }) {
  return (
    <span
      data-slot="person-badge"
      title={person.name}
      className={cn(
        varsFor(person),
        'grid flex-none place-items-center rounded-full bg-(--p) leading-none text-hm-on-person',
        'font-[Apple_Color_Emoji,Segoe_UI_Emoji,Noto_Color_Emoji,var(--font-sans)] [filter:var(--hm-emoji-filter)]',
        SIZES[size],
        ring && 'shadow-[inset_0_0_0_calc(0.1875rem*var(--hm-density))_var(--p)]',
        className,
      )}
      {...props}
    >
      {badgeFor(person)}
    </span>
  )
}

/** A row of badges for an event that is for several people. */
export function PersonBadges({
  people,
  size = 'md',
  ring,
  className,
}: {
  people: Person[]
  size?: keyof typeof SIZES
  ring?: boolean
  className?: string
}) {
  return (
    <span className={cn('flex flex-none', size === 'sm' ? 'gap-hm-4' : 'gap-hm-6', className)}>
      {people.map((p) => (
        <PersonBadge key={p.id} person={p} size={size} ring={ring} />
      ))}
    </span>
  )
}

/** Who's-who pill for the Week screen: badge plus name on the person's tint. */
export function PersonLegend({
  person,
  className,
  ...props
}: React.ComponentProps<'span'> & { person: Person }) {
  return (
    <span
      data-slot="person-legend"
      className={cn(
        varsFor(person),
        'inline-flex h-hm-52 items-center gap-hm-10 rounded-full bg-(--p-tint) pr-hm-20 pl-hm-8 text-hm-caption font-medium whitespace-nowrap text-hm-ink',
        className,
      )}
      {...props}
    >
      <PersonBadge person={person} size="legend" title={undefined} aria-hidden />
      {person.name}
    </span>
  )
}

export type { Person, Swatch }
