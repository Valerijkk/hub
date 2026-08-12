import { expect } from "@playwright/test";

import { BasePage, expectVisible } from "./base.page";

/** Page Object: модуль мессенджера. */
export class MessengerPage extends BasePage {
	readonly composer = this.page.getByPlaceholder(/Написать/i);
	readonly sendButton = this.page.getByTitle("Отправить");

	async openInbox(): Promise<void> {
		await this.goto("/messenger");
		await expect(this.page).toHaveURL(/\/messenger\/[0-9a-f-]+/, {
			timeout: 20_000,
		});
	}

	async sendMessage(text: string): Promise<void> {
		await expectVisible(this.composer, 15_000);
		await this.composer.fill(text);
		await this.sendButton.click();
		const sentMessage = this.page
			.getByRole("paragraph")
			.filter({ hasText: text })
			.last();
		await expect(sentMessage).toHaveText(text, { timeout: 15_000 });
	}
}
