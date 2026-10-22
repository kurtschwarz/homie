import { readFileSync } from 'node:fs'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/graphql': {
        target: process.env.HOMIE_API_URL ?? 'http://localhost:3000',
        ws: true,
        configure: (proxy) => {
          let bearerToken = process.env.HOMIE_AUTH_TOKEN?.trim()
          proxy.on('proxyReq', (proxyReq) => {
            if (!bearerToken) {
              try {
                bearerToken = readFileSync(
                  process.env.HOMIE_AUTH_TOKEN_PATH ?? '/run/homie-token/token',
                  'utf8',
                ).trim()
              } catch {
                // The API creates the shared token file during its own startup.
                // Retry on the next proxied request if the display starts first.
              }
            }
            if (bearerToken) proxyReq.setHeader('Authorization', `Bearer ${bearerToken}`)
          })
        },
      },
      '/photos': {
        target: process.env.HOMIE_API_URL ?? 'http://localhost:3000',
      },
    },
  },
})
