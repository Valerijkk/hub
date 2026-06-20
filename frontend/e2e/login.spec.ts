import { expect, skipUnlessBackend, skipUnlessDevMode, test } from "./fixtures";
import { ADMIN_EMAIL } from "./helpers/auth";

test.describe("Login flow", () => {
	test("login page loads with Hub branding", async ({ page }) => {
		await page.goto("/login");

		await expect(page).toHaveURL(/\/login/);
		await expect(page.getByText("Hub").first()).toBeVisible();
		await expect(
			page.getByText("Единая платформа для команд"),
		).toBeVisible();
		await expect(
			page.getByText("Войдите через Telegram — быстро и безопасно."),
		).toBeVisible();
	});

	test("dev login panel authenticates and redirects to projects", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		await page.goto("/login");

		await expect(page.getByText("Демо-доступ")).toBeVisible({
			timeout: 15_000,
		});

		await page.getByRole("button", { name: "Админ Хабов" }).click();

		await expect(page).toHaveURL(/\/projects/, { timeout: 15_000 });
		await expect(page.getByText("Hub").first()).toBeVisible();
	});

	test("stores auth session in sessionStorage after dev login", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		await page.goto("/login");
		await page.getByRole("button", { name: "Админ Хабов" }).click();
		await expect(page).toHaveURL(/\/projects/);

		const authRaw = await page.evaluate(() =>
			sessionStorage.getItem("hub-auth"),
		);
		expect(authRaw).toBeTruthy();

		const auth = JSON.parse(authRaw ?? "{}") as {
			state?: { isAuthenticated?: boolean; user?: { email?: string } };
		};
		expect(auth.state?.isAuthenticated).toBe(true);
		expect(auth.state?.user?.email).toBe(ADMIN_EMAIL);
	});
});
