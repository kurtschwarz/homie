import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthGate } from '@/features/auth'
import { AddCalendarAccountPage, CalendarsPage } from '@/features/calendars'
import { DisplayPage } from '@/features/display'
import { PeoplePage } from '@/features/people'
import { SystemPage } from '@/features/system'
import { WeatherPage } from '@/features/weather'
import { PlaceholderPage } from '@/shared/ui/placeholder-page'
import { AdminShell } from './shell/admin-shell'

export function App() {
  return (
    <AuthGate>
      <Routes>
        <Route element={<AdminShell />}>
          <Route index element={<Navigate to="people" replace />} />
          <Route path="people" element={<PeoplePage />} />
          <Route path="calendars" element={<CalendarsPage />} />
          <Route path="calendars/add" element={<AddCalendarAccountPage />} />
          <Route path="calendars/:accountId/edit" element={<AddCalendarAccountPage />} />
          <Route path="weather" element={<WeatherPage />} />
          <Route path="display" element={<DisplayPage />} />
          <Route
            path="phones"
            element={
              <PlaceholderPage title="Phones" description="Connected phones will appear here." />
            }
          />
          <Route path="system" element={<SystemPage />} />
          <Route path="*" element={<Navigate to="people" replace />} />
        </Route>
      </Routes>
    </AuthGate>
  )
}
