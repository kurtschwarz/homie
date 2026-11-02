export type Tool = 'brush' | 'rainbow' | 'eraser' | 'fill' | 'sticker'

export type Point = { x: number; y: number }

export const BRUSH_MIN = 2
export const BRUSH_MAX = 80
export const BRUSH_DEFAULT = 22
export const BRUSH_SNAPS = [6, 10, 22, 40, 64]

const SNAP_RANGE = 3
const ERASER_MAX_WIDTH = 100

export function clampBrushWidth(width: number) {
  return Math.min(BRUSH_MAX, Math.max(BRUSH_MIN, Math.round(width)))
}

/** Pulls a width onto the nearest snap target within reach; the smaller target wins a tie. */
export function snapBrushWidth(width: number) {
  const clamped = clampBrushWidth(width)
  let best = clamped
  let bestDistance = SNAP_RANGE + 1
  for (const target of BRUSH_SNAPS) {
    const distance = Math.abs(target - clamped)
    if (distance <= SNAP_RANGE && distance < bestDistance) {
      best = target
      bestDistance = distance
    }
  }
  return best
}

/** The eraser draws at twice the brush width, so its brush stops growing once it hits the cap. */
export function effectiveBrushWidth(tool: Tool, width: number) {
  return tool === 'eraser' ? Math.min(width, ERASER_MAX_WIDTH / 2) : width
}

export const COLOURS = [
  { name: 'Black', value: '#16171a' },
  { name: 'White', value: '#ffffff' },
  { name: 'Red', value: '#d24a35' },
  { name: 'Orange', value: '#e49340' },
  { name: 'Yellow', value: '#edc65d' },
  { name: 'Mint', value: '#70c092' },
  { name: 'Green', value: '#4f9d62' },
  { name: 'Blue', value: '#5590db' },
  { name: 'Purple', value: '#6a4abf' },
  { name: 'Pink', value: '#d4689c' },
  { name: 'Brown', value: '#835d37' },
]

export type StrokeState = {
  tool: Tool
  color: string
  size: number
  travelled: number
  width: number
}

export type SegmentStyle = {
  width: number
  color: string
  composite: GlobalCompositeOperation
}

const RAINBOW_DEGREES_PER_PX = 0.6

export function startStroke(tool: Tool, color: string, size: number): StrokeState {
  return { tool, color, size, travelled: 0, width: size }
}

/** Advances the stroke by one segment and returns how to draw it. */
export function nextSegmentStyle(state: StrokeState, from: Point, to: Point): SegmentStyle {
  const length = Math.hypot(to.x - from.x, to.y - from.y)
  state.travelled += length

  if (state.tool === 'eraser') {
    return { width: state.size * 2, color: '#000000', composite: 'destination-out' }
  }

  if (state.tool === 'rainbow') {
    const hue = (state.travelled * RAINBOW_DEGREES_PER_PX) % 360
    return { width: state.size, color: `hsl(${hue} 80% 55%)`, composite: 'source-over' }
  }

  // Fast strokes thin out and slow ones swell; easing keeps the width from jittering between events.
  const target = state.size * Math.min(1.2, Math.max(0.5, 1.2 - length / (state.size * 3)))
  state.width += (target - state.width) * 0.35
  return { width: state.width, color: state.color, composite: 'source-over' }
}

export function paintCurve(
  ctx: CanvasRenderingContext2D,
  state: StrokeState,
  from: Point,
  control: Point,
  to: Point,
) {
  const style = nextSegmentStyle(state, from, to)
  ctx.globalCompositeOperation = style.composite
  ctx.strokeStyle = style.color
  ctx.lineWidth = style.width
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.beginPath()
  ctx.moveTo(from.x, from.y)
  ctx.quadraticCurveTo(control.x, control.y, to.x, to.y)
  ctx.stroke()
}

export function paintSegment(
  ctx: CanvasRenderingContext2D,
  state: StrokeState,
  from: Point,
  to: Point,
) {
  const style = nextSegmentStyle(state, from, to)
  ctx.globalCompositeOperation = style.composite
  ctx.strokeStyle = style.color
  ctx.lineWidth = style.width
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.beginPath()
  ctx.moveTo(from.x, from.y)
  ctx.lineTo(to.x, to.y)
  ctx.stroke()
}
