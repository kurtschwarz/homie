import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/graphql": {
        target: process.env.HOMIE_API_URL ?? "http://localhost:3000",
        ws: true,
      },
    },
  },
});
