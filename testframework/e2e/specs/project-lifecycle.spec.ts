/**
 * E2E-компонент Hub: project-lifecycle.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AppShellPage } from "../pages";
import { AuthSteps } from "../steps";

// E2E-группа: Жизненный цикл проекта (FB-07).
test.describe("Project lifecycle (FB-07)", () => {
	// Проверяет: создание проекта, доска, удаление и восстановление из корзины.
	test("create project, verify board, delete and restore from trash", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const suffix = Date.now().toString(36).toUpperCase().slice(-6);
		const projectName = `E2E Project ${suffix}`;
		const projectKey = `E${suffix}`.slice(0, 10);

		const auth = new AuthSteps(page);
		const shell = new AppShellPage(page);
		await auth.loginAsAdmin(page);

		await page.goto("/projects/new");
		await expect(page.locator("#proj-name")).toBeVisible({ timeout: 15_000 });
		await page.locator("#proj-name").fill(projectName);
		await page.locator("#proj-key").fill(projectKey);
		await page.getByRole("button", { name: "Создать проект" }).click();

		await expect(page).toHaveURL(/\/projects\/[0-9a-f-]+/, { timeout: 30_000 });
		const projectId = page.url().match(/\/projects\/([0-9a-f-]+)/)?.[1];
		expect(projectId).toBeTruthy();
		await page.goto(`/projects/${projectId}/board`);
		await expect(page.locator("#main-content")).toBeVisible({ timeout: 15_000 });

		await page.goto(`/projects/${projectId}`);
		await expect(page.getByRole("tab", { name: "Обзор" })).toBeVisible({
			timeout: 15_000,
		});
		const settingsTab = page
			.getByRole("tablist")
			.getByRole("tab", { name: "Настройки" });
		await expect(settingsTab).toBeVisible({ timeout: 15_000 });
		await settingsTab.click();
		await page.getByRole("button", { name: "Удалить" }).click();
		await page.getByRole("button", { name: "Удалить" }).last().click();

		await expect(page).toHaveURL(/\/projects/, { timeout: 15_000 });

		await page.goto("/admin/trash");
		await page.getByRole("button", { name: "Проекты" }).click();
		const trashRow = page.getByRole("row").filter({ hasText: projectName });
		await expect(trashRow).toBeVisible({ timeout: 15_000 });

		await trashRow.getByRole("button", { name: "Восстановить" }).click();
		await page.getByRole("button", { name: "Восстановить" }).last().click();

		await shell.gotoProjects();
		await expect(
			page
				.locator("#main-content")
				.getByRole("link")
				.filter({ hasText: projectName }),
		).toBeVisible({ timeout: 15_000 });
	});
});
