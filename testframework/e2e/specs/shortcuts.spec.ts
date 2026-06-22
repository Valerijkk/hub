/**
 * E2E-компонент Hub: shortcuts.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AppShellPage } from "../pages";
import { AuthSteps } from "../steps";

// E2E-группа: Оверлей горячих клавиш (FB-01 / E2E).
test.describe("Shortcuts overlay (FB-01 / E2E)", () => {
	// Проверяет: диалог горячих клавиш открывается по клавише ?.
	test("opens shortcuts dialog on ? key", async ({
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
		await expect(
			page.getByRole("heading", { name: "Проекты" }),
		).toBeVisible();

		await page.keyboard.press("Shift+/");
		await expect(page.getByRole("dialog")).toBeVisible({ timeout: 5000 });
		await expect(
			page.getByRole("heading", { name: /горяч|shortcuts/i }),
		).toBeVisible();

		await page.keyboard.press("Escape");
		await expect(page.getByRole("dialog")).toBeHidden({ timeout: 3000 });
	});
});
