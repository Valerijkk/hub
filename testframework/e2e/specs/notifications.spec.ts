/**
 * E2E-компонент Hub: notifications.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AuthSteps } from "../steps";

// E2E-группа: Настройки уведомлений.
test.describe("Notification settings", () => {
	// Проверяет: карточка привязки Telegram на странице настроек.
	test("shows Telegram bind card on settings page", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsDev(page);
		await page.goto("/settings/notifications");

		await expect(page).toHaveURL(/\/settings\/notifications/);
		await expect(page.getByRole("heading", { name: "Telegram" })).toBeVisible();
		await expect(
			page.getByRole("button", { name: "Привязать Telegram" }).or(
				page.getByRole("button", { name: "Отвязать" }),
			),
		).toBeVisible();
		await expect(page.getByText("Завершение прогона")).toBeVisible();
		await expect(page.getByText("Ошибки в прогоне")).toBeVisible();
	});
});
