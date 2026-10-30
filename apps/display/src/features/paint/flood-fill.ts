export type Rgb = [number, number, number]

export type FloodFillOptions = {
  x: number
  y: number
  color: Rgb
  closeGaps: boolean
}

const TOLERANCE = 32
const GAP_RADIUS = 4
const EDGE_RADIUS = 2

export function hexToRgb(hex: string): Rgb {
  const value = Number.parseInt(hex.replace('#', ''), 16)
  return [(value >> 16) & 255, (value >> 8) & 255, value & 255]
}

/** Marks every pixel within `radius` (Chebyshev) of a set pixel. */
function dilate(src: Uint8Array, width: number, height: number, radius: number): Uint8Array {
  const horizontal = new Uint8Array(src.length)
  const out = new Uint8Array(src.length)

  const sweep = (
    from: Uint8Array,
    to: Uint8Array,
    lines: number,
    length: number,
    lineStride: number,
    stride: number,
  ) => {
    for (let line = 0; line < lines; line++) {
      const base = line * lineStride
      let distance = radius + 1
      for (let i = 0; i < length; i++) {
        const index = base + i * stride
        distance = from[index] ? 0 : distance + 1
        if (distance <= radius) to[index] = 1
      }
      distance = radius + 1
      for (let i = length - 1; i >= 0; i--) {
        const index = base + i * stride
        distance = from[index] ? 0 : distance + 1
        if (distance <= radius) to[index] = 1
      }
    }
  }

  sweep(src, horizontal, height, width, width, 1)
  sweep(horizontal, out, width, height, 1, width)
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
  { x, y, color, closeGaps }: FloodFillOptions,
): number {
  if (x < 0 || y < 0 || x >= width || y >= height) return 0

  const seed = (y * width + x) * 4
  const sr = data[seed] as number
  const sg = data[seed + 1] as number
  const sb = data[seed + 2] as number
  const sa = data[seed + 3] as number

  const matches = (pixel: number) => {
    const offset = pixel * 4
    return (
      Math.abs((data[offset] as number) - sr) <= TOLERANCE &&
      Math.abs((data[offset + 1] as number) - sg) <= TOLERANCE &&
      Math.abs((data[offset + 2] as number) - sb) <= TOLERANCE &&
      Math.abs((data[offset + 3] as number) - sa) <= TOLERANCE
    )
  }

  const total = width * height
  let blocked: Uint8Array | null = null
  if (closeGaps) {
    const ink = new Uint8Array(total)
    for (let i = 0; i < total; i++) if (!matches(i)) ink[i] = 1
    blocked = dilate(ink, width, height, GAP_RADIUS)
    blocked[y * width + x] = 0
  }

  let region: Uint8Array = new Uint8Array(total)
  const open = (pixel: number) =>
    region[pixel] === 0 && (blocked ? blocked[pixel] === 0 : true) && matches(pixel)

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

  if (blocked) {
    const grown = dilate(region, width, height, GAP_RADIUS)
    for (let i = 0; i < total; i++)
      if (grown[i] && !region[i] && matches(i)) grown[i] = 1
      else grown[i] = region[i] as number
    region = grown
  }

  const [fr, fg, fb] = color
  let changed = 0
  const ring = dilate(region, width, height, EDGE_RADIUS)
  const original = data.slice()

  // Finds the stroke colour behind an edge pixel: the nearby pixel furthest from the seed.
  const strokeNear = (px: number, py: number): [number, number, number, number] | null => {
    let best: [number, number, number, number] | null = null
    for (let dy = -EDGE_RADIUS - 1; dy <= EDGE_RADIUS + 1; dy++) {
      for (let dx = -EDGE_RADIUS - 1; dx <= EDGE_RADIUS + 1; dx++) {
        const nx = px + dx
        const ny = py + dy
        if (nx < 0 || ny < 0 || nx >= width || ny >= height) continue
        const o = (ny * width + nx) * 4
        const dr = (original[o] as number) - sr
        const dg = (original[o + 1] as number) - sg
        const db = (original[o + 2] as number) - sb
        const distance = dr * dr + dg * dg + db * db
        if (!best || distance > best[3]) best = [dr, dg, db, distance]
      }
    }
    return best && best[3] > 0 ? best : null
  }

  for (let py = 0; py < height; py++) {
    for (let px = 0; px < width; px++) {
      const pixel = py * width + px
      const offset = pixel * 4
      let r: number
      let g: number
      let b: number

      if (region[pixel]) {
        r = fr
        g = fg
        b = fb
      } else {
        if (!ring[pixel]) continue
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
