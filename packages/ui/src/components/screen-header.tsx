'use client'

import { ChevronLeft } from 'lucide-react'
import type * as React from 'react'

import { cn } from '../lib/utils'
import { Button } from './button'
import { Clock } from './clock'

/**
 * Top bar of a full screen (Week, Weather): Home button, title, a middle slot, and the small clock.
 * On mobile the middle slot drops to its own line and the clock hides (the phone shows the time).
 * On kiosk, pair with useIdle() to drift back home after 60 s without a touch.
 */
export function ScreenHeader({
  title,
  onHome,
  homeLabel = 'Home',
  children,
  className,
  ...props
}: Omit<React.ComponentProps<'header'>, 'title'> & {
  title: React.ReactNode
  onHome: () => void
  homeLabel?: string
}) {
  return (
    <header
      data-slot="screen-header"
      className={cn('flex items-center gap-hm-32 mobile:flex-wrap mobile:gap-y-hm-12', className)}
      {...props}
    >
      <Button onClick={onHome}>
        <ChevronLeft aria-hidden />
        {homeLabel}
      </Button>
      <h1 className="m-0 text-hm-title-lg mobile:flex-1">{title}</h1>
      {children && (
        <div className="min-w-0 flex-1 mobile:order-3 mobile:basis-full">{children}</div>
      )}
      <Clock size="sm" className={cn('mobile:hidden', !children && 'ml-auto')} />
    </header>
  )
}
