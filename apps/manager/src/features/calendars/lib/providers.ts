export type CalendarProviderPreset = {
  id: string
  name: string
  detail: string
  url: string
  userLabel: string
  passwordLabel: string
  help: string[]
  available: boolean
}

export const CALENDAR_PROVIDERS: [CalendarProviderPreset, ...CalendarProviderPreset[]] = [
  {
    id: 'fastmail',
    name: 'Fastmail',
    detail: 'Preset · caldav.fastmail.com',
    url: 'https://caldav.fastmail.com/dav/',
    userLabel: 'Fastmail email',
    passwordLabel: 'App password',
    help: [
      'In Fastmail, open Settings → Privacy & Security.',
      'Make a new app password with calendar (CalDAV) access.',
      'Paste it here. Homie never sees your real password.',
    ],
    available: true,
  },
  {
    id: 'icloud',
    name: 'iCloud',
    detail: 'Preset · needs an app-specific password',
    url: 'https://caldav.icloud.com/',
    userLabel: 'Apple Account email',
    passwordLabel: 'App-specific password',
    help: [
      'Open account.apple.com and choose Sign-In and Security.',
      'Create an app-specific password for Homie.',
      'Paste it here. Homie never stores your Apple Account password.',
    ],
    available: true,
  },
  {
    id: 'nextcloud',
    name: 'Nextcloud',
    detail: 'Your server · /remote.php/dav',
    url: 'https://your-server.example/remote.php/dav/',
    userLabel: 'Username',
    passwordLabel: 'App password',
    help: [
      'Open your Nextcloud security settings.',
      'Create an app password for Homie.',
      'Replace “your-server.example” above with your server.',
    ],
    available: true,
  },
  {
    id: 'yahoo',
    name: 'Yahoo',
    detail: 'Preset · caldav.calendar.yahoo.com',
    url: 'https://caldav.calendar.yahoo.com/',
    userLabel: 'Yahoo email',
    passwordLabel: 'App password',
    help: [
      'Open Yahoo Account Security.',
      'Generate an app password for Homie.',
      'Use that password here instead of your account password.',
    ],
    available: true,
  },
  {
    id: 'synology',
    name: 'Synology Calendar',
    detail: 'Your NAS · /caldav',
    url: 'https://your-nas.example/caldav/',
    userLabel: 'Username',
    passwordLabel: 'Password',
    help: [
      'Enable CalDAV in Synology Calendar.',
      'Replace “your-nas.example” with your NAS address.',
      'Enter the account allowed to read these calendars.',
    ],
    available: true,
  },
  {
    id: 'other',
    name: 'Other CalDAV server',
    detail: 'Radicale, Baïkal, Zoho… enter it yourself',
    url: '',
    userLabel: 'Username',
    passwordLabel: 'Password',
    help: [
      'Find the CalDAV server URL in your provider’s documentation.',
      'Enter the username and password used for CalDAV.',
      'Use an app password if your provider supports one.',
    ],
    available: true,
  },
  {
    id: 'google',
    name: 'Google Calendar',
    detail: 'Coming later · needs Google sign-in',
    url: '',
    userLabel: 'Google email',
    passwordLabel: 'Password',
    help: [],
    available: false,
  },
]
