import { graphql } from '@/shared/lib/graphql'
import type { DisplayModeKey, DisplaySchedule, DisplayStatus, WifiStatus } from '../types'

const SCHEDULE_FIELDS = `
  modes { mode startsAtMinutes brightness }
  powerSavingEnabled wakeForAlerts ddcBus
`

const STATUS_FIELDS = `mode brightness powered overrideActive overrideEndsAt`

export const displayApi = {
  schedule: async () => {
    const data = await graphql<{ displaySchedule: DisplaySchedule }>(
      `query ManageDisplaySchedule { displaySchedule { ${SCHEDULE_FIELDS} } }`,
    )
    return data.displaySchedule
  },

  status: async () => {
    const data = await graphql<{ displayStatus: DisplayStatus }>(
      `query ManageDisplayStatus { displayStatus { ${STATUS_FIELDS} } }`,
    )
    return data.displayStatus
  },

  wifi: async () => {
    const data = await graphql<{ wifiStatus: WifiStatus }>(
      `query ManageWifiStatus { wifiStatus { connected ssid signalPercent } }`,
    )
    return data.wifiStatus
  },

  /** ddcBus is always sent back: the input defaults it to empty, which would disable panel control. */
  updateSchedule: async (schedule: DisplaySchedule) => {
    const data = await graphql<{ updateDisplaySchedule: DisplaySchedule }>(
      `mutation UpdateDisplaySchedule($input: UpdateDisplayScheduleInput!) {
        updateDisplaySchedule(input: $input) { ${SCHEDULE_FIELDS} }
      }`,
      { input: schedule },
    )
    return data.updateDisplaySchedule
  },

  preview: async ({ mode, seconds }: { mode: DisplayModeKey; seconds: number }) => {
    const data = await graphql<{ previewDisplayMode: DisplayStatus }>(
      `mutation PreviewDisplayMode($mode: DisplayMode!, $seconds: Int!) {
        previewDisplayMode(mode: $mode, seconds: $seconds) { ${STATUS_FIELDS} }
      }`,
      { mode, seconds },
    )
    return data.previewDisplayMode
  },
}
