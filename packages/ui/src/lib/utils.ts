import { clsx, type ClassValue } from "clsx"
import { extendTailwindMerge } from "tailwind-merge"

import { HM_RADII, HM_SHADOWS, HM_SIZES, HM_TEXT_STYLES } from "./type-scale"

const hm = (names: readonly string[]) => names.map((n) => `hm-${n}`)

// Teach tailwind-merge the Homie scales, so text-hm-clock is a font size (not a colour),
// and h-hm-touch, rounded-hm-24 and shadow-hm-sheet override their defaults correctly.
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      text: hm(HM_TEXT_STYLES),
      spacing: hm(HM_SIZES),
      radius: hm(HM_RADII),
      shadow: hm(HM_SHADOWS),
    },
  },
})

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
