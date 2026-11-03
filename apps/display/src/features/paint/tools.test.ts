import { describe, expect, it } from 'vitest'
import {
  BRUSH_MAX,
  BRUSH_MIN,
  clampBrushWidth,
  effectiveBrushWidth,
  nextSegmentStyle,
  paintCurve,
  snapBrushWidth,
  startStroke,
} from './tools'

describe('snapBrushWidth', () => {
  it('snaps to a target within 3', () => {
    expect(snapBrushWidth(8)).toBe(6)
    expect(snapBrushWidth(24)).toBe(22)
    expect(snapBrushWidth(43)).toBe(40)
    expect(snapBrushWidth(61)).toBe(64)
  })

  it('leaves in-between widths alone', () => {
    expect(snapBrushWidth(15)).toBe(15)
    expect(snapBrushWidth(30)).toBe(30)
    expect(snapBrushWidth(44)).toBe(44)
  })

  it('lets the smaller target win a tie', () => {
    expect(snapBrushWidth(8)).toBe(6)
  })

  it('clamps to the brush range', () => {
    expect(snapBrushWidth(-10)).toBe(BRUSH_MIN)
    expect(snapBrushWidth(500)).toBe(BRUSH_MAX)
    expect(clampBrushWidth(3.6)).toBe(4)
  })
})

describe('effectiveBrushWidth', () => {
  it('caps the eraser so it draws at most 100', () => {
    expect(effectiveBrushWidth('eraser', 80)).toBe(50)
    expect(effectiveBrushWidth('eraser', 20)).toBe(20)
  })

  it('leaves other tools alone', () => {
    expect(effectiveBrushWidth('brush', 80)).toBe(80)
  })
})

describe('paintCurve', () => {
  it('draws a quadratic curve through the control point', () => {
    const calls: unknown[][] = []
    const ctx = {
      beginPath: () => {},
      moveTo: (...args: unknown[]) => calls.push(['moveTo', ...args]),
      quadraticCurveTo: (...args: unknown[]) => calls.push(['quadraticCurveTo', ...args]),
      stroke: () => calls.push(['stroke']),
    } as unknown as CanvasRenderingContext2D

    paintCurve(
      ctx,
      startStroke('brush', '#16171a', 14),
      { x: 0, y: 0 },
      { x: 10, y: 10 },
      { x: 20, y: 0 },
    )

    expect(calls).toEqual([['moveTo', 0, 0], ['quadraticCurveTo', 10, 10, 20, 0], ['stroke']])
  })
})

const origin = { x: 0, y: 0 }

describe('nextSegmentStyle', () => {
  it('erases with destination-out at twice the size', () => {
    const style = nextSegmentStyle(startStroke('eraser', '#ffffff', 14), origin, { x: 10, y: 0 })

    expect(style.composite).toBe('destination-out')
    expect(style.width).toBe(28)
  })

  it('advances the rainbow hue with the distance travelled', () => {
    const state = startStroke('rainbow', '#000000', 14)

    const first = nextSegmentStyle(state, origin, { x: 10, y: 0 })
    const second = nextSegmentStyle(state, { x: 10, y: 0 }, { x: 20, y: 0 })

    expect(first.color).not.toBe(second.color)
    expect(second.color).toBe('hsl(12 80% 55%)')
  })

  it('wraps the rainbow hue past a full turn', () => {
    const state = startStroke('rainbow', '#000000', 14)

    const style = nextSegmentStyle(state, origin, { x: 700, y: 0 })

    expect(style.color).toBe('hsl(60 80% 55%)')
  })

  it('thins the brush as the pointer speeds up', () => {
    const slow = startStroke('brush', '#16171a', 14)
    const fast = startStroke('brush', '#16171a', 14)

    const slowStyle = nextSegmentStyle(slow, origin, { x: 1, y: 0 })
    const fastStyle = nextSegmentStyle(fast, origin, { x: 80, y: 0 })

    expect(fastStyle.width).toBeLessThan(slowStyle.width)
  })

  it('keeps the brush colour and draws over existing ink', () => {
    const style = nextSegmentStyle(startStroke('brush', '#d24a35', 14), origin, { x: 5, y: 0 })

    expect(style.color).toBe('#d24a35')
    expect(style.composite).toBe('source-over')
  })
})
