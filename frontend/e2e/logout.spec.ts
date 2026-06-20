import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "./fixtures";
import { authenticatePage, devLogin, DEV_EMAIL } from "./helpers/auth";

test.describe("Logout", () => {
	test("profile logout calls POST /api/auth/logout", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const login = await devLogin(DEV_EMAIL);
		await authenticatePage(page, login);

		await page.goto("/settings/profile");
		await expect(
			page.getByRole("button", { name: "Выйти из аккаунта" }),
		).toBeVisible();

		const logoutRequest = page.waitForRequest(
			(req) =>
				req.method() === "POST" && req.url().includes("/api/auth/logout"),
		);

		await page.getByRole("button", { name: "Выйти из аккаунта" }).click();

		const request = await logoutRequest;
		expect(request.method()).toBe("POST");
		await expect(page).toHaveURL(/\/login/, { timeout: 10_000 });

		const authRaw = await page.evaluate(() =>
			sessionStorage.getItem("hub-auth"),
		);
		const auth = JSON.parse(authRaw ?? "{}") as {
			state?: { isAuthenticated?: boolean };
		};
		expect(auth.state?.isAuthenticated).toBe(false);
	});
});
