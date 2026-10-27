'use client'

import { Paintbrush } from 'lucide-react'
import type * as React from 'react'

import { cn } from '../lib/utils'
import { Button } from './button'

/** Floating pill that opens painting. Sits over the photo panel's bottom-right corner; hidden at Night. */
export function PaintButton({
  className,
  ...props
}: Omit<React.ComponentProps<typeof Button>, 'variant'>) {
  return (
    <Button
      data-slot="paint-button"
      variant="outline"
      className={cn(
        'absolute right-hm-28 bottom-hm-28 rounded-full border-0 shadow-hm-float night:hidden',
        className,
      )}
      {...props}
    >
      <Paintbrush aria-hidden className="text-hm-accent" strokeWidth={2} />
      Paint
    </Button>
  )
}
