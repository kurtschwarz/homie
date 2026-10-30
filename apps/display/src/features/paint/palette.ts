import { isHex } from './colour'

export const PALETTE_SIZE = 8

const PALETTE_KEY = 'homie.display.paint.palette'
const LAST_MIX_KEY = 'homie.display.paint.lastMix'

type Store = Pick<Storage, 'getItem' | 'setItem'>

/**
 * Returns the browser's `localStorage`, or `null` when it is missing or
 * access throws (for example when storage is blocked).
 */
function defaultStore(): Store | null {
  try {
    return globalThis.localStorage ?? null
  } catch {
    return null
  }
}

/**
 * Reads the saved colour palette.
 *
 * Stored data is untrusted: anything that is not a JSON array is ignored,
 * entries that are not `#rrggbb` hex colours are dropped, duplicates are
 * removed, hex values are lowercased and the result is capped at
 * {@link PALETTE_SIZE}.
 *
 * @param store - Storage to read from. Defaults to `localStorage`.
 * @returns The saved colours, or an empty palette when nothing usable is stored.
 */
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

/**
 * Writes the colour palette to storage. Failures, such as a full or blocked
 * store, are ignored so painting is never interrupted.
 *
 * @param palette - Hex colours to save.
 * @param store - Storage to write to. Defaults to `localStorage`.
 */
export function savePalette(palette: string[], store: Store | null = defaultStore()) {
  try {
    store?.setItem(PALETTE_KEY, JSON.stringify(palette))
  } catch {}
}

/**
 * Reads the colour last chosen with **Use it** in the mix panel.
 *
 * @param store - Storage to read from. Defaults to `localStorage`.
 * @returns The lowercased hex colour, or `null` when nothing valid is stored.
 */
export function loadLastMix(store: Store | null = defaultStore()): string | null {
  try {
    const value = store?.getItem(LAST_MIX_KEY)
    return isHex(value) ? value.toLowerCase() : null
  } catch {
    return null
  }
}

/**
 * Remembers the last mixed colour. Failures are ignored.
 *
 * @param hex - The colour to remember, as `#rrggbb`.
 * @param store - Storage to write to. Defaults to `localStorage`.
 */
export function saveLastMix(hex: string, store: Store | null = defaultStore()) {
  try {
    store?.setItem(LAST_MIX_KEY, hex)
  } catch {}
}

/**
 * Adds a colour to the palette without mutating it.
 *
 * A colour that is already saved is not added again, and nothing is added
 * once the palette holds {@link PALETTE_SIZE} colours.
 *
 * @param palette - The current palette.
 * @param hex - The colour to add, as `#rrggbb`.
 * @returns The resulting palette and the colour's slot in it. `index` is
 * the existing slot for a duplicate, or `-1` when the palette is full and
 * the colour was not added.
 */
export function addColour(palette: string[], hex: string): { palette: string[]; index: number } {
  const existing = palette.indexOf(hex)
  if (existing !== -1) return { palette, index: existing }
  if (palette.length >= PALETTE_SIZE) return { palette, index: -1 }
  return { palette: [...palette, hex], index: palette.length }
}

/**
 * Removes the colour at a slot without mutating the palette.
 *
 * @param palette - The current palette.
 * @param index - Slot to remove.
 * @returns A new palette without that slot. Later colours shift up.
 */
export function removeColour(palette: string[], index: number): string[] {
  return palette.filter((_, position) => position !== index)
}
