'use client'

import { DotLottieReact } from '@lottiefiles/dotlottie-react'
import type * as React from 'react'
import { cn } from '../lib/utils'

/**
 * The looping Homie mascot animation used by first-run setup screens.
 *
 * The default size is deliberately a single unprefixed utility. A responsive
 * default (`sm:size-64`) sits in a different tailwind-merge group from a
 * caller's `size-*`, so it survives the merge and then wins the cascade at
 * every width above the `sm` breakpoint — leaving callers unable to resize it.
 * Callers that want a breakpoint-aware size pass their own prefixed class.
 */
export function HomieAnimation({
  src,
  className,
  ...props
}: React.ComponentProps<'div'> & { src: string }) {
  return (
    <div
      data-slot="homie-animation"
      className={cn('size-56', className)}
      role="img"
      aria-label="Homie"
      {...props}
    >
      <DotLottieReact
        src={src}
        autoplay
        loop
        renderConfig={{
          devicePixelRatio: typeof window === 'undefined' ? 1 : window.devicePixelRatio,
        }}
      />
    </div>
  )
}
