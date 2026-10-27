import { QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import { App } from './app'
import { DisplayTheme } from './features/display/display-theme'
import { queryClient } from './lib/query-client'

// Lets index.css keep the pointer visible while you are developing; the
// kiosk build never sets this, so the panel stays cursorless.
if (import.meta.env.DEV) {
  document.documentElement.dataset.dev = ''
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <DisplayTheme>
        <App />
      </DisplayTheme>
    </QueryClientProvider>
  </StrictMode>,
)
