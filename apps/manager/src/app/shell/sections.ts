import {
  CalendarDays,
  CloudSun,
  type LucideIcon,
  Monitor,
  Settings,
  Smartphone,
  Users,
} from 'lucide-react'

/** The admin sections, in sidebar order. Paths are relative to the /manage basename. */
export const sections: { to: string; label: string; icon: LucideIcon }[] = [
  { to: '/people', label: 'People', icon: Users },
  { to: '/calendars', label: 'Calendars', icon: CalendarDays },
  { to: '/weather', label: 'Weather', icon: CloudSun },
  { to: '/kitchen-screen', label: 'Kitchen screen', icon: Monitor },
  { to: '/phones', label: 'Phones', icon: Smartphone },
  { to: '/system', label: 'System', icon: Settings },
]
