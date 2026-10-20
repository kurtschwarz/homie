import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/graphql": {
        target: process.env.HOMIE_API_URL ?? "http://localhost:3000",
        ws: true,
        configure: (proxy) => proxy.on("proxyReq", (proxyReq) => {
          if (process.env.HOMIE_AUTH_TOKEN) proxyReq.setHeader("Authorization", `Bearer ${process.env.HOMIE_AUTH_TOKEN}`);
        }),
      },
      "/photos": {
        target: process.env.HOMIE_API_URL ?? "http://localhost:3000",
      },
    },
  },
});
