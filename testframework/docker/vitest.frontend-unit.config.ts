import { defineConfig } from 'vitest/config'

export default defineConfig({
  root: '/opt/hub/source/frontend',
  test: {
    include: ['/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}'],
    environment: 'jsdom',
  },
})
