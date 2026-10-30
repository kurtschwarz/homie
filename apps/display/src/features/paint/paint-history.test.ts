import { describe, expect, it } from 'vitest'
import { createPaintHistory, type PaintRecord } from './paint-history'

const stroke = (tool: 'brush' | 'eraser' = 'brush'): PaintRecord => ({
  kind: 'stroke',
  tool,
  color: '#000000',
  size: 10,
  smooth: true,
  points: [{ x: 1, y: 1 }],
})
const clear: PaintRecord = { kind: 'clear' }

function setup(maxDepth = 3) {
  const baked: PaintRecord[] = []
  const replayed: PaintRecord[] = []
  let restores = 0
  const history = createPaintHistory({
    maxDepth,
    bake: (record) => baked.push(record),
    restore: () => {
      restores++
      replayed.length = 0
    },
    replay: (record) => replayed.push(record),
  })
  return { history, baked, replayed, restores: () => restores }
}

describe('paint history', () => {
  it('bakes each evicted record exactly once, oldest first', () => {
    const { history, baked } = setup(2)
    const records = [stroke(), stroke(), stroke(), stroke()]
    for (const record of records) history.push(record)
    expect(baked).toEqual([records[0], records[1]])
  })

  it('undo restores the baseline and replays the survivors', () => {
    const { history, replayed, restores } = setup()
    const [a, b, c] = [stroke(), stroke(), stroke()]
    for (const record of [a, b, c]) history.push(record)
    expect(history.undo()).toBe(true)
    expect(restores()).toBe(1)
    expect(replayed).toEqual([a, b])
  })

  it('refuses to undo once records are exhausted, including baked ones', () => {
    const { history, restores } = setup(1)
    history.push(stroke())
    history.push(stroke())
    expect(history.undo()).toBe(true)
    expect(history.canUndo).toBe(false)
    expect(history.undo()).toBe(false)
    expect(restores()).toBe(1)
  })

  it('tracks marks through undo back to empty', () => {
    const { history } = setup()
    expect(history.hasMarks).toBe(false)
    history.push(stroke())
    expect(history.hasMarks).toBe(true)
    history.undo()
    expect(history.hasMarks).toBe(false)
  })

  it('clear removes marks and undoing it brings them back', () => {
    const { history } = setup()
    history.push(stroke())
    history.push(clear)
    expect(history.hasMarks).toBe(false)
    history.undo()
    expect(history.hasMarks).toBe(true)
  })

  it('keeps marks after the marking stroke has been baked', () => {
    const { history } = setup(1)
    history.push(stroke())
    history.push(stroke('eraser'))
    expect(history.hasMarks).toBe(true)
  })

  it('has no marks after a baked clear', () => {
    const { history } = setup(1)
    history.push(stroke())
    history.push(clear)
    history.push(stroke('eraser'))
    expect(history.hasMarks).toBe(false)
  })

  it('reset drops records and marks', () => {
    const { history } = setup()
    history.push(stroke())
    history.reset()
    expect(history.canUndo).toBe(false)
    expect(history.hasMarks).toBe(false)
  })
})
