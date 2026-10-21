import { Navigate, Route, Routes } from "react-router-dom";
import { AppDataProvider } from "./app-data";
import { AuthGate } from "./auth/auth-gate";
import { CalendarsPage } from "./features/calendars/calendars-page";
import { PeoplePage } from "./features/people/people-page";
import { PlaceholderPage } from "./features/placeholder/placeholder-page";
import { SystemPage } from "./features/system/system-page";
import { AdminShell } from "./shell/admin-shell";

export function App() {
  return (
    <AuthGate>
      {(user, logout) => (
        <AppDataProvider user={user} logout={logout}>
          <Routes>
            <Route element={<AdminShell />}>
              <Route index element={<Navigate to="people" replace />} />
              <Route path="people" element={<PeoplePage />} />
              <Route path="calendars" element={<CalendarsPage />} />
              <Route path="weather" element={<PlaceholderPage title="Weather" description="Weather settings will live here." />} />
              <Route path="kitchen-screen" element={<PlaceholderPage title="Kitchen screen" description="Kitchen screen settings will live here." />} />
              <Route path="phones" element={<PlaceholderPage title="Phones" description="Connected phones will appear here." />} />
              <Route path="system" element={<SystemPage />} />
              <Route path="*" element={<Navigate to="people" replace />} />
            </Route>
          </Routes>
        </AppDataProvider>
      )}
    </AuthGate>
  );
}
