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
    // Owner-bound Calendar WIP remains outside this merge authority.
    exclude: [
      "/opt/hub/source/frontend/tests/calls/call-calendar-a11y.test.tsx",
    ],
    testTimeout: 15_000,
    // The Docker QA service is capped at 2 CPUs / 2 GiB; bound jsdom workers to it.
    maxWorkers: 2,
    reporters: ["default", "json"],
    outputFile: {
      json: "/reports/frontend-unit/vitest.report.json",
    },
    environment: "jsdom",
    globals: true,
    setupFiles: ["/opt/hub/source/frontend/src/test/setup.ts"],
  },
});
