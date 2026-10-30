import type { StrokeSpec } from './stroke-player'
import type { Point } from './tools'

export type PaintRecord = ({ kind: 'stroke'; points: Point[] } & StrokeSpec) | { kind: 'clear' }

export type PaintHistoryDeps = {
  maxDepth: number
  /** Draws a record onto the baseline. Called once per record, when it is evicted. */
  bake: (record: PaintRecord) => void
  /** Puts the live canvas back to the baseline. */
  restore: () => void
  /** Draws a record onto the live canvas. */
  replay: (record: PaintRecord) => void
}

function leavesMarks(record: PaintRecord, before: boolean) {
  if (record.kind === 'clear') return false
  return record.tool === 'eraser' ? before : true
}

export function createPaintHistory(deps: PaintHistoryDeps) {
  let records: PaintRecord[] = []
  let baselineHasMarks = false

  return {
    push(record: PaintRecord) {
      records.push(record)
      if (records.length > deps.maxDepth) {
        const oldest = records.shift()
        if (oldest) {
          deps.bake(oldest)
          baselineHasMarks = leavesMarks(oldest, baselineHasMarks)
        }
      }
    },
    undo(): boolean {
      if (records.length === 0) return false
      records.pop()
      deps.restore()
      for (const record of records) deps.replay(record)
      return true
    },
    reset() {
      records = []
      baselineHasMarks = false
    },
    get canUndo() {
      return records.length > 0
    },
    get hasMarks() {
      return records.reduce((marks, record) => leavesMarks(record, marks), baselineHasMarks)
    },
  }
}
