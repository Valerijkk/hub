import { defineConfig } from "vitest/config";

export default defineConfig({
  root: "/opt/hub/source/frontend",
  resolve: {
    alias: {
      "@": "/opt/hub/source/frontend/src",
    },
  },
  test: {
    include: ["/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}"],
    environment: "jsdom",
    globals: true,
    setupFiles: ["/opt/hub/source/frontend/src/test/setup.ts"],
  },
});
