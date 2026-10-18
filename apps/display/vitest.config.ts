import { defineConfig } from "vitest/config";

export default defineConfig({
  resolve: {
    alias: {
      "@homie/ui": new URL("../../packages/ui/src", import.meta.url).pathname,
    },
  },
});
