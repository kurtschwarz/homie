export type Hsv = { h: number; s: number; v: number }

const HEX = /^#[0-9a-f]{6}$/i

export function isHex(value: unknown): value is string {
  return typeof value === 'string' && HEX.test(value)
}

export function hsvToHex({ h, s, v }: Hsv): string {
  const channel = (n: number) => {
    const k = (n + h / 60) % 6
    return v - v * s * Math.max(0, Math.min(k, 4 - k, 1))
  }
  const byte = (n: number) =>
    Math.round(channel(n) * 255)
      .toString(16)
      .padStart(2, '0')
  return `#${byte(5)}${byte(3)}${byte(1)}`
}

export function readableInk(hex: string): string {
  const r = Number.parseInt(hex.slice(1, 3), 16)
  const g = Number.parseInt(hex.slice(3, 5), 16)
  const b = Number.parseInt(hex.slice(5, 7), 16)
  return 0.299 * r + 0.587 * g + 0.114 * b > 160 ? '#16171a' : '#ffffff'
}

export function hexToHsv(hex: string): Hsv {
  const r = Number.parseInt(hex.slice(1, 3), 16) / 255
  const g = Number.parseInt(hex.slice(3, 5), 16) / 255
  const b = Number.parseInt(hex.slice(5, 7), 16) / 255
  const max = Math.max(r, g, b)
  const delta = max - Math.min(r, g, b)
  let h = 0
  if (delta > 0) {
    if (max === r) h = ((g - b) / delta) % 6
    else if (max === g) h = (b - r) / delta + 2
    else h = (r - g) / delta + 4
    h = (h * 60 + 360) % 360
  }
  return { h, s: max === 0 ? 0 : delta / max, v: max }
}
