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

		await page.goto("/admin/dashboard");
		await page.getByRole("button", { name: "AI" }).click();

		await expect(page.getByRole("heading", { name: "AI Provider Hub" })).toBeVisible();
		await expect(page.getByText("System prompt @hub-ai")).toBeVisible();
		await expect(page.getByRole("button", { name: "DeepSeek" })).toBeVisible();
		await expect(page.getByRole("button", { name: "Открыть DM с @hub-ai" })).toBeVisible();
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
