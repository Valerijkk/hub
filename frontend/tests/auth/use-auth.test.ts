import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as authApi from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

describe("useAuth", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
	});

	it("logout calls POST /auth/logout before clearing store", async () => {
		const logoutSpy = vi.spyOn(authApi, "logout").mockResolvedValue(undefined);
		useAuthStore.getState().setTokens("access-token", "refresh-token");

		const { result } = renderHook(() => useAuth());

		await act(async () => {
			await result.current.logout();
		});

		expect(logoutSpy).toHaveBeenCalledOnce();
		expect(useAuthStore.getState().isAuthenticated).toBe(false);
		expect(useAuthStore.getState().accessToken).toBeNull();
	});

	it("logout clears local session when API fails", async () => {
		vi.spyOn(authApi, "logout").mockRejectedValue(new Error("network"));
		useAuthStore.getState().setTokens("access-token", "refresh-token");

		const { result } = renderHook(() => useAuth());

		await act(async () => {
			await result.current.logout();
		});

		expect(useAuthStore.getState().isAuthenticated).toBe(false);
		expect(useAuthStore.getState().refreshToken).toBeNull();
	});
});
