import { Cloud, CloudMoon, CloudRain, CloudSun, type LucideProps, Sun } from 'lucide-react'

export type WeatherCondition = 'sun' | 'cloud-sun' | 'cloud' | 'rain' | 'moon'

const ICONS = { sun: Sun, 'cloud-sun': CloudSun, cloud: Cloud, rain: CloudRain, moon: CloudMoon }

export const CONDITION_LABEL: Record<WeatherCondition, string> = {
  sun: 'Sunny',
  'cloud-sun': 'Partly cloudy',
  cloud: 'Cloudy',
  rain: 'Rain',
  moon: 'Partly cloudy',
}

/** Lucide weather icon. Stroke 1.75 at 36–40px, 1.5 at 128px (pass strokeWidth). */
export function WeatherIcon({
  condition,
  strokeWidth = 1.75,
  ...props
}: LucideProps & { condition: WeatherCondition }) {
  const Icon = ICONS[condition]
  return <Icon aria-hidden strokeWidth={strokeWidth} {...props} />
}
