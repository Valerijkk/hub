/**
 * Базовый Page Object: общая навигация и проверка видимости.
 */
import type { Page, Locator } from "@playwright/test";
import { expect } from "@playwright/test";

/** Базовый Page Object — общая навигация. */
export abstract class BasePage {
	constructor(protected readonly page: Page) {}

	async goto(path: string): Promise<void> {
		await this.page.goto(path);
	}

	protected locator(selector: string): Locator {
		return this.page.locator(selector);
	}
}

/** Хелпер проверки видимости (AAA — этап Assert). */
export async function expectVisible(loc: Locator, timeout?: number): Promise<void> {
	await expect(loc).toBeVisible(timeout ? { timeout } : undefined);
}
