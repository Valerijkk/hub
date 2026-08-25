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
	let useAuthStore: typeof import("@/stores/auth-store").useAuthStore;

	beforeEach(async () => {
		localStorage.clear();
		vi.resetModules();
		vi.stubGlobal("BroadcastChannel", undefined);

		({ subscribeAuthSync, useAuthStore } = await import("@/stores/auth-store"));
	});

	afterEach(() => {
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
		vi.resetModules();
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
	it("clearSessionLocal never writes the storage fallback", () => {
		const setItemSpy = vi.spyOn(Storage.prototype, "setItem");
		useAuthStore.setState({
			accessToken: "access-remote",
			refreshToken: "refresh-remote",
			isAuthenticated: true,
		});
		setItemSpy.mockClear();

		useAuthStore.getState().clearSessionLocal();

		expect(setItemSpy).not.toHaveBeenCalled();
	});

	it("logout writes one valid logout event to the storage fallback", () => {
		const setItemSpy = vi.spyOn(Storage.prototype, "setItem");
		useAuthStore.setState({
			accessToken: "access-local",
			refreshToken: "refresh-local",
			isAuthenticated: true,
		});
		setItemSpy.mockClear();

		useAuthStore.getState().logout();

		expect(setItemSpy).toHaveBeenCalledTimes(1);
		expect(setItemSpy).toHaveBeenCalledWith(
			"hub-auth-event",
			JSON.stringify({ type: "logout" }),
		);
	});
});
