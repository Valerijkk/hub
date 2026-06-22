/**
 * E2E-компонент Hub: login.spec.ts.
 */
import { expect, skipUnlessBackend, skipUnlessDevMode, test } from "../fixtures";
import { AuthSteps } from "../steps";

// E2E-группа: Сценарии входа.
test.describe("Login flow", () => {
	// Страница входа отображает брендинг Hub.
	test("login page loads with Hub branding", async ({ page }) => {
		const auth = new AuthSteps(page);
		await auth.openLoginAndExpectBranding();
	});

	// Dev-login через UI аутентифицирует и ведёт на /projects.
	test("dev login panel authenticates and redirects to projects", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.devLoginAsAdminUI();
	});

	// После dev-login сессия сохраняется в sessionStorage.
	test("stores auth session in sessionStorage after dev login", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.devLoginAsAdminUI();
		await auth.expectSessionInStorage(page);
	});
});
