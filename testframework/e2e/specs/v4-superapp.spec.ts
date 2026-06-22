/**
 * E2E-компонент Hub: v4-superapp.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AuthSteps } from "../steps";

// E2E-группа: Smoke маршрутов Hub One v4 (Phase 51–57).
test.describe("Hub One v4 routes smoke (Phase 51–57)", () => {
	// Проверяет: v4 super-app страницы загружаются для авторизованного администратора.
	test("v4 super-app pages load for authenticated admin", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsAdmin(page);

		const routes = [
			{ path: "/calendar", heading: /Календарь/i },
			{ path: "/crm", heading: /^CRM$/i },
			{ path: "/hr", heading: /HR.*отпуска/i },
			{ path: "/marketplace", heading: /Барахолка/i },
			{ path: "/bases", heading: /Hub Base/i },
			{ path: "/intranet", heading: /^Intranet$/i },
			{ path: "/accounting", heading: /Бухгалтерия/i },
		] as const;

		for (const route of routes) {
			await page.goto(route.path);
			await expect(page.locator("#main-content")).toBeVisible({ timeout: 15_000 });
			await expect(page.getByRole("heading", { name: route.heading })).toBeVisible({
				timeout: 15_000,
			});
		}

		await page.goto("/accounting");
		await expect(page.getByText(/Счета на оплату/i)).toBeVisible({
			timeout: 15_000,
		});
		await expect(page.getByText(/Выручка/i)).toBeVisible();

		await page.goto("/status");
		await expect(
			page.getByRole("heading", { name: /Все системы работают|Обнаружены проблемы/i }),
		).toBeVisible({ timeout: 15_000 });
	});
});
