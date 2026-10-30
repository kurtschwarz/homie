import { type Point, paintCurve, paintSegment, startStroke, type Tool } from './tools'

export type StrokeSpec = { tool: Tool; color: string; size: number; smooth: boolean }

export type StrokePlayer = {
  begin: (point: Point) => void
  move: (point: Point) => void
  end: () => void
}

export function createStrokePlayer(ctx: CanvasRenderingContext2D, spec: StrokeSpec): StrokePlayer {
  const state = startStroke(spec.tool, spec.color, spec.size)
  let last: Point = { x: 0, y: 0 }
  let mid: Point = last

  return {
    begin(point) {
      last = point
      mid = point
      paintSegment(ctx, state, point, { x: point.x + 0.01, y: point.y })
    },
    move(point) {
      if (spec.smooth) {
        const next = { x: (last.x + point.x) / 2, y: (last.y + point.y) / 2 }
        paintCurve(ctx, state, mid, last, next)
        mid = next
      } else {
        paintSegment(ctx, state, last, point)
      }
      last = point
    },
    end() {
      if (spec.smooth) paintSegment(ctx, state, mid, last)
    },
  }
}

export function playStroke(ctx: CanvasRenderingContext2D, spec: StrokeSpec, points: Point[]) {
  const [first, ...rest] = points
  if (!first) return
  const player = createStrokePlayer(ctx, spec)
  player.begin(first)
  for (const point of rest) player.move(point)
  player.end()
}
