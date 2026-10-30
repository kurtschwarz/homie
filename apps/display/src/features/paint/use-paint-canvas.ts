import * as React from 'react'
import { createPaintHistory, type PaintRecord } from './paint-history'
import { createStrokePlayer, playStroke, type StrokePlayer, type StrokeSpec } from './stroke-player'
import type { Point } from './tools'

const MAX_UNDO = 20
const BACKING_LONG_EDGE = 2048

type ActiveStroke = {
  id: number
  player: StrokePlayer
  spec: StrokeSpec
  points: Point[]
}

function applyRecord(ctx: CanvasRenderingContext2D, record: PaintRecord) {
  if (record.kind === 'clear') {
    ctx.globalCompositeOperation = 'source-over'
    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height)
  } else {
    playStroke(ctx, record, record.points)
  }
}

export function usePaintCanvas(brush: StrokeSpec) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const brushRef = React.useRef(brush)
  brushRef.current = brush
  const baselineRef = React.useRef<HTMLCanvasElement | null>(null)
  const stroke = React.useRef<ActiveStroke | null>(null)
  const [dirty, setDirty] = React.useState(false)
  const [canUndo, setCanUndo] = React.useState(false)

  const history = React.useMemo(
    () =>
      createPaintHistory({
        maxDepth: MAX_UNDO,
        bake: (record) => {
          const ctx = baselineRef.current?.getContext('2d')
          if (ctx) applyRecord(ctx, record)
        },
        restore: () => {
          const baseline = baselineRef.current
          const ctx = canvasRef.current?.getContext('2d')
          if (!baseline || !ctx) return
          ctx.globalCompositeOperation = 'source-over'
          ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height)
          ctx.drawImage(baseline, 0, 0)
        },
        replay: (record) => {
          const ctx = canvasRef.current?.getContext('2d')
          if (ctx) applyRecord(ctx, record)
        },
      }),
    [],
  )

  const sync = React.useCallback(() => {
    setCanUndo(history.canUndo)
    setDirty(history.hasMarks)
  }, [history])

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const baseline = document.createElement('canvas')
    baselineRef.current = baseline
    const resize = () => {
      const { width, height } = canvas.getBoundingClientRect()
      if (!width || !height) return
      const scale = BACKING_LONG_EDGE / Math.max(width, height)
      const w = Math.round(width * scale)
      const h = Math.round(height * scale)
      if (canvas.width === w && canvas.height === h && baseline.width === w) return
      canvas.width = w
      canvas.height = h
      baseline.width = w
      baseline.height = h
      stroke.current = null
      history.reset()
      sync()
    }
    resize()
    const observer = new ResizeObserver(resize)
    observer.observe(canvas)
    return () => {
      observer.disconnect()
      baselineRef.current = null
      stroke.current = null
      history.reset()
    }
  }, [history, sync])

  const pointAt = (canvas: HTMLCanvasElement, clientX: number, clientY: number): Point => {
    const rect = canvas.getBoundingClientRect()
    return {
      x: ((clientX - rect.left) / rect.width) * canvas.width,
      y: ((clientY - rect.top) / rect.height) * canvas.height,
    }
  }

  const onPointerDown = (event: React.PointerEvent<HTMLCanvasElement>) => {
    if (stroke.current) return
    const canvas = event.currentTarget
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    canvas.setPointerCapture(event.pointerId)

    const ratio = canvas.width / canvas.getBoundingClientRect().width
    const { tool, color, size, smooth } = brushRef.current
    const spec = { tool, color, size: size * ratio, smooth }
    const player = createStrokePlayer(ctx, spec)
    const point = pointAt(canvas, event.clientX, event.clientY)
    stroke.current = { id: event.pointerId, player, spec, points: [point] }
    player.begin(point)
    if (tool !== 'eraser') setDirty(true)
  }

  const onPointerMove = (event: React.PointerEvent<HTMLCanvasElement>) => {
    const current = stroke.current
    if (!current || current.id !== event.pointerId) return
    const canvas = event.currentTarget
    const events = event.nativeEvent.getCoalescedEvents?.() ?? [event.nativeEvent]
    for (const sample of events) {
      const point = pointAt(canvas, sample.clientX, sample.clientY)
      current.points.push(point)
      current.player.move(point)
    }
  }

  const onPointerEnd = (event: React.PointerEvent<HTMLCanvasElement>) => {
    const current = stroke.current
    if (current?.id !== event.pointerId) return
    stroke.current = null
    current.player.end()
    history.push({ kind: 'stroke', ...current.spec, points: current.points })
    sync()
  }

  const undo = () => {
    if (stroke.current) return
    if (history.undo()) sync()
  }

  const clear = () => {
    const canvas = canvasRef.current
    const ctx = canvas?.getContext('2d')
    if (!canvas || !ctx || stroke.current || !history.hasMarks) return
    ctx.globalCompositeOperation = 'source-over'
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    history.push({ kind: 'clear' })
    sync()
  }

  return {
    canvasRef,
    dirty,
    canUndo,
    undo,
    clear,
    handlers: {
      onPointerDown,
      onPointerMove,
      onPointerUp: onPointerEnd,
      onPointerCancel: onPointerEnd,
      onLostPointerCapture: onPointerEnd,
    },
  }
}
