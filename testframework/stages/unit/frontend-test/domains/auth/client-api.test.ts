import { beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/api/client";
import { useAuthStore } from "@/stores/auth-store";

function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: "stale-access",
		refreshToken: "refresh-token",
		user: null,
		isAuthenticated: true,
	});
}

describe("api client", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
	});

	it("refresh single-flight: parallel 401s trigger one refresh request", async () => {
		let refreshCalls = 0;

		const fetchMock = vi.fn(
			async (input: RequestInfo | URL, init?: RequestInit) => {
				const url = String(input);

				if (url.endsWith("/api/auth/refresh")) {
					refreshCalls += 1;
					await new Promise((resolve) => setTimeout(resolve, 30));
					return new Response(
						JSON.stringify({
							access_token: "new-access",
							refresh_token: "new-refresh",
						}),
						{ status: 200, headers: { "Content-Type": "application/json" } },
					);
				}

				if (url.includes("/api/protected/")) {
					const headers = init?.headers as Record<string, string> | undefined;
					if (headers?.Authorization === "Bearer new-access") {
						return new Response(JSON.stringify({ ok: true }), {
							status: 200,
							headers: { "Content-Type": "application/json" },
						});
					}
					return new Response("", { status: 401 });
				}

				return new Response("", { status: 404 });
			},
		);

		vi.stubGlobal("fetch", fetchMock);

		const [a, b, c] = await Promise.all([
			api.get<{ ok: boolean }>("/protected/a"),
			api.get<{ ok: boolean }>("/protected/b"),
			api.get<{ ok: boolean }>("/protected/c"),
		]);

		expect(refreshCalls).toBe(1);
		expect(a).toEqual({ ok: true });
		expect(b).toEqual({ ok: true });
		expect(c).toEqual({ ok: true });
		expect(useAuthStore.getState().accessToken).toBe("new-access");
	});

	it("attaches Authorization header when access token is present", async () => {
		useAuthStore.getState().setTokens("my-access", "my-refresh");

		const fetchMock = vi.fn(
			async (_input: RequestInfo | URL, init?: RequestInit) => {
				const headers = init?.headers as Record<string, string> | undefined;
				expect(headers?.Authorization).toBe("Bearer my-access");
				return new Response(JSON.stringify({ ok: true }), {
					status: 200,
					headers: { "Content-Type": "application/json" },
				});
			},
		);
		vi.stubGlobal("fetch", fetchMock);

		await api.get<{ ok: boolean }>("/protected/me");

		expect(fetchMock).toHaveBeenCalledOnce();
	});

	it("refresh failure logs out the session", async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url.endsWith("/api/auth/refresh")) {
				return new Response("", { status: 401 });
			}
			return new Response("", { status: 401 });
		});
		vi.stubGlobal("fetch", fetchMock);

		await expect(api.get("/protected/x")).rejects.toThrow("Session expired");

		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(false);
		expect(state.accessToken).toBeNull();
		expect(state.refreshToken).toBeNull();
	});
});
