import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	skipUnlessFrontend,
	test,
} from "./fixtures";
import { formatViolations, scanAccessibility } from "./helpers/a11y";
import { ADMIN_EMAIL, authenticatePage, devLogin, DEV_EMAIL } from "./helpers/auth";

test.describe("Login page accessibility", () => {
	test("full page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);

		await page.goto("/login");
		await expect(page).toHaveURL(/\/login/);
		await expect(page.getByText("Hub").first()).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("login form panel has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);

		await page.goto("/login");
		const formPanel = page.locator(".login-form-animate");
		await expect(formPanel).toBeVisible();
		await expect(
			page.getByText("Войдите через Telegram — быстро и безопасно."),
		).toBeVisible();

		const violations = await scanAccessibility(page, {
			include: ".login-form-animate",
		});
		expect(violations, formatViolations(violations)).toEqual([]);
	});
});

test.describe("Authenticated routes accessibility", () => {
	// Requires backend dev login (seed + /api/dev/login). Skips when backend or
	// DEV_MODE is unavailable — same gate as other authenticated e2e specs.
	test("settings profile has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/settings/profile");

		await expect(page).toHaveURL(/\/settings\/profile/);
		await expect(page.getByRole("main")).toBeVisible();
		await expect(page.getByRole("heading", { name: "Профиль" })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("settings appearance has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/settings/appearance");

		await expect(page).toHaveURL(/\/settings\/appearance/);
		await expect(page.getByRole("main")).toBeVisible();
		await expect(page.getByRole("heading", { name: /тема/i })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("help page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/help");

		await expect(page).toHaveURL(/\/help/);
		await expect(page.getByRole("heading", { name: /справка hub/i })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("projects list has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/projects");

		await expect(page).toHaveURL(/\/projects/);
		await expect(page.getByRole("heading", { name: "Проекты" })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("messenger channel has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/messenger");

		await expect(page).toHaveURL(/\/messenger\/[^/]+/, { timeout: 15_000 });
		await expect(page.getByRole("main")).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("kanban board has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/projects");
		await page.getByRole("link", { name: /Hub Demo/i }).click();
		await expect(page).toHaveURL(/\/projects\/[^/]+/, { timeout: 10_000 });

		const projectUrl = page.url();
		const projectId = projectUrl.split("/projects/")[1]?.split(/[/?#]/)[0];
		expect(projectId).toBeTruthy();
		await page.goto(`/projects/${projectId}/board`);
		await expect(page.getByRole("main")).toBeVisible({ timeout: 15_000 });

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("admin audit page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/admin/audit");

		await expect(page).toHaveURL(/\/admin\/audit/);
		await expect(page.getByRole("button", { name: "Применить" })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("trash page has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/trash");

		await expect(page).toHaveURL(/\/trash/);
		await expect(page.getByText("Корзина")).toBeVisible();
		await expect(page.getByRole("main")).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});

	test("notification settings has no critical or serious axe violations", async ({
		page,
		frontendAvailable,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessFrontend(frontendAvailable);
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/settings/notifications");

		await expect(page).toHaveURL(/\/settings\/notifications/);
		await expect(page.getByRole("heading", { name: "Уведомления" })).toBeVisible();

		const violations = await scanAccessibility(page);
		expect(violations, formatViolations(violations)).toEqual([]);
	});
});
