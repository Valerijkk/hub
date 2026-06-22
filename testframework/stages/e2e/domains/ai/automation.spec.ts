/**
 * E2E-компонент Hub: automation.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { AppShellPage } from "../../shared/pages";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: Автоматизация workflow (PROD-01).
test.describe("Workflow automation (PROD-01)", () => {
	// Проверяет: панель правил автоматизации в настройках workflow.
	test("automation rules panel on workflow settings", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsDev(page);

		const projectId = await shell.openDemoProject();

		await page.goto(`/projects/${projectId}/settings/workflow`);
		await expect(
			page.getByRole("heading", { name: "Настройки Workflow" }),
		).toBeVisible({ timeout: 15_000 });

		await expect(page.getByText("Новое правило")).toBeVisible();
		await expect(page.getByText("Правила (0/3)")).toBeVisible({
			timeout: 15_000,
		});
		await expect(page.getByText("Нет правил автоматизации")).toBeVisible();
	});
});
