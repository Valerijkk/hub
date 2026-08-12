/**
 * E2E-компонент Hub: login.spec.ts.
 */
import { expect, skipUnlessBackend, skipUnlessDevMode, test } from "../../shared/fixtures";
import { AuthSteps } from "../../shared/steps";

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

	// После dev-login JWT не пишется в sessionStorage (memory-only).
	test("keeps JWT out of sessionStorage after dev login", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		await auth.devLoginAsAdminUI();
		await auth.expectSessionInMemory(page);
	});
});
