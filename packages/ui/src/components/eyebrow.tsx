import type * as React from 'react'

import { cn } from '../lib/utils'

/** Uppercase section label ("After that", "Next 12 hours", metric names) in ink-3. */
export function Eyebrow({
  className,
  as: Tag = 'h2',
  ...props
}: React.ComponentProps<'h2'> & { as?: 'h2' | 'h3' | 'span' | 'div' }) {
  return (
    <Tag
      data-slot="eyebrow"
      className={cn('m-0 text-hm-eyebrow uppercase text-hm-ink-3', className)}
      {...props}
    />
  )
}
