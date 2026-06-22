/**
 * Конфиг перенесён в frontend/playwright.config.ts (ESM + @playwright/test resolution).
 * Запуск из frontend/: npm run test:e2e | npm run test:e2e:smoke
 */
import config from "../../frontend/playwright.config.ts";

export default config;
