import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as authApi from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

type Deferred<T> = {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
};

function deferred<T>(): Deferred<T> {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise;
		reject = rejectPromise;
	});
	return { promise, resolve, reject };
}

const authActions = {
	setTokens: useAuthStore.getState().setTokens,
	setUser: useAuthStore.getState().setUser,
	logout: useAuthStore.getState().logout,
	clearSessionLocal: useAuthStore.getState().clearSessionLocal,
};

function user(id: string): authApi.UserResponse {
	return {
		id,
		email: `${id}@hub.local`,
		system_role: "member",
		timezone: "UTC",
		language: "ru",
		onboarding_completed: true,
	};
}

function resetAuthStore(): void {
	useAuthStore.setState({
		...authActions,
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

describe("useAuth session races", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
	});

	it("does not set the old user when token A resolves after token B", async () => {
		const tokenA = deferred<authApi.UserResponse>();
		const tokenB = deferred<authApi.UserResponse>();
		const setUser = vi.fn(authActions.setUser);

		vi.spyOn(authApi, "getMe")
			.mockReturnValueOnce(tokenA.promise)
			.mockReturnValueOnce(tokenB.promise);
		useAuthStore.setState({ setUser });
		useAuthStore.getState().setTokens("token-a", "refresh-a");

		const { unmount } = renderHook(() => useAuth());

		await act(async () => {
			useAuthStore.getState().setTokens("token-b", "refresh-b");
		});

		await act(async () => {
			tokenA.resolve(user("user-a"));
			await Promise.resolve();
		});

		expect(setUser).not.toHaveBeenCalled();
		expect(useAuthStore.getState().user).toBeNull();

		await act(async () => {
			tokenB.resolve(user("user-b"));
			await Promise.resolve();
		});

		expect(setUser).toHaveBeenCalledOnce();
		expect(useAuthStore.getState().user?.id).toBe("user-b");
		unmount();
	});

	it("does not log out the current session when token A rejects after logout", async () => {
		const tokenA = deferred<authApi.UserResponse>();
		const storeLogout = vi.fn(authActions.logout);

		vi.spyOn(authApi, "getMe").mockReturnValue(tokenA.promise);
		useAuthStore.setState({ logout: storeLogout });
		useAuthStore.getState().setTokens("token-a", "refresh-a");

		const { unmount } = renderHook(() => useAuth());

		await act(async () => {
			useAuthStore.getState().logout();
		});
		storeLogout.mockClear();

		await act(async () => {
			tokenA.reject(new Error("expired token-a"));
			await Promise.resolve();
		});

		expect(storeLogout).not.toHaveBeenCalled();
		expect(useAuthStore.getState()).toMatchObject({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
		});
		unmount();
	});
});
