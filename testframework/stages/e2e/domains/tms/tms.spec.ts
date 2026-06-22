/**
 * E2E-компонент Hub: tms.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { AppShellPage } from "../../shared/pages";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: TMS (FB-09).
test.describe("TMS (FB-09)", () => {
	// Проверяет: создание ручного тест-кейса в DEMO-проекте.
	test("create manual test case in DEMO project", async ({
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

		await page.goto(`/projects/${projectId}/tms/cases`);
		await expect(page.getByRole("button", { name: "Создать кейс" })).toBeVisible({
			timeout: 15_000,
		});

		const caseTitle = `E2E Case ${Date.now()}`;
		await page.getByRole("button", { name: "Создать кейс" }).click();
		await page.getByPlaceholder("Название тест-кейса").fill(caseTitle);
		await page.getByRole("button", { name: "Создать" }).click();

		await expect(page.getByText(caseTitle)).toBeVisible({ timeout: 15_000 });
	});
});
