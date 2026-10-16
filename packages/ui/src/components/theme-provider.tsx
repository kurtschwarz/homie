"use client"

import * as React from "react"

import { useHomieTheme, type HomieThemeSetting } from "../hooks/use-homie-theme"
import { cn } from "../lib/utils"

const THEME_CLASSES = ["light", "dark", "night"]

export type HomieSurface = "auto" | "desktop" | "mobile" | "kiosk"

/**
 * Applies the Homie theme and surface.
 *
 * theme:   "auto" follows the clock (Day, Evening from 7 pm, Night 9 pm–6:30 am), or force one.
 * surface: "auto" lets the viewport decide (mobile under 768px, desktop above). Use "kiosk" on the kitchen
 *          display; it's opt-in because a big desktop monitor shouldn't get kiosk-sized type.
 *
 * With scope="document" (default) it sets data-theme, the theme class and data-surface on <html>, so stock shadcn
 * components and portals follow along. With scope="local" it wraps its children instead.
 */
export function HomieThemeProvider({
  theme = "auto",
  surface = "auto",
  scope = "document",
  className,
  children,
}: {
  theme?: HomieThemeSetting
  surface?: HomieSurface
  scope?: "document" | "local"
  className?: string
  children: React.ReactNode
}) {
  const mode = useHomieTheme(theme)

  React.useEffect(() => {
    if (scope !== "document") return
    const root = document.documentElement
    root.dataset.theme = mode
    root.classList.remove(...THEME_CLASSES)
    root.classList.add(mode)
    if (surface === "auto") delete root.dataset.surface
    else root.dataset.surface = surface
  }, [mode, surface, scope])

  if (scope === "document") return <>{children}</>
  return (
    <div
      data-theme={mode}
      data-surface={surface === "auto" ? undefined : surface}
      className={cn(mode, "bg-hm-bg text-hm-ink transition-colors duration-1500", className)}
    >
      {children}
    </div>
  )
}
