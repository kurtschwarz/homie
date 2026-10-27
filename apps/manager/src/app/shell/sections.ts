import {
  CalendarDays,
  CloudSun,
  type LucideIcon,
  Monitor,
  Settings,
  Smartphone,
  Users,
} from 'lucide-react'

/** The admin sections, in sidebar order. Paths are relative to the site root. */
export const sections: { to: string; label: string; icon: LucideIcon }[] = [
  { to: '/people', label: 'People', icon: Users },
  { to: '/calendars', label: 'Calendars', icon: CalendarDays },
  { to: '/weather', label: 'Weather', icon: CloudSun },
  { to: '/display', label: 'Display', icon: Monitor },
  { to: '/phones', label: 'Phones', icon: Smartphone },
  { to: '/system', label: 'System', icon: Settings },
]
