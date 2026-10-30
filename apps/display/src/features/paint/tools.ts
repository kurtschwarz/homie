export type Tool = 'brush' | 'rainbow' | 'eraser'

export type Point = { x: number; y: number }

export type SizeName = 'small' | 'medium' | 'large'

export const TOOLS: Tool[] = ['brush', 'rainbow', 'eraser']

export const SIZE_PX: Record<SizeName, number> = { small: 6, medium: 14, large: 30 }

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
