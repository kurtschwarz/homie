import { Outlet } from 'react-router-dom'
import { Sidebar } from './sidebar'

/** Sidebar beside a raised panel that holds the active section. */
export function AdminShell() {
  return (
    <main className="flex min-h-dvh flex-col gap-4 bg-hm-shell p-3 text-hm-ink md:flex-row md:gap-0 md:p-0">
      <Sidebar />
      <div className="flex min-h-[70vh] min-w-0 flex-1 flex-col overflow-hidden rounded-[20px] bg-hm-surface shadow-hm-panel md:my-3 md:mr-3 md:ml-0">
        <Outlet />
      </div>
    </main>
  )
}
