"use client"

import * as React from "react"
import { Drawer } from "@base-ui/react/drawer"

import { cn } from "../lib/utils"

/**
 * The Coming up sheet, built on Base UI's Drawer (non-modal, swipe down to close).
 *
 *   <ComingUp open={open} onOpenChange={setOpen}>
 *     <ComingUpSwipeArea>             // swipe up here to open (the After that list)
 *       <ComingUpHandle />             // tap to open
 *       ...rows
 *     </ComingUpSwipeArea>
 *     <ComingUpSheet container={columnRef} action={<Button variant="text">Week</Button>}>
 *       ...compact rows
 *     </ComingUpSheet>
 *   </ComingUp>
 *
 * On kiosk and desktop pass the events column as `container` so the sheet rises inside it. On mobile leave it
 * out and it becomes a normal bottom sheet over the screen. It closes on its own after `autoCloseMs` (45 s).
 */
function ComingUp({
  open,
  onOpenChange,
  autoCloseMs = 45_000,
  children,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  autoCloseMs?: number
  children: React.ReactNode
}) {
  React.useEffect(() => {
    if (!open || !autoCloseMs) return
    const id = setTimeout(() => onOpenChange(false), autoCloseMs)
    return () => clearTimeout(id)
  }, [open, autoCloseMs, onOpenChange])

  return (
    <Drawer.Root open={open} onOpenChange={(next) => onOpenChange(next)} modal={false} swipeDirection="down">
      {children}
    </Drawer.Root>
  )
}

function ComingUpSwipeArea({ className, ...props }: Drawer.SwipeArea.Props) {
  return <Drawer.SwipeArea data-slot="coming-up-swipe-area" className={cn("flex touch-none flex-col", className)} {...props} />
}

const handleClass =
  "grid h-[max(32px,calc(44px*var(--hm-density)))] w-[max(120px,calc(200px*var(--hm-density)))] flex-none cursor-pointer place-items-center self-center border-0 bg-transparent p-0 before:block before:h-[max(4px,calc(6px*var(--hm-density)))] before:w-[max(36px,calc(72px*var(--hm-density)))] before:rounded-full before:bg-hm-line before:content-[''] focus-visible:outline-3 focus-visible:outline-hm-accent"

/** The grab bar. Outside the sheet it opens it; the sheet has its own that closes it. */
function ComingUpHandle({ className, ...props }: Drawer.Trigger.Props) {
  return <Drawer.Trigger aria-label="Show more upcoming events" className={cn(handleClass, className)} {...props} />
}

function ComingUpSheet({
  container,
  title = "Coming up",
  hint = "Swipe down to close",
  action,
  className,
  viewportClassName,
  children,
}: {
  /** The element the sheet rises inside (kiosk and desktop: the events column). Leave out on mobile. */
  container?: React.RefObject<HTMLElement | null> | HTMLElement | null
  title?: React.ReactNode
  hint?: React.ReactNode
  action?: React.ReactNode
  className?: string
  /** Position of the sheet's frame inside the container (default: fills it). The kiosk prototype uses "-left-6 right-2". */
  viewportClassName?: string
  children: React.ReactNode
}) {
  return (
    <Drawer.Portal container={container}>
      <Drawer.Viewport className={cn(container ? "absolute" : "fixed", "inset-0 z-50 flex items-end", viewportClassName)}>
        <Drawer.Popup
          data-slot="coming-up-sheet"
          className={cn(
            "[--bleed:32px] relative -mb-(--bleed) flex max-h-[calc(100%-8px+var(--bleed))] w-full flex-col overflow-hidden rounded-t-hm-28 bg-hm-surface px-hm-32 pt-1 pb-[calc(24px*var(--hm-density)+var(--bleed))] text-hm-ink shadow-hm-sheet outline-none",
            "[transform:translateY(var(--drawer-swipe-movement-y))] transition-transform duration-340 ease-hm-sheet data-swiping:select-none data-swiping:duration-0",
            "data-starting-style:[transform:translateY(calc(100%+80px))] data-ending-style:[transform:translateY(calc(100%+80px))]",
            className,
          )}
        >
          <Drawer.Close aria-label="Collapse the list" className={handleClass} />
          <div className="mt-1 mb-hm-12 flex items-center justify-between gap-hm-16">
            <div className="flex flex-col gap-1">
              <Drawer.Title className="m-0 text-hm-title-sm">{title}</Drawer.Title>
              {hint && <Drawer.Description className="m-0 text-hm-caption text-hm-ink-3">{hint}</Drawer.Description>}
            </div>
            {action}
          </div>
          <Drawer.Content className="flex min-h-0 flex-col overflow-y-auto">{children}</Drawer.Content>
        </Drawer.Popup>
      </Drawer.Viewport>
    </Drawer.Portal>
  )
}

export { ComingUp, ComingUpSwipeArea, ComingUpHandle, ComingUpSheet }
