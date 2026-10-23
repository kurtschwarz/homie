import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthGate } from '@/features/auth'
import { CalendarsPage } from '@/features/calendars/calendars-page'
import { PeoplePage } from '@/features/people/people-page'
import { SystemPage } from '@/features/system'
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
          <Route
            path="weather"
            element={
              <PlaceholderPage title="Weather" description="Weather settings will live here." />
            }
          />
          <Route
            path="kitchen-screen"
            element={
              <PlaceholderPage
                title="Kitchen screen"
                description="Kitchen screen settings will live here."
              />
            }
          />
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
