export type Rgb = [number, number, number]

export type FloodFillOptions = {
  x: number
  y: number
  color: Rgb
}

const TOLERANCE = 32
const EDGE_RADIUS = 2

export function hexToRgb(hex: string): Rgb {
  const value = Number.parseInt(hex.replace('#', ''), 16)
  return [(value >> 16) & 255, (value >> 8) & 255, value & 255]
}

type Bounds = { x0: number; y0: number; x1: number; y1: number }

/** Marks every pixel within `radius` (Chebyshev) of a set pixel, inside `bounds`. */
function dilate(
  src: Uint8Array,
  width: number,
  radius: number,
  { x0, y0, x1, y1 }: Bounds,
): Uint8Array {
  const horizontal = new Uint8Array(src.length)
  const out = new Uint8Array(src.length)

  for (let y = y0; y <= y1; y++) {
    const base = y * width
    let distance = radius + 1
    for (let x = x0; x <= x1; x++) {
      const index = base + x
      distance = src[index] ? 0 : distance + 1
      if (distance <= radius) horizontal[index] = 1
    }
    distance = radius + 1
    for (let x = x1; x >= x0; x--) {
      const index = base + x
      distance = src[index] ? 0 : distance + 1
      if (distance <= radius) horizontal[index] = 1
    }
  }

  for (let x = x0; x <= x1; x++) {
    let distance = radius + 1
    for (let y = y0; y <= y1; y++) {
      const index = y * width + x
      distance = horizontal[index] ? 0 : distance + 1
      if (distance <= radius) out[index] = 1
    }
    distance = radius + 1
    for (let y = y1; y >= y0; y--) {
      const index = y * width + x
      distance = horizontal[index] ? 0 : distance + 1
      if (distance <= radius) out[index] = 1
    }
  }

  return out
}

/**
 * Floods the region of similar pixels around (x, y) with an opaque colour, in place.
 * Antialiased edge pixels keep their own alpha as coverage over the fill. Returns how
 * many pixels changed.
 */
export function floodFill(
  data: Uint8ClampedArray,
  width: number,
  height: number,
  { x, y, color }: FloodFillOptions,
): number {
  if (x < 0 || y < 0 || x >= width || y >= height) return 0

  const seed = (y * width + x) * 4
  const sr = data[seed] as number
  const sg = data[seed + 1] as number
  const sb = data[seed + 2] as number
  const sa = data[seed + 3] as number

  // Which pixels count as part of the region, read once: recomputing the colour test per
  // pixel costs more than the flood itself.
  const total = width * height
  const matching = new Uint8Array(total)
  for (let pixel = 0, offset = 0; pixel < total; pixel++, offset += 4) {
    if (
      Math.abs((data[offset] as number) - sr) <= TOLERANCE &&
      Math.abs((data[offset + 1] as number) - sg) <= TOLERANCE &&
      Math.abs((data[offset + 2] as number) - sb) <= TOLERANCE &&
      Math.abs((data[offset + 3] as number) - sa) <= TOLERANCE
    ) {
      matching[pixel] = 1
    }
  }

  const region = new Uint8Array(total)
  const open = (pixel: number) => region[pixel] === 0 && matching[pixel] === 1

  let minX = x
  let maxX = x
  let minY = y
  let maxY = y

  const stack = [y * width + x]
  while (stack.length > 0) {
    const start = stack.pop() as number
    if (!open(start)) continue
    const row = Math.floor(start / width)
    const rowStart = row * width
    let left = start
    while (left > rowStart && open(left - 1)) left--
    let right = start
    while (right < rowStart + width - 1 && open(right + 1)) right++
    for (let p = left; p <= right; p++) region[p] = 1

    if (row < minY) minY = row
    if (row > maxY) maxY = row
    if (left - rowStart < minX) minX = left - rowStart
    if (right - rowStart > maxX) maxX = right - rowStart

    for (const next of [row - 1, row + 1]) {
      if (next < 0 || next >= height) continue
      const nextStart = next * width
      let inRun = false
      for (let p = left; p <= right; p++) {
        const isOpen = open(nextStart + (p - rowStart))
        if (isOpen && !inRun) stack.push(nextStart + (p - rowStart))
        inRun = isOpen
      }
    }
  }

  const [fr, fg, fb] = color
  let changed = 0

  // Everything from here on only touches the filled region and the edge around it, so it
  // works inside that box rather than walking the whole canvas. The padding covers the
  // edge ring plus the window strokeNear reads around it.
  const pad = 2 * EDGE_RADIUS + 1
  const bounds: Bounds = {
    x0: Math.max(0, minX - pad),
    y0: Math.max(0, minY - pad),
    x1: Math.min(width - 1, maxX + pad),
    y1: Math.min(height - 1, maxY + pad),
  }
  const ring = dilate(region, width, EDGE_RADIUS, bounds)
  const rowOffset = bounds.y0 * width * 4
  const original = data.slice(rowOffset, (bounds.y1 + 1) * width * 4)

  // Finds the stroke colour behind an edge pixel: the nearby pixel furthest from the seed.
  const strokeNear = (px: number, py: number): [number, number, number, number] | null => {
    let best: [number, number, number, number] | null = null
    for (let dy = -EDGE_RADIUS - 1; dy <= EDGE_RADIUS + 1; dy++) {
      for (let dx = -EDGE_RADIUS - 1; dx <= EDGE_RADIUS + 1; dx++) {
        const nx = px + dx
        const ny = py + dy
        if (nx < bounds.x0 || ny < bounds.y0 || nx > bounds.x1 || ny > bounds.y1) continue
        const o = (ny * width + nx) * 4 - rowOffset
        const dr = (original[o] as number) - sr
        const dg = (original[o + 1] as number) - sg
        const db = (original[o + 2] as number) - sb
        const distance = dr * dr + dg * dg + db * db
        if (!best || distance > best[3]) best = [dr, dg, db, distance]
      }
    }
    return best && best[3] > 0 ? best : null
  }

  for (let py = bounds.y0; py <= bounds.y1; py++) {
    for (let px = bounds.x0; px <= bounds.x1; px++) {
      const pixel = py * width + px
      // Background well clear of the fill keeps its colour.
      if (region[pixel] === 0 && ring[pixel] === 0) continue

      const offset = pixel * 4
      let r: number
      let g: number
      let b: number

      if (region[pixel]) {
        r = fr
        g = fg
        b = fb
      } else {
        const alpha = data[offset + 3] as number
        if (alpha < 255) {
          const coverage = alpha / 255
          r = Math.round((data[offset] as number) * coverage + fr * (1 - coverage))
          g = Math.round((data[offset + 1] as number) * coverage + fg * (1 - coverage))
          b = Math.round((data[offset + 2] as number) * coverage + fb * (1 - coverage))
        } else {
          // Opaque edge pixel: a blend of the seed colour and a stroke colour. Swap the
          // seed share of the blend for the fill colour.
          if (sa !== 255) continue
          const stroke = strokeNear(px, py)
          if (!stroke) continue
          const [dr, dg, db, distance] = stroke
          const along =
            (((data[offset] as number) - sr) * dr +
              ((data[offset + 1] as number) - sg) * dg +
              ((data[offset + 2] as number) - sb) * db) /
            distance
          const seedShare = 1 - Math.min(1, Math.max(0, along))
          r = Math.round((data[offset] as number) + (fr - sr) * seedShare)
          g = Math.round((data[offset + 1] as number) + (fg - sg) * seedShare)
          b = Math.round((data[offset + 2] as number) + (fb - sb) * seedShare)
        }
      }

      if (
        data[offset] !== r ||
        data[offset + 1] !== g ||
        data[offset + 2] !== b ||
        data[offset + 3] !== 255
      ) {
        data[offset] = r
        data[offset + 1] = g
        data[offset + 2] = b
        data[offset + 3] = 255
        changed++
      }
    }
  }

  return changed
}
