import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "./fixtures";
import { authenticatePage, devLogin, DEV_EMAIL } from "./helpers/auth";

test.describe("Authenticated navigation smoke", () => {
	test("app shell sidebar links navigate to main sections", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await authenticatePage(page, login);
		await page.goto("/projects");

		await expect(page.getByRole("complementary")).toBeVisible();
		await expect(page.getByRole("main")).toBeVisible();

		const routes = [
			{ link: "Дашборд", url: /\/dashboard/ },
			{ link: "Все проекты", url: /\/projects/ },
			{ link: "Справка", url: /\/help/ },
			{ link: "Настройки", url: /\/settings\/profile/ },
		] as const;

		for (const route of routes) {
			await page.getByRole("link", { name: route.link }).click();
			await expect(page).toHaveURL(route.url, { timeout: 10_000 });
		}
	});

	test("unauthenticated user is redirected to login", async ({
		page,
		backendAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);

		await page.goto("/projects");
		await expect(page).toHaveURL(/\/login/, { timeout: 10_000 });
	});
});
