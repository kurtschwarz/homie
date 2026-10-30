'use client'

import { DotLottieReact } from '@lottiefiles/dotlottie-react'
import * as React from 'react'
import { cn } from '../lib/utils'

type DotLottie = Parameters<
  NonNullable<React.ComponentProps<typeof DotLottieReact>['dotLottieRefCallback']>
>[0] & {}

function hexToRgb(hex: string): [number, number, number] | null {
  const match = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex)
  if (!match) return null
  return [1, 2, 3].map((i) => Number.parseInt(match[i] as string, 16) / 255) as [
    number,
    number,
    number,
  ]
}

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
  colorSlots,
  className,
  ...props
}: React.ComponentProps<'div'> & {
  src: string
  /** Hex colours keyed by Lottie slot id; a missing or null entry keeps the animation's own colour. */
  colorSlots?: Record<string, string | null>
}) {
  const [player, setPlayer] = React.useState<DotLottie | null>(null)

  React.useEffect(() => {
    if (!player || !colorSlots) return
    const apply = () => {
      for (const [id, hex] of Object.entries(colorSlots)) {
        const rgb = hex ? hexToRgb(hex) : null
        if (rgb) player.setColorSlot(id, [...rgb, 1])
        else player.resetSlot(id)
      }
    }
    if (player.isLoaded) apply()
    player.addEventListener('load', apply)
    return () => player.removeEventListener('load', apply)
  }, [player, colorSlots])

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
        dotLottieRefCallback={setPlayer}
        autoplay
        loop
        renderConfig={{
          devicePixelRatio: typeof window === 'undefined' ? 1 : window.devicePixelRatio,
        }}
      />
    </div>
  )
}
