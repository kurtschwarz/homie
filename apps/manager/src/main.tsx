import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import { HomieThemeProvider } from '@homie/ui/components/theme-provider'
import { App } from './app'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <HomieThemeProvider theme="auto" surface="auto">
      <BrowserRouter basename="/manage">
        <App />
      </BrowserRouter>
    </HomieThemeProvider>
  </StrictMode>,
)
