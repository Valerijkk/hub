/**
 * TASK-128 / A3W10-5-003…004 — onboarding + login beforeLoad guards.
 *
 * Mutation: удалить beforeLoad → assert isRedirect FAIL.
 */
import { isRedirect } from "@tanstack/react-router";
import { beforeEach, describe, expect, it } from "vitest";
import type { UserResponse } from "@/api/auth";
import { Route as LoginRoute } from "@/routes/login";
import { Route as OnboardingRoute } from "@/routes/onboarding";
import { useAuthStore } from "@/stores/auth-store";

function baseUser(overrides: Partial<UserResponse> = {}): UserResponse {
	return {
		id: "u1",
		email: "u@hub.local",
		timezone: "Europe/Moscow",
		language: "ru",
		system_role: "member",
		onboarding_completed: true,
		...overrides,
	};
}

function redirectTo(err: unknown): string | undefined {
	if (!isRedirect(err)) return undefined;
	return (err as { options?: { to?: string } }).options?.to;
}

describe("A3W10-5-003 onboarding beforeLoad", () => {
	beforeEach(() => {
		useAuthStore.setState({
			isAuthenticated: false,
			user: null,
			accessToken: null,
		});
	});

	it("unauthenticated → redirect /login", () => {
		let caught: unknown;
		try {
			OnboardingRoute.options.beforeLoad?.({} as never);
		} catch (e) {
			caught = e;
		}
		expect(isRedirect(caught)).toBe(true);
		expect(redirectTo(caught)).toBe("/login");
	});

	it("authenticated + onboarding_completed → redirect /projects", () => {
		useAuthStore.setState({
			isAuthenticated: true,
			user: baseUser({ onboarding_completed: true }),
			accessToken: "t",
		});
		let caught: unknown;
		try {
			OnboardingRoute.options.beforeLoad?.({} as never);
		} catch (e) {
			caught = e;
		}
		expect(isRedirect(caught)).toBe(true);
		expect(redirectTo(caught)).toBe("/projects");
	});

	it("authenticated + incomplete onboarding → no redirect", () => {
		useAuthStore.setState({
			isAuthenticated: true,
			user: baseUser({ onboarding_completed: false }),
			accessToken: "t",
		});
		expect(() =>
			OnboardingRoute.options.beforeLoad?.({} as never),
		).not.toThrow();
	});
});

describe("A3W10-5-004 login beforeLoad", () => {
	beforeEach(() => {
		useAuthStore.setState({
			isAuthenticated: false,
			user: null,
			accessToken: null,
		});
	});

	it("unauthenticated → no redirect", () => {
		expect(() => LoginRoute.options.beforeLoad?.({} as never)).not.toThrow();
	});

	it("authenticated + completed → redirect /projects", () => {
		useAuthStore.setState({
			isAuthenticated: true,
			user: baseUser({ onboarding_completed: true }),
			accessToken: "t",
		});
		let caught: unknown;
		try {
			LoginRoute.options.beforeLoad?.({} as never);
		} catch (e) {
			caught = e;
		}
		expect(isRedirect(caught)).toBe(true);
		expect(redirectTo(caught)).toBe("/projects");
	});

	it("authenticated + incomplete → redirect /onboarding", () => {
		useAuthStore.setState({
			isAuthenticated: true,
			user: baseUser({ onboarding_completed: false }),
			accessToken: "t",
		});
		let caught: unknown;
		try {
			LoginRoute.options.beforeLoad?.({} as never);
		} catch (e) {
			caught = e;
		}
		expect(isRedirect(caught)).toBe(true);
		expect(redirectTo(caught)).toBe("/onboarding");
	});
});
