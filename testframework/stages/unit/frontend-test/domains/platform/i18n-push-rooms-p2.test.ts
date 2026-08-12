import { afterEach, describe, expect, it } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { formatRelativeTime } from "@/lib/format-time";
import {
	dismissPushBanner,
	PUSH_DISMISS_TTL_MS,
	shouldShowPushBanner,
} from "@/lib/push-notifications";

describe("formatRelativeTime i18n (P2-FE-15 / TASK-056)", () => {
	const now = Date.parse("2026-07-28T12:00:00.000Z");

	it("locale=en does not return Russian strings", () => {
		const past = new Date(now - 5 * 60_000).toISOString();
		const text = formatRelativeTime(past, { now, locale: "en" });
		expect(text).not.toMatch(/Только что|мин|Вчера|дн|назад/);
		expect(text.toLowerCase()).toMatch(/minute/);
	});

	it("future dates are not Just now / Только что", () => {
		const future = new Date(now + 10 * 60_000).toISOString();
		const ru = formatRelativeTime(future, { now, locale: "ru" });
		const en = formatRelativeTime(future, { now, locale: "en" });
		expect(ru).not.toMatch(/Только что|сейчас/i);
		expect(en.toLowerCase()).not.toMatch(/just now|^now$/);
		expect(ru).toMatch(/через/);
		expect(en.toLowerCase()).toMatch(/in /);
	});
});

describe("push dismiss TTL (P2-FE-15 / TASK-056)", () => {
	afterEach(() => {
		localStorage.clear();
	});

	function mockPushSupported() {
		Object.defineProperty(navigator, "serviceWorker", {
			value: {},
			configurable: true,
		});
		Object.defineProperty(window, "PushManager", {
			value: class PushManager {},
			configurable: true,
		});
		Object.defineProperty(window, "Notification", {
			value: { permission: "default" },
			configurable: true,
		});
	}

	it("dismissLater hides banner until TTL, then shows again", () => {
		mockPushSupported();
		expect(PUSH_DISMISS_TTL_MS).toBe(7 * 24 * 60 * 60 * 1000);

		const t0 = Date.parse("2026-07-01T00:00:00.000Z");
		dismissPushBanner();
		// overwrite with fixed timestamp
		localStorage.setItem("hub_push_dismissed", String(t0));

		expect(shouldShowPushBanner(t0 + 1000)).toBe(false);
		expect(shouldShowPushBanner(t0 + PUSH_DISMISS_TTL_MS - 1000)).toBe(false);
		expect(shouldShowPushBanner(t0 + PUSH_DISMISS_TTL_MS + 1000)).toBe(true);
		expect(localStorage.getItem("hub_push_dismissed")).toBeNull();
	});
});

describe("useRooms dead export removed (A2W11-4-002 / TASK-123)", () => {
	it("does not export useRooms (0 call sites → deleted)", () => {
		const src = readFileSync(
			resolve(
				__dirname,
				"../../../../../../frontend/src/hooks/use-websocket.ts",
			),
			"utf8",
		);
		expect(src).not.toMatch(/\bexport\s+function\s+useRooms\b/);
		expect(src).not.toMatch(/\broomsKey\b/);
		expect(src).toMatch(/\bexport\s+function\s+useWebSocket\b/);
		expect(src).toMatch(/\bexport\s+function\s+useWsEventListener\b/);
	});
});
