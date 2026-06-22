/**
 * E2E-компонент Hub: admin-audit.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: Страница аудита admin.
test.describe("Admin audit page", () => {
	// Проверяет: просмотр журнала аудита загружается для admin.
	test("loads audit log viewer for admin user", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsAdmin(page);
		expect(login.user.system_role).toBe("admin");

		await page.goto("/admin/audit");

		await expect(page).toHaveURL(/\/admin\/audit/);
		await expect(page.getByRole("button", { name: "Применить" })).toBeVisible({
			timeout: 15_000,
		});
		await expect(
			page.getByRole("button", { name: /csv|экспорт/i }),
		).toBeVisible();
		await expect(
			page.getByRole("columnheader", { name: "Время", exact: true }),
		).toBeVisible({ timeout: 15_000 });
		await expect(
			page.getByRole("columnheader", { name: "Действие" }),
		).toBeVisible({ timeout: 15_000 });
	});

	// Проверяет: фильтр действий аудита содержит опции алертов.
	test("audit action filter includes alert options", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);
		await page.goto("/admin/audit");

		await page.getByLabel("Действие").click();
		await expect(page.getByRole("option", { name: "Алерт (сраб.)" })).toBeVisible();
		await expect(page.getByRole("option", { name: "Алерт (снят)" })).toBeVisible();
	});

	// Проверяет: не-admin перенаправляется со страницы аудита.
	test("redirects non-admin users away from audit page", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsDev(page);
		expect(login.user.system_role).toBe("member");

		await page.goto("/admin/audit");

		await expect(page).not.toHaveURL(/\/admin\/audit/);
	});
});
