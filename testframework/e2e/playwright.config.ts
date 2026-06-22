/**
 * Конфиг перенесён в frontend/playwright.config.ts (ESM + @playwright/test resolution).
 * Запуск из frontend/: npx playwright test -c playwright.config.ts
 */
import config from "../../frontend/playwright.config.ts";

export default config;
