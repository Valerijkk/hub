/**
 * E2E-компонент Hub: admin-metrics.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: Метрики операций admin.
test.describe("Admin operation metrics", () => {
	// Проверяет: вкладка метрик загружается на панели admin.
	test("metrics tab loads on admin dashboard", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsAdmin(page);
		expect(login.user.system_role).toBe("admin");

		await page.goto("/admin/dashboard");

		await expect(page).toHaveURL(/\/admin\/dashboard/);
		await page.getByRole("tab", { name: "Метрики" }).click();

		await expect(page.getByText("Метрики операций")).toBeVisible();
		await expect(
			page.getByRole("button", { name: "Обновить" }),
		).toBeVisible();
	});
});
