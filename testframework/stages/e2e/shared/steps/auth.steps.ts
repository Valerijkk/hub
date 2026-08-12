/**
 * Шаги для сценариев аутентификации и навигации (паттерн Step).
 */
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";

import { LoginPage, AppShellPage } from "../pages";
import {
	ADMIN_EMAIL,
	authenticatePage,
	devLogin,
	DEV_EMAIL,
	type DevLoginResult,
} from "../helpers/auth";

/**
 * AuthSteps — бизнес-шаги аутентификации (паттерн Step поверх POM).
 * Единственная ответственность: только вход (принцип SRP).
 */
export class AuthSteps {
	private readonly loginPage: LoginPage;
	private readonly shell: AppShellPage;

	constructor(page: Page) {
		this.loginPage = new LoginPage(page);
		this.shell = new AppShellPage(page);
	}

	async openLoginAndExpectBranding(): Promise<void> {
		await this.loginPage.open();
		await this.loginPage.expectBrandingLoaded();
	}

	async devLoginViaUI(buttonLabel: string): Promise<void> {
		await this.loginPage.open();
		await this.loginPage.devLoginAs(buttonLabel);
	}

	async devLoginViaAPI(email: string): Promise<DevLoginResult> {
		return devLogin(email);
	}

	async injectSession(page: Page, login: DevLoginResult): Promise<void> {
		await authenticatePage(page, login);
	}

	async loginAndInjectSession(
		page: Page,
		email: string,
	): Promise<DevLoginResult> {
		const login = await devLogin(email);
		await authenticatePage(page, login);
		return login;
	}

	async loginAsDev(page: Page): Promise<DevLoginResult> {
		return this.loginAndInjectSession(page, DEV_EMAIL);
	}

	async loginAsAdmin(page: Page): Promise<DevLoginResult> {
		return this.loginAndInjectSession(page, ADMIN_EMAIL);
	}

	async devLoginAsAdminUI(): Promise<void> {
		await this.devLoginViaUI("Админ Хабов");
	}

	async expectSessionInMemory(page: Page): Promise<void> {
		// A2W11-2-001: JWT не в sessionStorage; сессия живёт в memory-store.
		const authRaw = await page.evaluate(() =>
			sessionStorage.getItem("hub-auth"),
		);
		expect(authRaw).toBeNull();

		const hasTokensInStorage = await page.evaluate(() => {
			const blob = `${sessionStorage.getItem("hub-auth") ?? ""}${localStorage.getItem("hub-auth") ?? ""}`;
			return blob.includes("accessToken") || blob.includes("refreshToken");
		});
		expect(hasTokensInStorage).toBe(false);

		await expect(page).toHaveURL(/\/projects/, { timeout: 10_000 });
	}

	async expectLoggedOut(page: Page): Promise<void> {
		await expect(page).toHaveURL(/\/login/, { timeout: 10_000 });

		const authRaw = await page.evaluate(() =>
			sessionStorage.getItem("hub-auth"),
		);
		expect(authRaw).toBeNull();
	}

	async logoutFromProfile(page: Page): Promise<void> {
		await this.shell.openProfile();
		await expect(this.shell.logoutButton()).toBeVisible();
		await this.shell.logoutButton().click();
	}
}

/** NavigationSteps — навигация после успешного входа. */
export class NavigationSteps {
	private readonly shell: AppShellPage;

	constructor(page: Page) {
		this.shell = new AppShellPage(page);
	}

	async fromProjectsToDashboardAndSettings(): Promise<void> {
		await this.shell.gotoProjects();
		await this.shell.openDashboard();
		await this.shell.openSettings();
	}

	async smokeSidebarRoutes(
		routes: ReadonlyArray<{ link: string; url: RegExp }>,
	): Promise<void> {
		await this.shell.gotoProjects();
		await this.shell.expectShellVisible();

		for (const route of routes) {
			await this.shell.navigateTo(route.link);
			await this.shell.expectUrl(route.url);
		}
	}
}

export { ADMIN_EMAIL, DEV_EMAIL } from "../helpers/auth";
