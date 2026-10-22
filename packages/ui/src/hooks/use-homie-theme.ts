"use client";

import { useNow } from "./use-now";

export type HomieTheme = "light" | "dark" | "night";
export type HomieThemeSetting = HomieTheme | "auto";

/** Auto schedule: Day until 7 pm, Evening from 7 pm, Night from 9 pm until 6:30 am. */
export function themeForTime(d: Date): HomieTheme {
  const hr = d.getHours() + d.getMinutes() / 60;
  return hr >= 21 || hr < 6.5 ? "night" : hr >= 19 ? "dark" : "light";
}

/** Resolves a theme setting to the theme to show now, following the clock when set to "auto". */
export function useHomieTheme(setting: HomieThemeSetting = "auto"): HomieTheme {
  const now = useNow(60_000);
  return setting === "auto" ? themeForTime(now) : setting;
}
