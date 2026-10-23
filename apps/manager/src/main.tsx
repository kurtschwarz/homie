import { QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import { HomieThemeProvider } from '@homie/ui/components/theme-provider'
import { App } from './app'
import { queryClient } from './lib/query-client'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <HomieThemeProvider theme="auto" surface="auto">
      <QueryClientProvider client={queryClient}>
        <BrowserRouter basename="/manage">
          <App />
        </BrowserRouter>
      </QueryClientProvider>
    </HomieThemeProvider>
  </StrictMode>,
)
