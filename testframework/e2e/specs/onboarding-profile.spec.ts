/**
 * E2E-компонент Hub: onboarding-profile.spec.ts.
 */
import {
	expect,
	skipUnlessBackend,
	skipUnlessDevMode,
	test,
} from "../fixtures";
import { AuthSteps } from "../steps";

// E2E-группа: Шаг онбординга «Профиль» (UX-P1-07 / FB-12).
test.describe("Onboarding profile step (UX-P1-07 / FB-12)", () => {
	// Проверяет: «Далее» заблокирована, пока имя невалидно.
	test("blocks Next until name is valid", async ({
		page,
		backendAvailable,
		devModeAvailable,
	}) => {
		skipUnlessBackend(backendAvailable);
		skipUnlessDevMode(devModeAvailable);

		const auth = new AuthSteps(page);
		const login = await auth.loginAsDev(page);
		await auth.injectSession(page, {
			...login,
			user: { ...login.user, onboarding_completed: false, name: "" },
		});

		await page.goto("/onboarding");
		await expect(
			page.getByText(/профил|profile/i, { exact: false }).first(),
		).toBeVisible({ timeout: 10_000 });

		const nextBtn = page.getByRole("button", { name: /далее|next/i });
		await expect(nextBtn).toBeDisabled();

		await page.locator("#onb-name").fill("Alex");
		await expect(nextBtn).toBeEnabled({ timeout: 5000 });

		await nextBtn.click();
		await expect(
			page.getByRole("heading", { name: /hub|знаком/i }),
		).toBeVisible({ timeout: 5000 });
	});
});
