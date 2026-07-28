import { beforeEach, describe, expect, it, vi } from "vitest";
import { downloadCalendarExport } from "@/api/calendar";
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

	it("refresh failure best-effort POSTs /auth/logout then clears session", async () => {
		const logoutBodies: string[] = [];
		const fetchMock = vi.fn(
			async (input: RequestInfo | URL, init?: RequestInit) => {
				const url = String(input);
				if (url.endsWith("/api/auth/refresh")) {
					return new Response("", { status: 401 });
				}
				if (url.endsWith("/api/auth/logout")) {
					logoutBodies.push(String(init?.method ?? "GET"));
					const headers = init?.headers as Record<string, string> | undefined;
					expect(headers?.Authorization).toBe("Bearer stale-access");
					return new Response(JSON.stringify({ message: "Logged out" }), {
						status: 200,
						headers: { "Content-Type": "application/json" },
					});
				}
				return new Response("", { status: 401 });
			},
		);
		vi.stubGlobal("fetch", fetchMock);

		await expect(api.get("/protected/x")).rejects.toThrow("Session expired");

		expect(logoutBodies).toEqual(["POST"]);
		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(false);
		expect(state.accessToken).toBeNull();
		expect(state.refreshToken).toBeNull();
	});

	it("refresh failure logs out the session", async () => {
		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url.endsWith("/api/auth/refresh")) {
				return new Response("", { status: 401 });
			}
			if (url.endsWith("/api/auth/logout")) {
				return new Response("", { status: 200 });
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

	it("one-retry: second 401 after successful refresh logs out without another refresh", async () => {
		let refreshCalls = 0;
		let protectedCalls = 0;

		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url.endsWith("/api/auth/refresh")) {
				refreshCalls += 1;
				return new Response(
					JSON.stringify({
						access_token: "new-access",
						refresh_token: "new-refresh",
					}),
					{ status: 200, headers: { "Content-Type": "application/json" } },
				);
			}
			if (url.endsWith("/api/auth/logout")) {
				return new Response("", { status: 200 });
			}
			if (url.includes("/api/forbidden")) {
				protectedCalls += 1;
				return new Response("", { status: 401 });
			}
			return new Response("", { status: 404 });
		});
		vi.stubGlobal("fetch", fetchMock);

		await expect(api.get("/forbidden")).rejects.toThrow("Session expired");

		expect(refreshCalls).toBe(1);
		expect(protectedCalls).toBe(2);
		expect(useAuthStore.getState().isAuthenticated).toBe(false);
	});

	it("getBlob one-retry: second 401 after refresh does not loop", async () => {
		let refreshCalls = 0;
		let blobCalls = 0;

		const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
			const url = String(input);
			if (url.endsWith("/api/auth/refresh")) {
				refreshCalls += 1;
				return new Response(
					JSON.stringify({
						access_token: "new-access",
						refresh_token: "new-refresh",
					}),
					{ status: 200, headers: { "Content-Type": "application/json" } },
				);
			}
			if (url.endsWith("/api/auth/logout")) {
				return new Response("", { status: 200 });
			}
			if (url.includes("/api/calendar/export.ics")) {
				blobCalls += 1;
				return new Response("", { status: 401 });
			}
			return new Response("", { status: 404 });
		});
		vi.stubGlobal("fetch", fetchMock);

		await expect(
			api.getBlob("/calendar/export.ics?from=a&to=b"),
		).rejects.toThrow("Session expired");

		expect(refreshCalls).toBe(1);
		expect(blobCalls).toBe(2);
		expect(useAuthStore.getState().isAuthenticated).toBe(false);
	});
});

describe("downloadCalendarExport via api.getBlob", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
		useAuthStore.getState().setTokens("cal-access", "cal-refresh");
	});

	it("sends Bearer and downloads blob without raw fetch", async () => {
		const createObjectURL = vi.fn(() => "blob:hub-cal");
		const revokeObjectURL = vi.fn();
		vi.stubGlobal("URL", {
			...URL,
			createObjectURL,
			revokeObjectURL,
		});

		const click = vi.fn();
		const createElementSpy = vi
			.spyOn(document, "createElement")
			.mockImplementation((tag: string) => {
				if (tag === "a") {
					return {
						href: "",
						download: "",
						click,
					} as unknown as HTMLAnchorElement;
				}
				return document.createElement(tag);
			});

		const fetchMock = vi.fn(
			async (_input: RequestInfo | URL, init?: RequestInit) => {
				const headers = init?.headers as Record<string, string> | undefined;
				expect(headers?.Authorization).toBe("Bearer cal-access");
				return new Response("BEGIN:VCALENDAR", {
					status: 200,
					headers: { "Content-Type": "text/calendar" },
				});
			},
		);
		vi.stubGlobal("fetch", fetchMock);

		await downloadCalendarExport("2026-01-01", "2026-01-31");

		expect(fetchMock).toHaveBeenCalledOnce();
		const url = String(fetchMock.mock.calls[0]?.[0]);
		expect(url).toContain("/api/calendar/export.ics");
		expect(click).toHaveBeenCalledOnce();
		createElementSpy.mockRestore();
	});
});

describe("import-export FormData via api.post", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
		useAuthStore.getState().setTokens("import-access", "import-refresh");
	});

	it("posts FormData with Authorization Bearer", async () => {
		const fetchMock = vi.fn(
			async (_input: RequestInfo | URL, init?: RequestInit) => {
				const headers = init?.headers as Record<string, string> | undefined;
				expect(headers?.Authorization).toBe("Bearer import-access");
				expect(headers?.["Content-Type"]).toBeUndefined();
				expect(init?.body).toBeInstanceOf(FormData);
				return new Response(
					JSON.stringify({ headers: ["title"], rows: [], total: 0 }),
					{ status: 200, headers: { "Content-Type": "application/json" } },
				);
			},
		);
		vi.stubGlobal("fetch", fetchMock);

		const form = new FormData();
		form.append("file", new Blob(["a,b"]), "issues.csv");
		const data = await api.post<{ headers: string[]; total: number }>(
			"/projects/proj-1/import/issues?mode=preview",
			form,
		);

		expect(data.headers).toEqual(["title"]);
		expect(String(fetchMock.mock.calls[0]?.[0])).toContain(
			"/api/projects/proj-1/import/issues?mode=preview",
		);
	});
});
