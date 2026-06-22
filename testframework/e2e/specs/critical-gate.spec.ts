/**
 * E2E-01 — пять критических путей (PLAN-CLOSE-GAPS Stage 3.2).
 * CI: e2e-smoke выполняет путь 1; e2e-full — пути 2–5 с backend.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	skipUnlessFrontend,
	test,
} from "../fixtures";
import { fetchHealth } from "../helpers/backend";
import { AuthSteps, NavigationSteps } from "../steps";
import { LoginPage, AppShellPage } from "../pages";

// E2E-группа: E2E-01 критические пути контроля.
test.describe("E2E-01 critical gate paths", () => {
	// Проверяет: путь 1 — страница входа загружается без backend.
	test("path 1: login page loads without backend", async ({
		page,
		frontendAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);

		const auth = new AuthSteps(page);
		await auth.openLoginAndExpectBranding();
	});

	// Проверяет: путь 2 — dev-login ведёт на список проектов.
	test("path 2: dev login reaches projects list", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const loginPage = new LoginPage(page);
		await loginPage.open();
		await loginPage.devLoginAs("Дмитрий Кодов");
	});

	// Проверяет: путь 3 — боковая панель переходит на дашборд и настройки.
	test("path 3: sidebar navigates dashboard and settings", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const nav = new NavigationSteps(page);
		await auth.loginAsDev(page);
		await nav.fromProjectsToDashboardAndSettings();
	});

	// Проверяет: путь 4 — глобальный поиск открывается по Ctrl+K.
	test("path 4: global search opens with Ctrl+K", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsDev(page);
		await shell.gotoProjects();
		await shell.openCommandPalette();
	});

	// Проверяет: путь 5 — GET /api/health возвращает статус сервисов.
	test("path 5: API health reports service status", async ({
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		const health = await fetchHealth();
		expect(health).not.toBeNull();
		expect(health?.status).toMatch(/^(ok|degraded)$/);
		expect(health?.services).toHaveProperty("postgres");
	});
});
