/**
 * Playwright-фикстуры Hub: доступность backend/frontend и хелперы пропуска.
 */
import { test as base } from "@playwright/test";

import { isBackendReachable, isDevModeAvailable } from "./helpers/backend";
import { isFrontendReachable } from "./helpers/frontend";

type HubFixtures = {
	backendAvailable: boolean;
	devModeAvailable: boolean;
	frontendAvailable: boolean;
};

export const test = base.extend<HubFixtures>({
	backendAvailable: async ({}, use) => {
		await use(await isBackendReachable());
	},
	devModeAvailable: async ({}, use) => {
		await use(await isDevModeAvailable());
	},
	frontendAvailable: async ({}, use) => {
		await use(await isFrontendReachable());
	},
});

export { expect } from "@playwright/test";

// Пропуск если backend недоступен.
export function skipUnlessBackend(backendAvailable: boolean): void {
	test.skip(
		!backendAvailable,
		"Backend not running — start with: docker compose -f docker-compose.local.yml up -d",
	);
}

// Пропуск если dev-mode выключен на backend.
export function skipUnlessDevMode(devModeAvailable: boolean): void {
	test.skip(
		!devModeAvailable,
		"Dev mode not available — set DEV_MODE=true on backend",
	);
}

// Пропуск если фронтенд недоступен.
export function skipUnlessFrontend(frontendAvailable: boolean): void {
	test.skip(
		!frontendAvailable,
		"Frontend not running — start with: npm run dev (or unset PLAYWRIGHT_SKIP_WEBSERVER=1)",
	);
}
