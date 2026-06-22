import { beforeEach, describe, expect, it } from "vitest";
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

describe("auth-store", () => {
	beforeEach(() => {
		resetAuthStore();
	});

	it("setTokens stores tokens and sets isAuthenticated", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");

		const state = useAuthStore.getState();
		expect(state.accessToken).toBe("access-abc");
		expect(state.refreshToken).toBe("refresh-xyz");
		expect(state.isAuthenticated).toBe(true);
	});

	it("setUser stores user profile", () => {
		const user = {
			id: "user-1",
			email: "dev@hub.local",
			name: "Dev",
			display_name: "Dev",
			timezone: "UTC",
			language: "ru",
			system_role: "member",
			onboarding_completed: true,
		};

		useAuthStore.getState().setUser(user);

		expect(useAuthStore.getState().user).toEqual(user);
	});

	it("logout clears tokens and isAuthenticated", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");
		useAuthStore.getState().setUser({
			id: "user-1",
			email: "dev@hub.local",
			name: "Dev",
			display_name: "Dev",
			timezone: "UTC",
			language: "ru",
			system_role: "admin",
			onboarding_completed: true,
		});

		useAuthStore.getState().logout();

		const state = useAuthStore.getState();
		expect(state.accessToken).toBeNull();
		expect(state.refreshToken).toBeNull();
		expect(state.user).toBeNull();
		expect(state.isAuthenticated).toBe(false);
	});
});
