import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { type Message, searchMessages } from "@/api/messenger";
import { useSearchDialogSession } from "@/components/messenger/search-dialog-session";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({ searchMessages: vi.fn() }));
vi.mock("@/api/messenger", () => ({ searchMessages: api.searchMessages }));
const userA = {
	id: "user-a", email: "user-a@example.test", timezone: "UTC", language: "ru",
	system_role: "member", onboarding_completed: true,
};
const clients = new Set<QueryClient>();
type Deferred<T> = {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
};
function message(content: string, channelId = "channel-a"): Message {
	return {
		id: `message-${content}`, channel_id: channelId, sender_id: "sender-a",
		content_text: content, reply_count: 0, created_at: "2026-09-06T00:00:00Z",
	};
}
function deferred<T>(): Deferred<T> {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((done, fail) => { resolve = done; reject = fail; });
	return { promise, resolve, reject };
}
function client() {
	return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}
function renderSession(queryClient = client()) {
	clients.add(queryClient);
	const wrapper = ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
	);
	return {
		queryClient,
		...renderHook(() => useSearchDialogSession({ userId: "user-a", channelId: "channel-a" }), { wrapper }),
	};
}
function scopedQueries(queryClient: QueryClient) {
	return queryClient.getQueryCache().findAll({ queryKey: ["messenger", "search-dialog"] });
}

describe("useSearchDialogSession", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(searchMessages).mockReset().mockResolvedValue([]);
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: true, user: userA });
		useMessengerStore.getState().reset();
	});
	afterEach(() => {
		cleanup();
		for (const queryClient of clients) queryClient.clear();
		clients.clear();
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: false, user: null });
		useMessengerStore.getState().reset();
	});

	it("does not fetch while trimming or changing scope before an explicit submit", async () => {
		const session = renderSession();
		act(() => {
			session.result.current.setDraft("  needle  ");
			session.result.current.setScope("global");
		});
		expect(searchMessages).not.toHaveBeenCalled();
		act(() => session.result.current.submit());
		await waitFor(() => expect(searchMessages).toHaveBeenCalledWith("needle", undefined));
	});

	it("removes only the discarded attempt while its successor stays pending", async () => {
		const oldRequest = deferred<Message[]>();
		const newRequest = deferred<Message[]>();
		vi.mocked(searchMessages)
			.mockImplementationOnce(() => oldRequest.promise)
			.mockImplementationOnce(() => newRequest.promise);
		const session = renderSession();
		act(() => { session.result.current.setDraft("old"); session.result.current.submit(); });
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
		act(() => { session.result.current.setDraft("new"); session.result.current.submit(); });
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(scopedQueries(session.queryClient)).toHaveLength(1));
		await act(async () => {
			oldRequest.resolve([message("old")]);
			await oldRequest.promise;
		});
		expect(session.result.current.pending).toBe(true);
		expect(session.result.current.results).toEqual([]);
		await act(async () => {
			newRequest.resolve([message("new")]);
			await newRequest.promise;
		});
		await waitFor(() => expect(session.result.current.results).toEqual([message("new")]));
	});

	it.each(["new submit", "scope change", "session disposal"] as const)(
		"denies a retained successful selection guard after %s",
		async (boundary) => {
			vi.mocked(searchMessages).mockResolvedValueOnce([message("old")]);
			const session = renderSession();
			act(() => { session.result.current.setDraft("old"); session.result.current.submit(); });
			await waitFor(() => expect(session.result.current.results).toEqual([message("old")]));
			const retainedGuard = session.result.current.currentSelectionGuard;
			expect(retainedGuard).not.toBeNull();
			if (!retainedGuard) throw new Error("selection guard was not retained");
			if (boundary === "new submit") {
				act(() => { session.result.current.setDraft("new"); session.result.current.submit(); });
			} else if (boundary === "scope change") {
				act(() => session.result.current.setScope("global"));
			} else {
				session.unmount();
			}
			expect(retainedGuard(message("old"))).toBe(false);
		},
	);

	it("denies global submit after the current channel is revoked", () => {
		const session = renderSession();
		act(() => {
			session.result.current.setScope("global");
			session.result.current.setDraft("needle");
			useMessengerStore.getState().revokeChannel("channel-a");
			session.result.current.submit();
		});
		expect(searchMessages).not.toHaveBeenCalled();
		expect(session.result.current.pending).toBe(false);
	});

	it("fences a committed current-channel revoke and its late completion", async () => {
		const request = deferred<Message[]>();
		vi.mocked(searchMessages).mockReturnValueOnce(request.promise);
		const session = renderSession();
		act(() => {
			session.result.current.setScope("global");
			session.result.current.setDraft("needle");
			session.result.current.submit();
		});
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
		act(() => useMessengerStore.getState().revokeChannel("channel-a"));
		await waitFor(() => expect(session.result.current.pending).toBe(false));
		expect(session.result.current.results).toEqual([]);
		await act(async () => { request.resolve([message("late")]); await request.promise; });
		expect(session.result.current.results).toEqual([]);
	});

	it("removes only its session queries on unmount and preserves unrelated cache", async () => {
		const request = deferred<Message[]>();
		vi.mocked(searchMessages).mockReturnValueOnce(request.promise);
		const session = renderSession();
		session.queryClient.setQueryData(["unrelated"], "preserved");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			act(() => { session.result.current.setDraft("discard"); session.result.current.submit(); });
			await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
			session.unmount();
			await waitFor(() => expect(scopedQueries(session.queryClient)).toHaveLength(0));
			expect(session.queryClient.getQueryData(["unrelated"])).toBe("preserved");
			await act(async () => { request.reject(new Error("discarded")); await Promise.resolve(); });
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it("removes live-revoked result channels while retaining valid result rows", async () => {
		vi.mocked(searchMessages).mockResolvedValueOnce([
			message("revoked", "channel-b"), message("valid", "channel-c"),
		]);
		const session = renderSession();
		act(() => {
			session.result.current.setScope("global");
			session.result.current.setDraft("needle");
			session.result.current.submit();
		});
		await waitFor(() => expect(session.result.current.results).toEqual([
			message("revoked", "channel-b"), message("valid", "channel-c"),
		]));
		act(() => useMessengerStore.getState().revokeChannel("channel-b"));
		await waitFor(() => expect(session.result.current.results).toEqual([message("valid", "channel-c")]));
		expect(session.result.current.successEmpty).toBe(false);
		expect(session.result.current.currentSelectionGuard?.(message("revoked", "channel-b"))).toBe(false);
		expect(session.result.current.currentSelectionGuard?.(message("valid", "channel-c"))).toBe(true);
	});

	it("starts a fresh explicit attempt after a successful identical submit", async () => {
		vi.mocked(searchMessages).mockResolvedValueOnce([message("first")]).mockResolvedValueOnce([message("second")]);
		const session = renderSession();
		act(() => { session.result.current.setDraft("needle"); session.result.current.submit(); });
		await waitFor(() => expect(session.result.current.results).toEqual([message("first")]));
		act(() => session.result.current.submit());
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(session.result.current.results).toEqual([message("second")]));
	});

	it("deduplicates a pending submit and retries a current failed attempt once", async () => {
		const failure = deferred<Message[]>();
		vi.mocked(searchMessages).mockImplementationOnce(() => failure.promise).mockResolvedValueOnce([message("recovered")]);
		const session = renderSession();
		act(() => {
			session.result.current.setDraft("needle");
			session.result.current.submit();
			session.result.current.submit();
		});
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
		await act(async () => { failure.reject(new Error("failed")); await failure.promise.catch(() => undefined); });
		await waitFor(() => {
			expect(session.result.current.error).toBe(true);
			expect(session.result.current.pending).toBe(false);
		});
		act(() => session.result.current.retry());
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(session.result.current.results).toEqual([message("recovered")]));
	});

	it("does not let a retained failed retry restart a later successful attempt", async () => {
		const failure = deferred<Message[]>();
		vi.mocked(searchMessages).mockImplementationOnce(() => failure.promise).mockResolvedValueOnce([message("second")]);
		const session = renderSession();
		act(() => { session.result.current.setDraft("first"); session.result.current.submit(); });
		await waitFor(() => expect(searchMessages).toHaveBeenCalledTimes(1));
		await act(async () => { failure.reject(new Error("first failed")); await failure.promise.catch(() => undefined); });
		await waitFor(() => expect(session.result.current.error).toBe(true));
		const retainedRetry = session.result.current.retry;
		act(() => { session.result.current.setDraft("second"); session.result.current.submit(); });
		await waitFor(() => expect(session.result.current.results).toEqual([message("second")]));
		act(() => retainedRetry());
		expect(searchMessages).toHaveBeenCalledTimes(2);
		expect(session.result.current.results).toEqual([message("second")]);
	});
});
