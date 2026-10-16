import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  base: "/manage/",
  plugins: [
    react(),
    tailwindcss(),
  ],
  server: {
    host: true,
    port: 5174,
    proxy: {
      "/graphql": {
        target: process.env.HOMIE_API_URL ?? "http://localhost:3000",
        ws: true,
      },
    },
  },
});
