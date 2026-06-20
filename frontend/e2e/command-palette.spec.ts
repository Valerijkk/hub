import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "./fixtures";
import { authenticatePage, devLogin, DEV_EMAIL } from "./helpers/auth";

test.describe("Global search Ctrl+K (E2E-06)", () => {
	test("opens command palette via keyboard and topbar", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/projects");

		await page.keyboard.press("Control+k");
		await expect(page.getByRole("dialog")).toBeVisible({ timeout: 5000 });
		await expect(page.getByPlaceholder(/поиск|search/i)).toBeVisible();

		await page.keyboard.press("Escape");
		await expect(page.getByRole("dialog")).toBeHidden({ timeout: 3000 });

		const searchTrigger = page.getByRole("button", {
			name: /глобальный поиск|global search/i,
		});
		if (await searchTrigger.isVisible()) {
			await searchTrigger.click();
			await expect(page.getByRole("dialog")).toBeVisible();
		}
	});
});
