/**
 * E2E-компонент Hub: logout.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../../shared/fixtures";
import { AuthSteps } from "../../shared/steps";

// E2E-группа: Выход из системы.
test.describe("Logout", () => {
	// Проверяет: выход из профиля вызывает POST /api/auth/logout.
	test("profile logout calls POST /api/auth/logout", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.loginAsDev(page);

		const logoutRequest = page.waitForRequest(
			(req) =>
				req.method() === "POST" && req.url().includes("/api/auth/logout"),
		);

		await auth.logoutFromProfile(page);

		const request = await logoutRequest;
		expect(request.method()).toBe("POST");
		await auth.expectLoggedOut(page);
	});
});
