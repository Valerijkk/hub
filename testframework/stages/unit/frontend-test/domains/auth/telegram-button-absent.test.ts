import { existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import en from "@/i18n/locales/en/common.json";
import ru from "@/i18n/locales/ru/common.json";

/**
 * A11-003: Telegram login UI must stay deleted (Magic Link only / ADR-AUTH-01).
 * Old component called POST /auth/telegram-code (no BE route → always 404).
 */
describe("TelegramButton absent (A11-003)", () => {
	it("does not ship telegram-button.tsx under components/auth", () => {
		const here = dirname(fileURLToPath(import.meta.url));
		// domains/auth → … → repo root (6 levels)
		const repoRoot = join(here, "..", "..", "..", "..", "..", "..");
		const deadPath = join(
			repoRoot,
			"frontend",
			"src",
			"components",
			"auth",
			"telegram-button.tsx",
		);
		expect(existsSync(deadPath)).toBe(false);
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
