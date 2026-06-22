/**
 * Page Object: страница /login.
 */
import { expect } from "@playwright/test";

import { BasePage, expectVisible } from "./base.page";

/** Page Object: страница /login (POM). */
export class LoginPage extends BasePage {
	readonly emailInput = this.page.getByLabel("Email");
	readonly hubTitle = this.page.getByText("Hub").first();
	readonly tagline = this.page.getByText("Единая платформа для команд");
	readonly magicHint = this.page.getByText(
		"Введите email — мы отправим одноразовый код для входа.",
	);
	readonly devPanel = this.page.getByText("Демо-доступ");

	async open(): Promise<void> {
		await this.goto("/login");
		await expect(this.page).toHaveURL(/\/login/);
	}

	async expectBrandingLoaded(): Promise<void> {
		await expectVisible(this.hubTitle);
		await expectVisible(this.tagline);
		await expectVisible(this.magicHint);
		await expectVisible(this.emailInput);
	}

	async devLoginAs(buttonName: string): Promise<void> {
		await expectVisible(this.devPanel, 15_000);
		await this.page.getByRole("button", { name: buttonName }).click();
		await expect(this.page).toHaveURL(/\/projects/, { timeout: 15_000 });
		await expectVisible(this.hubTitle);
	}

	devLoginButton(name: string) {
		return this.page.getByRole("button", { name });
	}
}
