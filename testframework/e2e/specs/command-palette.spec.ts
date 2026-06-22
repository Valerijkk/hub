/**
 * E2E-компонент Hub: command-palette.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AppShellPage } from "../pages";
import { AuthSteps } from "../steps";

// E2E-группа: Глобальный поиск Ctrl+K (E2E-06).
test.describe("Global search Ctrl+K (E2E-06)", () => {
	// Проверяет: палитра команд открывается с клавиатуры и из верхней панели.
	test("opens command palette via keyboard and topbar", async ({
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
