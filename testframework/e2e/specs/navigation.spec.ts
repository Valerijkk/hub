/**
 * E2E-компонент Hub: navigation.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AuthSteps, NavigationSteps } from "../steps";

// E2E-группа: Smoke навигации авторизованного пользователя.
test.describe("Authenticated navigation smoke", () => {
	// Проверяет: ссылки боковой панели ведут в основные разделы.
	test("app shell sidebar links navigate to main sections", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const nav = new NavigationSteps(page);
		await auth.loginAsDev(page);

		await nav.smokeSidebarRoutes([
			{ link: "Дашборд", url: /\/dashboard/ },
			{ link: "Все проекты", url: /\/projects/ },
			{ link: "Справка", url: /\/help/ },
			{ link: "Настройки", url: /\/settings\/profile/ },
		]);
	});

	// Проверяет: неавторизованный пользователь перенаправляется на /login.
	test("unauthenticated user is redirected to login", async ({
		page,
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		await page.goto("/projects");
		await expect(page).toHaveURL(/\/login/, { timeout: 10_000 });
	});
});
