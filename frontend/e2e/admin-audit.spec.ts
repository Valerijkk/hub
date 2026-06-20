import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "./fixtures";
import { ADMIN_EMAIL, authenticatePage, devLogin } from "./helpers/auth";

test.describe("Admin audit page", () => {
	test("loads audit log viewer for admin user", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(ADMIN_EMAIL);
		expect(login.user.system_role).toBe("admin");

		await authenticatePage(page, login);
		await page.goto("/admin/audit");

		await expect(page).toHaveURL(/\/admin\/audit/);
		await expect(page.getByRole("button", { name: "Применить" })).toBeVisible();
		await expect(page.getByRole("button", { name: "CSV" })).toBeVisible();
		await expect(page.getByRole("columnheader", { name: "Время" })).toBeVisible();
		await expect(
			page.getByRole("columnheader", { name: "Действие" }),
		).toBeVisible();
	});

	test("redirects non-admin users away from audit page", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin("dev1@hub.local");
		expect(login.user.system_role).toBe("member");

		await authenticatePage(page, login);
		await page.goto("/admin/audit");

		await expect(page).not.toHaveURL(/\/admin\/audit/);
	});
});
