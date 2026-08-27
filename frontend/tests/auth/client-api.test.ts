import {
	afterEach,
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from "vitest";

const { toastError, toastDismiss } = vi.hoisted(() => ({
	toastError: vi.fn(),
	toastDismiss: vi.fn(),
}));

vi.mock("sonner", () => ({
	toast: {
		error: toastError,
		dismiss: toastDismiss,
	},
}));

import { downloadCalendarExport } from "@/api/calendar";
import { api, type ApiError } from "@/api/client";
import "@/i18n";
import {
	formatApiError,
	resetRateLimitToastForTests,
	toastApiError,
} from "@/lib/format-api-error";
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

const WIRE_REQUEST_ID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890";

function wireErrorResponse(
	status: number,
	body: { error: string; request_id: string },
	headers?: Record<string, string>,
): Response {
	return new Response(JSON.stringify(body), {
		status,
		statusText: status === 429 ? "Too Many Requests" : "Error",
		headers: {
			"Content-Type": "application/json",
			...headers,
		},
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

describe("api.postBytes", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
		useAuthStore.getState().setTokens("binary-access", "binary-refresh");
	});

	it("sends an owned non-JSON byte body with binary content type and preserves caller bytes", async () => {
		const callerBytes = new Uint8Array([0, 1, 42, 255]);
		const abortController = new AbortController();
		let ownedRequestBytes: Uint8Array | undefined;

		vi.stubGlobal(
			"fetch",
			vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => {
				const headers = init?.headers as Record<string, string> | undefined;
				expect(init?.method).toBe("POST");
				expect(init?.signal).toBe(abortController.signal);
				expect(headers?.Authorization).toBe("Bearer binary-access");
				expect(headers?.["Content-Type"]).toBe(
					"application/octet-stream",
				);

				const requestBody = init?.body;
				expect(requestBody).toBeInstanceOf(Uint8Array);
				expect(typeof requestBody).not.toBe("string");
				expect(requestBody).not.toBe(callerBytes);
				expect([...((requestBody as Uint8Array) ?? [])]).toEqual([
					0,
					1,
					42,
					255,
				]);
				ownedRequestBytes = requestBody as Uint8Array;
				return new Response(JSON.stringify({ accepted: true }), {
					status: 200,
					headers: { "Content-Type": "application/json" },
				});
			}),
		);

		const result = await api.postBytes<{ accepted: boolean }>(
			"/mail/setup/bytes",
			callerBytes,
			{ signal: abortController.signal },
		);

		expect(result).toEqual({ accepted: true });
		expect([...callerBytes]).toEqual([0, 1, 42, 255]);
		expect([...((ownedRequestBytes as Uint8Array) ?? [])]).toEqual([
			0,
			0,
			0,
			0,
		]);
	});

	it("rejects empty and oversized byte bodies before fetching", async () => {
		const fetchMock = vi.fn();
		vi.stubGlobal("fetch", fetchMock);

		await expect(api.postBytes("/mail/setup/bytes", new Uint8Array())).rejects.toThrow(
			"Invalid binary request body",
		);
		await expect(
			api.postBytes("/mail/setup/bytes", new Uint8Array(257)),
		).rejects.toThrow("Invalid binary request body");
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it("retries byte body once after refresh with a fresh owned copy", async () => {
		const callerBytes = new Uint8Array([7, 8, 9]);
		let byteCalls = 0;
		let refreshCalls = 0;
		const sentBodies: number[][] = [];
		const authorizations: string[] = [];

		vi.stubGlobal(
			"fetch",
			vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
				const url = String(input);
				if (url.endsWith("/api/auth/refresh")) {
					refreshCalls += 1;
					return new Response(
						JSON.stringify({
							access_token: "binary-access-refreshed",
							refresh_token: "binary-refresh-refreshed",
						}),
						{ status: 200, headers: { "Content-Type": "application/json" } },
					);
				}

				byteCalls += 1;
				const headers = init?.headers as Record<string, string> | undefined;
				authorizations.push(headers?.Authorization ?? "");
				sentBodies.push([...((init?.body as Uint8Array) ?? [])]);
				if (byteCalls === 1) return new Response(null, { status: 401 });
				return new Response(JSON.stringify({ retried: true }), {
					status: 200,
					headers: { "Content-Type": "application/json" },
				});
			}),
		);

		await expect(
			api.postBytes<{ retried: boolean }>("/mail/setup/bytes", callerBytes),
		).resolves.toEqual({ retried: true });
		expect(refreshCalls).toBe(1);
		expect(byteCalls).toBe(2);
		expect(authorizations).toEqual([
			"Bearer binary-access",
			"Bearer binary-access-refreshed",
		]);
		expect(sentBodies).toEqual([
			[7, 8, 9],
			[7, 8, 9],
		]);
		expect([...callerBytes]).toEqual([7, 8, 9]);
	});

	it("uses the same wire error mapping for byte body requests", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(async () =>
				wireErrorResponse(
					429,
					{ error: "rate limited", request_id: WIRE_REQUEST_ID },
					{ "Retry-After": "4" },
				),
			),
		);

		await expect(
			api.postBytes("/mail/setup/bytes", new Uint8Array([12])),
		).rejects.toMatchObject({
			code: 429,
			message: "rate limited",
			request_id: WIRE_REQUEST_ID,
			retry_after: 4,
		});
	});

	it("returns undefined for a 204 byte response", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(async () => new Response(null, { status: 204 })),
		);

		await expect(
			api.postBytes<undefined>("/mail/setup/bytes", new Uint8Array([5])),
		).resolves.toBeUndefined();
	});
});

describe("api client wire error → ApiError (A3W11-1-001)", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks();
		vi.useFakeTimers();
		toastError.mockClear();
		toastDismiss.mockClear();
		useAuthStore.getState().setTokens("wire-access", "wire-refresh");
	});

	afterEach(() => {
		resetRateLimitToastForTests();
		vi.useRealTimers();
	});

	it("request: maps TZ wire error→message + request_id; 429 Retry-After → toast countdown", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(async () =>
				wireErrorResponse(
					429,
					{ error: "rate limited", request_id: WIRE_REQUEST_ID },
					{ "Retry-After": "5" },
				),
			),
		);

		let thrown: unknown;
		try {
			await api.get("/rate-limited");
		} catch (err) {
			thrown = err;
		}

		const apiErr = thrown as ApiError;
		expect(apiErr.code).toBe(429);
		expect(apiErr.message).toBe("rate limited");
		expect(apiErr.request_id).toBe(WIRE_REQUEST_ID);
		expect(apiErr.retry_after).toBe(5);

		const formatted = formatApiError(apiErr);
		expect(formatted).toContain("5");
		expect(formatted.toLowerCase()).not.toContain("unknown");

		toastApiError(apiErr);
		expect(toastError).toHaveBeenCalledWith(
			expect.stringContaining("5"),
			expect.objectContaining({
				id: "hub-api-rate-limit",
				description: expect.stringContaining(WIRE_REQUEST_ID),
			}),
		);
	});

	it("request: 404 wire error reaches formatApiError with request_id toast", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(async () =>
				wireErrorResponse(404, {
					error: "not found",
					request_id: WIRE_REQUEST_ID,
				}),
			),
		);

		let thrown: unknown;
		try {
			await api.get("/missing");
		} catch (err) {
			thrown = err;
		}

		const apiErr = thrown as ApiError;
		expect(apiErr.code).toBe(404);
		expect(apiErr.message).toBe("not found");
		expect(apiErr.request_id).toBe(WIRE_REQUEST_ID);
		expect(formatApiError(apiErr)).toBe("not found");

		toastApiError(apiErr);
		expect(toastError).toHaveBeenCalledWith(
			"not found",
			expect.objectContaining({
				description: expect.stringContaining(WIRE_REQUEST_ID),
			}),
		);
	});

	it("getBlob: maps wire error→message and request_id", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(async () =>
				wireErrorResponse(403, {
					error: "export forbidden",
					request_id: WIRE_REQUEST_ID,
				}),
			),
		);

		let thrown: unknown;
		try {
			await api.getBlob("/calendar/export.ics?from=a&to=b");
		} catch (err) {
			thrown = err;
		}

		const apiErr = thrown as ApiError;
		expect(apiErr.code).toBe(403);
		expect(apiErr.message).toBe("export forbidden");
		expect(apiErr.request_id).toBe(WIRE_REQUEST_ID);
		expect(formatApiError(apiErr)).toBe("export forbidden");
	});

	it("request: invalid JSON on error path falls back to statusText", async () => {
		vi.stubGlobal(
			"fetch",
			vi.fn(
				async () =>
					new Response("not-json", {
						status: 502,
						statusText: "Bad Gateway",
						headers: { "Content-Type": "text/plain" },
					}),
			),
		);

		await expect(api.get("/broken")).rejects.toMatchObject({
			code: 502,
			message: "Bad Gateway",
		});
	});
});
