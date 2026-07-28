import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { api, type ApiError } from "@/api/client";
import {
	canSuggestComposer,
	applyComposerSuggestion,
} from "@/lib/composer-ai-suggest";
import { useAuthStore } from "@/stores/auth-store";

const fe = (...parts: string[]) =>
	resolve(__dirname, "../../../../../../frontend/src", ...parts);

describe("TASK-077 P3-FE-10..12 api / auth / naming", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		useAuthStore.setState({
			accessToken: "tok",
			refreshToken: "ref",
			user: null,
			isAuthenticated: true,
		});
	});

	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it("2xx invalid JSON throws ApiError code 502", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn().mockResolvedValue({
				status: 200,
				ok: true,
				headers: new Headers(),
				text: async () => "{not-json",
				json: async () => {
					throw new Error("no");
				},
			}),
		);

		await expect(api.get("/ping")).rejects.toMatchObject({
			code: 502,
			message: "invalid json",
		} satisfies Partial<ApiError>);
	});

	it("broadcastAuth mirrors login to localStorage key", () => {
		const setSpy = vi.spyOn(Storage.prototype, "setItem");
		const removeSpy = vi.spyOn(Storage.prototype, "removeItem");

		useAuthStore.getState().setTokens("a", "b");

		expect(setSpy).toHaveBeenCalledWith(
			"hub-auth-event",
			expect.stringContaining('"type":"login"'),
		);
		expect(removeSpy).toHaveBeenCalledWith("hub-auth-event");
	});

	it("auth-store broadcasts userId from setTokens/setUser", () => {
		const src = readFileSync(fe("stores/auth-store.ts"), "utf8");
		expect(src).toMatch(
			/const userId = useAuthStore\.getState\(\)\.user\?\.id \?\? null/,
		);
		expect(src).toMatch(
			/broadcastAuth\(\{\s*type:\s*"login",\s*userId:\s*user\.id\s*\}\)/,
		);
		expect(src).toMatch(
			/localStorage\.setItem\(\s*`\$\{AUTH_CHANNEL_NAME\}-event`/,
		);

		const layout = readFileSync(fe("routes/_authenticated.tsx"), "utf8");
		expect(layout).toMatch(
			/event\.userId && currentId && event\.userId === currentId/,
		);
	});

	it("github.ts no longer owns /link-preview client", () => {
		const src = readFileSync(fe("api/github.ts"), "utf8");
		expect(src).not.toMatch(/link-preview/);
		expect(src).not.toMatch(/getLinkPreview/);
	});

	it("messenger owns fetchLinkPreview / getLinkPreview alias", () => {
		const src = readFileSync(fe("api/messenger.ts"), "utf8");
		expect(src).toMatch(/export function fetchLinkPreview/);
		expect(src).toMatch(/export const getLinkPreview = fetchLinkPreview/);
	});

	it("composer helpers live in lib/, not hooks/use-*", () => {
		expect(canSuggestComposer("12345678")).toBe(true);
		expect(applyComposerSuggestion("Hi", "there")).toBe("Hi there");
		const libSrc = readFileSync(fe("lib/composer-ai-suggest.ts"), "utf8");
		expect(libSrc).toMatch(/canSuggestComposer/);
		expect(() =>
			readFileSync(fe("hooks/use-composer-ai-suggest.ts"), "utf8"),
		).toThrow();
	});
});
