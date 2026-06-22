/**
 * Playwright E2E — конфиг в frontend/ для корректного ESM и @playwright/test resolution.
 * Спеки остаются в testframework/e2e/specs/ (Wave B TF-M07).
 */
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig, devices } from "@playwright/test";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const e2eSpecsDir = path.resolve(__dirname, "../testframework/e2e/specs");
const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:5173";
const startWebServer = process.env.PLAYWRIGHT_SKIP_WEBSERVER !== "1";

export default defineConfig({
	testDir: e2eSpecsDir,
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: [["list"], ["html", { open: "never", outputFolder: "playwright-report" }]],
	use: {
		baseURL,
		locale: "ru-RU",
		trace: "on-first-retry",
		screenshot: "only-on-failure",
	},
	projects: [
		{
			name: "chromium",
			use: { ...devices["Desktop Chrome"] },
		},
	],
	...(startWebServer
		? {
				webServer: {
					command: "npm run dev -- --host 127.0.0.1",
					cwd: __dirname,
					url: baseURL,
					reuseExistingServer: !process.env.CI,
					timeout: 180_000,
				},
			}
		: {}),
});
