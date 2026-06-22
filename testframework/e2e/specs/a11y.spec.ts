/**
 * E2E-компонент Hub: a11y.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	skipUnlessFrontend,
	test,
} from "../fixtures";
import { formatViolations, scanAccessibility } from "../helpers/a11y";
import { AppShellPage, LoginPage } from "../pages";
import { AuthSteps } from "../steps";

/** Authenticated pages: scan main landmark only (sidebar is tracked separately). */
const MAIN_A11Y = { include: "main" } as const;

/** Messenger: channel list lives inside main but is covered by dedicated UI tests. */
const MESSENGER_MAIN_A11Y = {
	include: "main",
	exclude: ["[data-testid=messenger-channel-list]"],
} as const;

/** Dev-only login panel — excluded from login page axe (contrast badges are dev UX). */
const LOGIN_A11Y_EXCLUDE = { exclude: ["[data-testid=dev-login-panel]"] } as const;

// E2E-группа: Доступность страницы входа.
test.describe("Login page accessibility", () => {
	// Проверяет: на всей странице нет критических или серьёзных нарушений axe.
	test("full page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);

		const loginPage = new LoginPage(page);
		await loginPage.open();
		await expect(loginPage.hubTitle).toBeVisible();

		const violations = await scanAccessibility(page, LOGIN_A11Y_EXCLUDE);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: панель формы входа без критических или серьёзных нарушений axe.
	test("login form panel has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);

		const loginPage = new LoginPage(page);
		await loginPage.open();
		const formPanel = page.locator(".login-form-animate");
		await expect(formPanel).toBeVisible();
		await expect(loginPage.magicHint).toBeVisible();

		const violations = await scanAccessibility(page, {
			include: ".login-form-animate",
		});
		expect(violations, formatViolations(violations)).toEqual([]);
	});
});

// E2E-группа: Доступность авторизованных маршрутов.
test.describe("Authenticated routes accessibility", () => {
	// Требует dev-login на backend (seed + /api/dev/login). Пропускает тест,
	// если backend или DEV_MODE недоступен — то же условие пропуска, что у других авторизованных e2e.
	// Проверяет: профиль настроек без критических или серьёзных нарушений axe.
	test("settings profile has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsAdmin(page);
		await shell.openProfile();

		await expect(page).toHaveURL(/\/settings\/profile/);
		await expect(page.getByRole("main")).toBeVisible();
		await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: внешний вид настроек без критических или серьёзных нарушений axe.
	test("settings appearance has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/settings/appearance");

		await expect(page).toHaveURL(/\/settings\/appearance/);
		await expect(page.getByRole("main")).toBeVisible();
		await expect(page.getByTestId("appearance-theme-heading")).toBeVisible({
			timeout: 15_000,
		});

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница справки без критических или серьёзных нарушений axe.
	test("help page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/help");

		await expect(page).toHaveURL(/\/help/);
		await expect(page.getByTestId("help-page-heading")).toBeVisible({
			timeout: 15_000,
		});

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: список проектов без критических или серьёзных нарушений axe.
	test("projects list has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsAdmin(page);
		await shell.gotoProjects();

		await expect(page).toHaveURL(/\/projects/);
		await expect(page.getByTestId("projects-page-heading")).toBeVisible({
			timeout: 15_000,
		});

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: канал мессенджера без критических или серьёзных нарушений axe.
	test("messenger channel has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsDev(page);
		await page.goto("/messenger");
		await expect(page).toHaveURL(/\/messenger\/[^/]+/, { timeout: 15_000 });
		await shell.expectShellVisible();

		const violations = await scanAccessibility(page, MESSENGER_MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: kanban-доска без критических или серьёзных нарушений axe.
	test("kanban board has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsDev(page);

		const projectId = await shell.openDemoProject();
		await page.goto(`/projects/${projectId}/board`);
		await expect(page.getByRole("main")).toBeVisible({ timeout: 15_000 });

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница звонков без критических или серьёзных нарушений axe.
	test("calls page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsDev(page);
		await page.goto("/calls");

		await expect(page).toHaveURL(/\/calls/);
		await expect(
			page.getByRole("heading", { name: /звонки|calls/i }),
		).toBeVisible({ timeout: 15_000 });

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница TMS-кейсов без критических или серьёзных нарушений axe.
	test("TMS cases page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsDev(page);

		const projectId = await shell.openDemoProject();
		await page.goto(`/projects/${projectId}/tms/cases`);
		await expect(
			page.getByRole("button", { name: "Создать кейс" }),
		).toBeVisible({ timeout: 15_000 });

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница аудита admin без критических или серьёзных нарушений axe.
	test("admin audit page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/admin/audit");

		await expect(page).toHaveURL(/\/admin\/audit/);
		await expect(page.getByRole("button", { name: "Применить" })).toBeVisible();

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница корзины без критических или серьёзных нарушений axe.
	test("trash page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/trash");

		await expect(page).toHaveURL(/\/trash/);
		await expect(page.getByTestId("trash-page-heading")).toBeVisible({
			timeout: 15_000,
		});
		await expect(page.getByRole("main")).toBeVisible();

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: настройки уведомлений без критических или серьёзных нарушений axe.
	test("notification settings has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/settings/notifications");

		await expect(page).toHaveURL(/\/settings\/notifications/);
		await expect(page.getByRole("heading", { name: "Уведомления" })).toBeVisible();

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	// Проверяет: страница пользователей admin без критических или серьёзных нарушений axe.
	test("admin users page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/admin/users");

		await expect(page).toHaveURL(/\/admin\/users/);
		await expect(
			page.getByRole("columnheader", { name: /пользоват/i }),
		).toBeVisible({ timeout: 15_000 });

		const violations = await scanAccessibility(page, MAIN_A11Y);
		expect(violations, formatViolations(violations)).toEqual([]);
	});
});
