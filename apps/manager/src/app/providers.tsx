import { HomieThemeProvider } from '@homie/ui/components/theme-provider'
import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import type { ReactNode } from 'react'
import { BrowserRouter } from 'react-router-dom'
import { queryClient } from './query-client'

/**
 * Every provider the manager needs, in one place so `main.tsx` stays a
 * mount call. The devtools panel is dev-only: it ships as a devDependency,
 * and `import.meta.env.DEV` lets the production bundle drop it entirely.
 */
export function Providers({ children }: { children: ReactNode }) {
  return (
    <HomieThemeProvider theme="auto" surface="auto">
      <QueryClientProvider client={queryClient}>
        <BrowserRouter basename="/manage">{children}</BrowserRouter>
        {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
      </QueryClientProvider>
    </HomieThemeProvider>
  )
}
