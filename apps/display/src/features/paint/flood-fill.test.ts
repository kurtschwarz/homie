import { describe, expect, it } from 'vitest'
import { floodFill, hexToRgb } from './flood-fill'

const RED: [number, number, number] = [255, 0, 0]

function blank(width: number, height: number) {
  return new Uint8ClampedArray(width * height * 4)
}

function setInk(data: Uint8ClampedArray, width: number, x: number, y: number) {
  const offset = (y * width + x) * 4
  data[offset] = 0
  data[offset + 1] = 0
  data[offset + 2] = 0
  data[offset + 3] = 255
}

function isFilled(data: Uint8ClampedArray, width: number, x: number, y: number) {
  const offset = (y * width + x) * 4
  return data[offset] === 255 && data[offset + 1] === 0 && data[offset + 3] === 255
}

function box(data: Uint8ClampedArray, width: number, left: number, top: number, size: number) {
  for (let i = 0; i < size; i++) {
    setInk(data, width, left + i, top)
    setInk(data, width, left + i, top + size - 1)
    setInk(data, width, left, top + i)
    setInk(data, width, left + size - 1, top + i)
  }
}

describe('floodFill', () => {
  it('fills a blank canvas edge to edge', () => {
    const data = blank(20, 20)

    const changed = floodFill(data, 20, 20, { x: 5, y: 5, color: RED })

    expect(changed).toBe(400)
    expect(isFilled(data, 20, 0, 0)).toBe(true)
    expect(isFilled(data, 20, 19, 19)).toBe(true)
  })

  it('fills only the inside of a closed shape', () => {
    const data = blank(40, 40)
    box(data, 40, 10, 10, 20)

    floodFill(data, 40, 40, { x: 20, y: 20, color: RED })

    expect(isFilled(data, 40, 20, 20)).toBe(true)
    expect(isFilled(data, 40, 11, 11)).toBe(true)
    expect(isFilled(data, 40, 2, 2)).toBe(false)
    expect(isFilled(data, 40, 10, 10)).toBe(false)
  })

  it('leaks out of a shape with a gap', () => {
    const data = blank(40, 40)
    box(data, 40, 10, 10, 20)
    for (let y = 15; y < 19; y++) data.fill(0, (y * 40 + 10) * 4, (y * 40 + 10) * 4 + 4)

    floodFill(data, 40, 40, { x: 20, y: 20, color: RED })

    expect(isFilled(data, 40, 2, 2)).toBe(true)
  })

  it('composites the fill beneath antialiased edges instead of eating them', () => {
    const data = blank(20, 5)
    for (let y = 0; y < 5; y++) {
      setInk(data, 20, 10, y)
      const offset = (y * 20 + 9) * 4
      data[offset + 3] = 128
    }

    floodFill(data, 20, 5, { x: 2, y: 2, color: RED })

    const edge = (2 * 20 + 9) * 4
    expect(data[edge + 3]).toBe(255)
    expect(data[edge]).toBeGreaterThan(100)
    expect(data[edge]).toBeLessThan(255)
    expect(isFilled(data, 20, 10, 2)).toBe(false)
  })

  it('composites the edge of a small shape far from the canvas origin', () => {
    const size = 120
    const data = new Uint8ClampedArray(size * size * 4)
    const yellow: [number, number, number] = [237, 198, 93]
    // An opaque dark canvas with a yellow box, its border half blended into the background.
    for (let i = 0; i < size * size; i++) {
      const offset = i * 4
      data[offset] = 22
      data[offset + 1] = 23
      data[offset + 2] = 26
      data[offset + 3] = 255
    }
    for (let py = 80; py < 110; py++) {
      for (let px = 80; px < 110; px++) {
        const onBorder = px === 80 || px === 109 || py === 80 || py === 109
        const offset = (py * size + px) * 4
        const share = onBorder ? 0.5 : 1
        data[offset] = Math.round(22 * (1 - share) + (yellow[0] as number) * share)
        data[offset + 1] = Math.round(23 * (1 - share) + (yellow[1] as number) * share)
        data[offset + 2] = Math.round(26 * (1 - share) + (yellow[2] as number) * share)
      }
    }

    floodFill(data, size, size, { x: 95, y: 95, color: RED })

    const inside = (95 * size + 95) * 4
    expect([data[inside], data[inside + 1], data[inside + 2]]).toEqual([255, 0, 0])
    // The half-blended border takes the fill in place of the yellow it was blended from.
    const border = (95 * size + 80) * 4
    expect(data[border]).toBeGreaterThan(100)
    expect(data[border + 1]).toBeLessThan(60)
    // Background well away from the shape is untouched.
    const far = (10 * size + 10) * 4
    expect([data[far], data[far + 1], data[far + 2]]).toEqual([22, 23, 26])
  })

  it('leaves no dark halo when refilling an opaque background with the stroke colour', () => {
    const width = 20
    const data = new Uint8ClampedArray(width * 7 * 4)
    const yellow = [237, 198, 93]
    const column = [0, 0.25, 0.75, 1, 0.75, 0.25, 0]
    for (let y = 0; y < 7; y++) {
      for (let x = 0; x < width; x++) {
        const share = y < 6 && x >= 8 && x < 15 ? (column[x - 8] as number) : 0
        const offset = (y * width + x) * 4
        for (let c = 0; c < 3; c++) data[offset + c] = Math.round(share * (yellow[c] as number))
        data[offset + 3] = 255
      }
    }

    floodFill(data, width, 7, {
      x: 1,
      y: 3,
      color: yellow as [number, number, number],
    })

    for (let x = 0; x < width; x++) {
      const offset = (3 * width + x) * 4
      expect([data[offset], data[offset + 1], data[offset + 2]]).toEqual(yellow)
    }
  })

  it('changes nothing when the region already has the colour', () => {
    const data = blank(10, 10)
    floodFill(data, 10, 10, { x: 1, y: 1, color: RED })

    const changed = floodFill(data, 10, 10, { x: 1, y: 1, color: RED })

    expect(changed).toBe(0)
  })
})

describe('hexToRgb', () => {
  it('parses a hex colour', () => {
    expect(hexToRgb('#d24a35')).toEqual([210, 74, 53])
  })
})
