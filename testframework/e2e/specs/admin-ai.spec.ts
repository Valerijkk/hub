/**
 * E2E-компонент Hub: admin-ai.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AuthSteps } from "../steps";

// E2E-группа: Admin — AI Provider Hub.
test.describe("Admin AI Provider Hub", () => {
	// Проверяет: вкладка AI загружается на панели admin.
	test("loads AI tab on admin dashboard", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsAdmin(page);
		expect(login.user.system_role).toBe("admin");

		await page.goto("/admin/dashboard", { waitUntil: "domcontentloaded" });
		await expect(page.locator("#main-content")).toBeVisible({
			timeout: 30_000,
		});
		await expect(page.getByRole("button", { name: "Дашборд" })).toBeVisible({
			timeout: 30_000,
		});
		await expect(page.getByRole("button", { name: "AI" })).toBeVisible({
			timeout: 15_000,
		});
		await page.getByRole("button", { name: "AI" }).click();

		await expect(page.getByRole("heading", { name: "AI Provider Hub" })).toBeVisible({
			timeout: 15_000,
		});
		await expect(page.getByText(/System prompt @hub-ai/i)).toBeVisible({
			timeout: 15_000,
		});
		await expect(page.getByRole("button", { name: "DeepSeek" })).toBeVisible({
			timeout: 15_000,
		});
		await expect(
			page.getByRole("button", { name: "Открыть DM с @hub-ai" }),
		).toBeVisible({ timeout: 15_000 });
	});

	// Проверяет: не-admin перенаправляется с панели admin.
	test("redirects non-admin away from admin dashboard", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsDev(page);
		await page.goto("/admin/dashboard");

		await expect(page).not.toHaveURL(/\/admin\/dashboard/);
	});
});
