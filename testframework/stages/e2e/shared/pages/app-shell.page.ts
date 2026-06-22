/**
 * Page Object: авторизованная оболочка приложения (боковая панель).
 */
import { expect } from "@playwright/test";
import type { Locator } from "@playwright/test";

import { BasePage } from "./base.page";

/** Page Object: авторизованная оболочка приложения / боковая панель. */
export class AppShellPage extends BasePage {
	readonly complementary = this.page.getByRole("complementary");
	readonly main = this.page.getByRole("main");

	link(name: string | RegExp): Locator {
		return this.page.getByRole("link", { name });
	}

	async navigateTo(linkName: string): Promise<void> {
		await this.link(linkName).click();
	}

	async expectUrl(pattern: RegExp, timeout = 10_000): Promise<void> {
		await expect(this.page).toHaveURL(pattern, { timeout });
	}

	async expectShellVisible(): Promise<void> {
		await expect(this.complementary).toBeVisible();
		await expect(this.main).toBeVisible();
	}

	async gotoProjects(): Promise<void> {
		await this.goto("/projects");
	}

	async openDashboard(): Promise<void> {
		await this.navigateTo("Дашборд");
		await this.expectUrl(/\/dashboard/);
	}

	async openSettings(): Promise<void> {
		await this.navigateTo("Настройки");
		await this.expectUrl(/\/settings\/profile/);
	}

	async openProfile(): Promise<void> {
		await this.goto("/settings/profile");
		await this.expectUrl(/\/settings\/profile/);
	}

	logoutButton(): Locator {
		return this.page.getByRole("button", { name: "Выйти из аккаунта" });
	}

	/** Открывает seed-проект Hub Demo из списка проектов; возвращает id проекта. */
	async openDemoProject(): Promise<string> {
		await this.gotoProjects();
		await this.link(/Hub Demo/i).first().click();
		await this.expectUrl(/\/projects\/[0-9a-f-]+/);

		const projectUrl = this.page.url();
		const projectId = projectUrl.split("/projects/")[1]?.split(/[/?#]/)[0];
		expect(projectId).toBeTruthy();
		return projectId ?? "";
	}

	async openCommandPalette(): Promise<void> {
		await this.page.keyboard.press("Control+k");
		await expect(this.page.getByRole("dialog")).toBeVisible({ timeout: 5000 });
		await expect(this.page.getByPlaceholder(/поиск|search/i)).toBeVisible();
	}
}
