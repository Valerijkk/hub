import { describe, expect, it } from "vitest";
import en from "@/i18n/locales/en/common.json";
import ru from "@/i18n/locales/ru/common.json";

/**
 * A11-003: Telegram login UI must stay deleted (Magic Link only / ADR-AUTH-01).
 * Old component called POST /auth/telegram-code (no BE route → always 404).
 */
describe("TelegramButton absent (A11-003)", () => {
	it("does not resolve telegram-button module", async () => {
		await expect(
			import("@/components/auth/telegram-button"),
		).rejects.toBeTruthy();
	});

	it("documents Magic Link-only auth copy in i18n (no Telegram login CTA)", () => {
		expect(ru.auth.magicLinkOnly).toMatch(/Magic Link/i);
		expect(en.auth.magicLinkOnly).toMatch(/Magic Link/i);
		expect(ru.auth.telegramLoginRemoved).toBeTruthy();
		expect(en.auth.telegramLoginRemoved).toBeTruthy();
		expect(JSON.stringify(ru.auth)).not.toMatch(/войти через telegram/i);
		expect(JSON.stringify(en.auth)).not.toMatch(/sign in with telegram/i);
	});
});
