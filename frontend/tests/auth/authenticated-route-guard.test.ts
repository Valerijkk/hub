import { isRedirect } from "@tanstack/react-router";
import { beforeEach, describe, expect, it } from "vitest";
import { Route as AuthenticatedRoute } from "@/routes/_authenticated";
import { useAuthStore } from "@/stores/auth-store";

function runGuard(): unknown {
	const guard = AuthenticatedRoute.options.beforeLoad;
	if (!guard) {
		throw new Error("Authenticated route must define beforeLoad");
	}
	return guard({} as never);
}

function redirectTo(error: unknown): string | undefined {
	if (!isRedirect(error)) return undefined;
	return (error as { options?: { to?: string } }).options?.to;
}

describe("authenticated route beforeLoad", () => {
	beforeEach(() => {
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
		});
	});

	it("redirects an unauthenticated visitor to /login", () => {
		let caught: unknown;

		try {
			runGuard();
		} catch (error) {
			caught = error;
		}

		expect(isRedirect(caught)).toBe(true);
		expect(redirectTo(caught)).toBe("/login");
	});

	it("allows an authenticated session before the profile is loaded", () => {
		useAuthStore.setState({
			accessToken: "access-token",
			refreshToken: "refresh-token",
			user: null,
			isAuthenticated: true,
		});

		expect(runGuard).not.toThrow();
	});
});
