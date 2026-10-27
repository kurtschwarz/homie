import * as React from 'react'
import {
  type Point,
  paintCurve,
  paintSegment,
  type StrokeState,
  startStroke,
  type Tool,
} from './tools'

const MAX_UNDO = 10

type Brush = { tool: Tool; color: string; size: number; smooth: boolean }

export function usePaintCanvas(brush: Brush) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const brushRef = React.useRef(brush)
  brushRef.current = brush
  const history = React.useRef<ImageData[]>([])
  const stroke = React.useRef<{
    id: number
    state: StrokeState
    last: Point
    mid: Point
    smooth: boolean
  } | null>(null)
  const [dirty, setDirty] = React.useState(false)
  const [canUndo, setCanUndo] = React.useState(false)

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const resize = () => {
      const { width, height } = canvas.getBoundingClientRect()
      const ratio = window.devicePixelRatio || 1
      if (!width || !height) return
      canvas.width = Math.round(width * ratio)
      canvas.height = Math.round(height * ratio)
      history.current = []
      setCanUndo(false)
      setDirty(false)
    }
    resize()
    const observer = new ResizeObserver(resize)
    observer.observe(canvas)
    return () => observer.disconnect()
  }, [])

  const pointAt = (event: React.PointerEvent<HTMLCanvasElement>): Point => {
    const canvas = event.currentTarget
    const rect = canvas.getBoundingClientRect()
    return {
      x: ((event.clientX - rect.left) / rect.width) * canvas.width,
      y: ((event.clientY - rect.top) / rect.height) * canvas.height,
    }
  }

  const onPointerDown = (event: React.PointerEvent<HTMLCanvasElement>) => {
    if (stroke.current) return
    const canvas = event.currentTarget
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    canvas.setPointerCapture(event.pointerId)

    history.current.push(ctx.getImageData(0, 0, canvas.width, canvas.height))
    if (history.current.length > MAX_UNDO) history.current.shift()
    setCanUndo(true)
    setDirty(true)

    const ratio = canvas.width / canvas.getBoundingClientRect().width
    const { tool, color, size, smooth } = brushRef.current
    const state = startStroke(tool, color, size * ratio)
    const point = pointAt(event)
    stroke.current = { id: event.pointerId, state, last: point, mid: point, smooth }
    paintSegment(ctx, state, point, { x: point.x + 0.01, y: point.y })
  }

  const onPointerMove = (event: React.PointerEvent<HTMLCanvasElement>) => {
    const current = stroke.current
    if (!current || current.id !== event.pointerId) return
    const ctx = event.currentTarget.getContext('2d')
    if (!ctx) return
    const events = event.nativeEvent.getCoalescedEvents?.() ?? [event.nativeEvent]
    const rect = event.currentTarget.getBoundingClientRect()
    for (const sample of events) {
      const point = {
        x: ((sample.clientX - rect.left) / rect.width) * event.currentTarget.width,
        y: ((sample.clientY - rect.top) / rect.height) * event.currentTarget.height,
      }
      if (current.smooth) {
        const mid = { x: (current.last.x + point.x) / 2, y: (current.last.y + point.y) / 2 }
        paintCurve(ctx, current.state, current.mid, current.last, mid)
        current.mid = mid
      } else {
        paintSegment(ctx, current.state, current.last, point)
      }
      current.last = point
    }
  }

  const onPointerEnd = (event: React.PointerEvent<HTMLCanvasElement>) => {
    const current = stroke.current
    if (current?.id !== event.pointerId) return
    stroke.current = null
    const ctx = event.currentTarget.getContext('2d')
    if (ctx && current.smooth) paintSegment(ctx, current.state, current.mid, current.last)
  }

  const undo = () => {
    const canvas = canvasRef.current
    const ctx = canvas?.getContext('2d')
    const previous = history.current.pop()
    if (!ctx || !previous) return
    ctx.globalCompositeOperation = 'source-over'
    ctx.putImageData(previous, 0, 0)
    setCanUndo(history.current.length > 0)
    if (history.current.length === 0) setDirty(false)
  }

  const clear = () => {
    const canvas = canvasRef.current
    const ctx = canvas?.getContext('2d')
    if (!canvas || !ctx) return
    history.current = []
    ctx.globalCompositeOperation = 'source-over'
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    setCanUndo(false)
    setDirty(false)
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
