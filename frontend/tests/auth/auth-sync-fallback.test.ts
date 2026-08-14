import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AuthSyncEvent } from "@/stores/auth-store";

type SubscribeAuthSync = (
	handler: (event: AuthSyncEvent) => void,
) => () => void;

const AUTH_STORAGE_KEY = "hub-auth-event";

function dispatchStorageEvent(value: string): void {
	window.dispatchEvent(
		new StorageEvent("storage", {
			key: AUTH_STORAGE_KEY,
			newValue: value,
			storageArea: localStorage,
		}),
	);
}

describe("auth sync storage fallback", () => {
	let subscribeAuthSync: SubscribeAuthSync;

	beforeEach(async () => {
		localStorage.clear();
		vi.resetModules();
		vi.stubGlobal("BroadcastChannel", undefined);

		({ subscribeAuthSync } = await import("@/stores/auth-store"));
	});

	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
		localStorage.clear();
	});

	it("delivers valid login and logout storage events when BroadcastChannel is unavailable", () => {
		const received: AuthSyncEvent[] = [];
		const unsubscribe = subscribeAuthSync((event) => received.push(event));

		dispatchStorageEvent(JSON.stringify({ type: "login", userId: "user-42" }));
		dispatchStorageEvent(JSON.stringify({ type: "logout" }));

		expect(received).toEqual([
			{ type: "login", userId: "user-42" },
			{ type: "logout" },
		]);
		unsubscribe();
	});

	it("ignores malformed storage payloads", () => {
		const handler = vi.fn<(event: AuthSyncEvent) => void>();
		const unsubscribe = subscribeAuthSync(handler);

		dispatchStorageEvent("{not-json");

		expect(handler).not.toHaveBeenCalled();
		unsubscribe();
	});

	it("stops delivering storage events after unsubscribe", () => {
		const handler = vi.fn<(event: AuthSyncEvent) => void>();
		const unsubscribe = subscribeAuthSync(handler);

		unsubscribe();
		dispatchStorageEvent(JSON.stringify({ type: "logout" }));

		expect(handler).not.toHaveBeenCalled();
	});
});
