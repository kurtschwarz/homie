import { isHex } from './colour'

export const PALETTE_SIZE = 8

const PALETTE_KEY = 'homie.display.paint.palette'
const LAST_MIX_KEY = 'homie.display.paint.lastMix'

type Store = Pick<Storage, 'getItem' | 'setItem'>

function defaultStore(): Store | null {
  try {
    return globalThis.localStorage ?? null
  } catch {
    return null
  }
}

export function loadPalette(store: Store | null = defaultStore()): string[] {
  try {
    const parsed: unknown = JSON.parse(store?.getItem(PALETTE_KEY) ?? '[]')
    if (!Array.isArray(parsed)) return []
    const colours: string[] = []
    for (const entry of parsed) {
      if (!isHex(entry)) continue
      const hex = entry.toLowerCase()
      if (!colours.includes(hex)) colours.push(hex)
    }
    return colours.slice(0, PALETTE_SIZE)
  } catch {
    return []
  }
}

export function savePalette(palette: string[], store: Store | null = defaultStore()) {
  try {
    store?.setItem(PALETTE_KEY, JSON.stringify(palette))
  } catch {}
}

export function loadLastMix(store: Store | null = defaultStore()): string | null {
  try {
    const value = store?.getItem(LAST_MIX_KEY)
    return isHex(value) ? value.toLowerCase() : null
  } catch {
    return null
  }
}

export function saveLastMix(hex: string, store: Store | null = defaultStore()) {
  try {
    store?.setItem(LAST_MIX_KEY, hex)
  } catch {}
}

/** Adds a colour unless it is already saved or the palette is full; `index` is where the colour sits afterwards. */
export function addColour(palette: string[], hex: string): { palette: string[]; index: number } {
  const existing = palette.indexOf(hex)
  if (existing !== -1) return { palette, index: existing }
  if (palette.length >= PALETTE_SIZE) return { palette, index: -1 }
  return { palette: [...palette, hex], index: palette.length }
}

export function removeColour(palette: string[], index: number): string[] {
  return palette.filter((_, position) => position !== index)
}
