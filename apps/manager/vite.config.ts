import { fileURLToPath } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  base: '/manage/',
  plugins: [react(), tailwindcss()],
  resolve: {
    // The trailing slash matters: a bare '@' key would also match the
    // '@homie/ui' package specifier and rewrite it into src/.
    alias: { '@/': `${fileURLToPath(new URL('./src', import.meta.url))}/` },
  },
  server: {
    host: true,
    port: 5174,
    proxy: {
      '/graphql': {
        target: process.env.HOMIE_API_URL ?? 'http://localhost:3000',
        ws: true,
        configure: (proxy) =>
          proxy.on('proxyReq', (proxyReq) => {
            if (process.env.HOMIE_AUTH_TOKEN)
              proxyReq.setHeader('Authorization', `Bearer ${process.env.HOMIE_AUTH_TOKEN}`)
          }),
      },
    },
  },
})
