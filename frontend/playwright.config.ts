/**
 * Playwright E2E — конфиг в frontend/ для ESM и @playwright/test resolution.
 * Спеки: testframework/stages/e2e/domains/{domain}/*.spec.ts
 */
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig, devices } from "@playwright/test";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const e2eDomainsDir = path.resolve(__dirname, "../testframework/stages/e2e/domains");
const smokeSpecs = [
	"auth/login.spec.ts",
	"platform/a11y.spec.ts",
	"platform/critical-gate.spec.ts",
];
const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:5173";
const startWebServer = process.env.PLAYWRIGHT_SKIP_WEBSERVER !== "1";
const chromium = { ...devices["Desktop Chrome"] };

export default defineConfig({
	testDir: e2eDomainsDir,
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
			name: "chromium-smoke",
			testMatch: smokeSpecs,
			use: chromium,
		},
		{
			name: "chromium-full",
			testMatch: "**/*.spec.ts",
			use: chromium,
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
