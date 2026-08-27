import { describe, expect, it } from "vitest";
import {
	consumeMailOAuthFlow,
	isCanonicalMailOAuthState,
	rememberMailOAuthFlow,
	type MailOAuthFlow,
} from "@/modules/mail/mail-oauth-flow";

const STATE = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA";
const OTHER_STATE = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE";
const NOW_MS = 1_800_000_000_000;

class MemoryStorage implements Storage {
	private readonly values = new Map<string, string>();

	get length(): number {
		return this.values.size;
	}

	clear(): void {
		this.values.clear();
	}

	getItem(key: string): string | null {
		return this.values.get(key) ?? null;
	}

	key(index: number): string | null {
		return [...this.values.keys()][index] ?? null;
	}

	removeItem(key: string): void {
		this.values.delete(key);
	}

	setItem(key: string, value: string): void {
		this.values.set(key, value);
	}
}

class ThrowingStorage extends MemoryStorage {
	override getItem(_key: string): string | null {
		throw new Error("storage unavailable");
	}

	override setItem(_key: string, _value: string): void {
		throw new Error("storage unavailable");
	}

	override removeItem(_key: string): void {
		throw new Error("storage unavailable");
	}
}

function authorizationURL(state = STATE): string {
	return `https://mail.example.test/authorize?client_id=hub&state=${state}`;
}

function storedMarker(storage: Storage): [string, string] {
	const key = storage.key(0);
	if (!key) throw new Error("Expected exactly one stored marker");
	const value = storage.getItem(key);
	if (!value) throw new Error("Expected stored marker value");
	return [key, value];
}

describe("mail OAuth flow marker", () => {
	it("accepts only a canonical 32-byte unpadded base64url state", () => {
		expect(isCanonicalMailOAuthState(STATE)).toBe(true);
		expect(isCanonicalMailOAuthState(`${STATE}=`)).toBe(false);
		expect(isCanonicalMailOAuthState(STATE.replace(/A/, "+"))).toBe(false);
		expect(isCanonicalMailOAuthState("AQ")).toBe(false);
		expect(isCanonicalMailOAuthState("A".repeat(42))).toBe(false);
	});

	it.each([
		["non-local http", `http://mail.example.test/authorize?state=${STATE}`],
		["javascript", `javascript:alert(1)?state=${STATE}`],
		["relative", `/authorize?state=${STATE}`],
		["credentials", `https://user:secret@mail.example.test/authorize?state=${STATE}`],
	] as const)("rejects an unsafe %s authorization URL", (_name, url) => {
		const storage = new MemoryStorage();

		expect(rememberMailOAuthFlow(url, "setup", storage, NOW_MS)).toBe(false);
		expect(storage.length).toBe(0);
	});

	it.each<MailOAuthFlow>(["setup", "connection"])(
		"round-trips the %s flow once",
		(flow) => {
			const storage = new MemoryStorage();

			expect(rememberMailOAuthFlow(authorizationURL(), flow, storage, NOW_MS)).toBe(true);
			expect(consumeMailOAuthFlow(STATE, storage, NOW_MS + 1)).toBe(flow);
			expect(consumeMailOAuthFlow(STATE, storage, NOW_MS + 2)).toBeNull();
		},
	);

	it("keeps only one marker and retains a valid marker for a different callback state", () => {
		const storage = new MemoryStorage();

		expect(rememberMailOAuthFlow(authorizationURL(), "setup", storage, NOW_MS)).toBe(true);
		expect(rememberMailOAuthFlow(authorizationURL(OTHER_STATE), "connection", storage, NOW_MS + 1)).toBe(true);
		expect(storage.length).toBe(1);
		expect(consumeMailOAuthFlow(STATE, storage, NOW_MS + 2)).toBeNull();
		expect(consumeMailOAuthFlow(OTHER_STATE, storage, NOW_MS + 3)).toBe("connection");
	});

	it("removes expired and malformed markers without accepting them", () => {
		const expiredStorage = new MemoryStorage();
		expect(rememberMailOAuthFlow(authorizationURL(), "setup", expiredStorage, NOW_MS)).toBe(true);
		expect(consumeMailOAuthFlow(STATE, expiredStorage, NOW_MS + 11 * 60_000 + 1)).toBeNull();
		expect(expiredStorage.length).toBe(0);

		const malformedStorage = new MemoryStorage();
		malformedStorage.setItem("hub.mail.oauth-flow.v1", "{not-json");
		expect(consumeMailOAuthFlow(STATE, malformedStorage, NOW_MS)).toBeNull();
		expect(malformedStorage.length).toBe(0);
	});

	it("fails closed when session storage is unavailable", () => {
		const storage = new ThrowingStorage();

		expect(rememberMailOAuthFlow(authorizationURL(), "setup", storage, NOW_MS)).toBe(false);
		expect(consumeMailOAuthFlow(STATE, storage, NOW_MS)).toBeNull();
	});

	it("persists only the bounded local discriminator, never authorization URL, code, or token", () => {
		const storage = new MemoryStorage();
		const url = `${authorizationURL()}&code=never-store&access_token=never-store`;

		expect(rememberMailOAuthFlow(url, "setup", storage, NOW_MS)).toBe(true);
		const [key, serialized] = storedMarker(storage);
		expect(key).toBe("hub.mail.oauth-flow.v1");
		expect(JSON.parse(serialized)).toEqual({
			version: 1,
			state: STATE,
			flow: "setup",
			created_at_ms: NOW_MS,
		});
		expect(serialized).not.toContain(url);
		expect(serialized).not.toContain("never-store");
	});
});
