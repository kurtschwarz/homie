import { describe, expect, it } from 'vitest'
import { hexToHsv, hsvToHex } from './colour'
import {
  addColour,
  loadLastMix,
  loadPalette,
  PALETTE_SIZE,
  removeColour,
  saveLastMix,
  savePalette,
} from './palette'

function memoryStore(initial: Record<string, string> = {}) {
  const data = { ...initial }
  return {
    data,
    getItem: (key: string) => data[key] ?? null,
    setItem: (key: string, value: string) => {
      data[key] = value
    },
  }
}

const PALETTE_KEY = 'homie.display.paint.palette'

describe('loadPalette', () => {
  it('is empty when nothing is stored', () => {
    expect(loadPalette(memoryStore())).toEqual([])
  })

  it('is empty when there is no storage at all', () => {
    expect(loadPalette(null)).toEqual([])
  })

  it('round-trips a saved palette', () => {
    const store = memoryStore()
    savePalette(['#112233', '#abcdef'], store)

    expect(loadPalette(store)).toEqual(['#112233', '#abcdef'])
  })

  it.each([
    ['invalid JSON', '{nope'],
    ['an object', '{"a":1}'],
    ['a string', '"#112233"'],
    ['null', 'null'],
  ])('falls back to empty for %s', (_, raw) => {
    expect(loadPalette(memoryStore({ [PALETTE_KEY]: raw }))).toEqual([])
  })

  it('drops entries that are not hex colours and removes duplicates', () => {
    const raw = JSON.stringify(['#112233', 'red', 5, null, '#112233', '#ABCDEF'])

    expect(loadPalette(memoryStore({ [PALETTE_KEY]: raw }))).toEqual(['#112233', '#abcdef'])
  })

  it('keeps at most eight colours', () => {
    const many = Array.from({ length: 12 }, (_, i) => `#0000${i.toString(16).padStart(2, '0')}`)

    expect(loadPalette(memoryStore({ [PALETTE_KEY]: JSON.stringify(many) }))).toHaveLength(
      PALETTE_SIZE,
    )
  })

  it('survives a store that throws', () => {
    const broken = {
      getItem: () => {
        throw new Error('denied')
      },
      setItem: () => {
        throw new Error('denied')
      },
    }

    expect(loadPalette(broken)).toEqual([])
    expect(() => savePalette(['#112233'], broken)).not.toThrow()
  })
})

describe('last mix', () => {
  it('round-trips', () => {
    const store = memoryStore()
    saveLastMix('#112233', store)

    expect(loadLastMix(store)).toBe('#112233')
  })

  it('ignores unexpected stored data', () => {
    expect(loadLastMix(memoryStore({ 'homie.display.paint.lastMix': 'blue' }))).toBeNull()
    expect(loadLastMix(null)).toBeNull()
  })
})

describe('addColour', () => {
  it('appends a new colour', () => {
    expect(addColour(['#111111'], '#222222')).toEqual({
      palette: ['#111111', '#222222'],
      index: 1,
    })
  })

  it('points at the existing slot instead of duplicating', () => {
    const palette = ['#111111', '#222222']

    expect(addColour(palette, '#111111')).toEqual({ palette, index: 0 })
  })

  it('refuses to grow past eight', () => {
    const full = Array.from({ length: PALETTE_SIZE }, (_, i) => `#00000${i}`)

    expect(addColour(full, '#ffffff')).toEqual({ palette: full, index: -1 })
  })
})

describe('removeColour', () => {
  it('removes the colour at the index', () => {
    expect(removeColour(['#111111', '#222222', '#333333'], 1)).toEqual(['#111111', '#333333'])
  })
})

describe('hsv conversion', () => {
  it('converts primaries', () => {
    expect(hsvToHex({ h: 0, s: 1, v: 1 })).toBe('#ff0000')
    expect(hsvToHex({ h: 120, s: 1, v: 1 })).toBe('#00ff00')
    expect(hsvToHex({ h: 240, s: 1, v: 1 })).toBe('#0000ff')
    expect(hsvToHex({ h: 0, s: 0, v: 1 })).toBe('#ffffff')
  })

  it('round-trips a hex colour', () => {
    expect(hsvToHex(hexToHsv('#5590db'))).toBe('#5590db')
  })
})
