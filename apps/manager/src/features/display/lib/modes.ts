import type { DisplayModeKey } from '../types'

export type ModePresentation = {
  label: string
  /** Label on the right-now switcher, where power saving reads as "Off". */
  switcherLabel: string
  description: string
  /** Approximate palette of the theme, for swatches, timeline bands and the mock device. */
  palette: { bg: string; surface: string; ink: string; ink2: string; accent: string }
  /** Fill and label colour of the mode's band on the day timeline. */
  band: { bg: string; ink: string }
}

/** The backend enum is LIGHT, DARK and NIGHT; the admin calls them Light, Dim and Dark. */
export const MODES: Record<DisplayModeKey, ModePresentation> = {
  LIGHT: {
    label: 'Light',
    switcherLabel: 'Light',
    description: 'Bright and full colour, for daytime.',
    palette: {
      bg: '#f7f5f0',
      surface: '#ffffff',
      ink: '#1b1c1f',
      ink2: '#5d6068',
      accent: '#e8683a',
    },
    band: { bg: '#f8eecb', ink: '#5c4a14' },
  },
  DARK: {
    label: 'Dim',
    switcherLabel: 'Dim',
    description: 'Neutral grey on dark, for the evening.',
    palette: {
      bg: '#0e0f11',
      surface: '#17181b',
      ink: '#f2f2ef',
      ink2: '#b9bbc1',
      accent: '#ff9a62',
    },
    band: { bg: '#bcc1d0', ink: '#2f3340' },
  },
  NIGHT: {
    label: 'Dark',
    switcherLabel: 'Dark',
    description: "Black background with just the clock and what's next, in warm amber.",
    palette: {
      bg: '#0b0907',
      surface: '#16120e',
      ink: '#e2cfb6',
      ink2: '#b39e85',
      accent: '#e0a060',
    },
    band: { bg: '#3e4250', ink: '#eef0f5' },
  },
  OFF: {
    label: 'Power saving',
    switcherLabel: 'Off',
    description: 'The screen is switched off. Tap it to wake.',
    palette: {
      bg: '#000000',
      surface: '#000000',
      ink: '#000000',
      ink2: '#000000',
      accent: '#000000',
    },
    band: { bg: '#17181b', ink: '#8a8c93' },
  },
}

/** Display order, which is also the order the schedule must be in around the day. */
export const MODE_ORDER: DisplayModeKey[] = ['LIGHT', 'DARK', 'NIGHT', 'OFF']
