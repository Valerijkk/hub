import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { getBookmarks, removeBookmark, type Bookmark } from "@/api/messenger";
import { BookmarksPanel, bookmarksQueryKey } from "@/components/messenger/bookmarks-panel";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ getBookmarks: vi.fn(), removeBookmark: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));
const navigate = vi.hoisted(() => vi.fn());

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, getBookmarks: api.getBookmarks, removeBookmark: api.removeBookmark };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@tanstack/react-router", () => ({ useNavigate: () => navigate }));

const REMOVE_LABEL = "Удалить из закладок";
const ERROR_RU = "Не удалось удалить закладку";
const userA = { id: "bookmark-removal-user-a", email: "bookmark-removal-a@example.test", timezone: "UTC", language: "en", system_role: "member", onboarding_completed: true };
const userB = { ...userA, id: "bookmark-removal-user-b", email: "bookmark-removal-b@example.test" };
const queryClients: QueryClient[] = [];

type Deferred<T> = { promise: Promise<T>; resolve: (value: T) => void; reject: (reason?: unknown) => void };

function bookmark(overrides: Partial<Bookmark> = {}): Bookmark {
	return {
		id: "bookmark-removal-a", message_id: "bookmark-removal-message-a", channel_id: "bookmark-removal-channel-a",
		channel_name: "Bookmark removal channel A", content_preview: "Bookmark removal preview A",
		created_at: "2026-09-06T00:00:00Z", ...overrides,
	};
}

function deferred<T>(): Deferred<T> {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((onResolve, onReject) => { resolve = onResolve; reject = onReject; });
	return { promise, resolve, reject };
}

function observe<T>(request: Deferred<T>) {
	return request.promise.then(() => undefined, () => undefined);
}

async function settle<T>(request: Deferred<T>, outcome: "success" | "error", value: T, error: Error) {
	const observed = observe(request);
	await act(async () => {
		if (outcome === "success") request.resolve(value);
		else request.reject(error);
		await observed;
	});
}

function createQueryClient() {
	return new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
}

function renderPanel(queryClient = createQueryClient()) {
	queryClients.push(queryClient);
	return { ...renderWithProviders(<BookmarksPanel />, { queryClient }), queryClient };
}

function removeButtons() {
	return screen.getAllByRole("button", { name: REMOVE_LABEL });
}

async function waitForRows(...bookmarks: Bookmark[]) {
	for (const item of bookmarks) await screen.findByText(item.channel_name!);
	return removeButtons();
}

async function expectDelete(number: number, bookmarkId = "bookmark-removal-a") {
	await waitFor(() => expect(removeBookmark).toHaveBeenNthCalledWith(number, bookmarkId));
}

describe("bookmark removal lifecycle", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(getBookmarks).mockReset().mockResolvedValue([bookmark()]);
		vi.mocked(removeBookmark).mockReset().mockResolvedValue(undefined);
		vi.mocked(toastApiError).mockReset();
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: true, user: userA });
	});

	afterEach(() => {
		cleanup();
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: false, user: null });
	});

	it("single-flights same-tick duplicate removal for one bookmark and disables its row", async () => {
		const request = deferred<void>();
		vi.mocked(removeBookmark).mockReturnValue(request.promise);
		renderPanel();
		const [remove] = await waitForRows(bookmark());
		act(() => { remove.click(); remove.click(); });
		await expectDelete(1);
		expect(removeBookmark).toHaveBeenCalledTimes(1);
		expect(remove).toBeDisabled();
		await settle(request, "success", undefined, new Error("unused"));
	});

	it("allows distinct bookmark removals concurrently while each pending row is disabled", async () => {
		const bookmarkA = bookmark();
		const bookmarkB = bookmark({ id: "bookmark-removal-b", message_id: "bookmark-removal-message-b", channel_name: "Bookmark removal channel B" });
		const first = deferred<void>();
		const second = deferred<void>();
		vi.mocked(getBookmarks).mockResolvedValueOnce([bookmarkA, bookmarkB]);
		vi.mocked(removeBookmark).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
		renderPanel();
		const [removeA, removeB] = await waitForRows(bookmarkA, bookmarkB);
		act(() => { removeA.click(); removeB.click(); });
		await expectDelete(1, bookmarkA.id);
		await expectDelete(2, bookmarkB.id);
		expect(removeA).toBeDisabled();
		expect(removeB).toBeDisabled();
		await act(async () => { first.resolve(undefined); second.resolve(undefined); await Promise.all([first.promise, second.promise]); });
	});

	it("keeps current pending through a real exact QueryObserver refetch and leaves other-user cache unchanged", async () => {
		const bookmarkA = bookmark();
		const post = deferred<void>();
		const refetch = deferred<Bookmark[]>();
		const currentKey = bookmarksQueryKey(userA.id);
		const otherKey = bookmarksQueryKey(userB.id);
		const otherRows = [{ ...bookmark({ id: "bookmark-removal-other", channel_name: "Other user bookmark" }), marker: { private: true } }];
		const otherSnapshot = structuredClone(otherRows);
		vi.mocked(getBookmarks)
			.mockResolvedValueOnce([bookmarkA])
			.mockReturnValueOnce(refetch.promise)
			.mockResolvedValue([bookmarkA]);
		vi.mocked(removeBookmark).mockReturnValueOnce(post.promise);
		const view = renderPanel();
		view.queryClient.setQueryData(otherKey, otherRows);
		const observer = new QueryObserver(view.queryClient, {
			queryKey: currentKey, queryFn: () => refetch.promise, retry: false, enabled: false,
		});
		const unsubscribe = observer.subscribe(() => undefined);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			const [remove] = await waitForRows(bookmarkA);
			fireEvent.click(remove);
			await expectDelete(1, bookmarkA.id);
			await settle(post, "success", undefined, new Error("unused"));
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: currentKey, exact: true }));
			expect(invalidate).toHaveBeenCalledTimes(1);
			await waitFor(() => expect(observer.getCurrentResult().isFetching).toBe(true));
			expect(remove).toBeDisabled();
			expect(view.queryClient.getQueryData(otherKey)).toEqual(otherSnapshot);
			await act(async () => { refetch.resolve([bookmarkA]); await refetch.promise; });
			await waitFor(() => expect(removeButtons()[0]).toBeEnabled());
			expect(view.queryClient.getQueryData(otherKey)).toEqual(otherSnapshot);
		} finally {
			unsubscribe();
			invalidate.mockRestore();
		}
	});

	it("reports a current delete error once, does not invalidate, and retries only after a later click", async () => {
		const failure = new Error("bookmark removal failed");
		vi.mocked(removeBookmark).mockRejectedValueOnce(failure).mockResolvedValueOnce(undefined);
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			const [remove] = await waitForRows(bookmark());
			fireEvent.click(remove);
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, ERROR_RU));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(invalidate).not.toHaveBeenCalled();
			expect(remove).toBeEnabled();
			fireEvent.click(remove);
			await expectDelete(2);
			expect(toastApiError).toHaveBeenCalledTimes(1);
		} finally { invalidate.mockRestore(); }
	});

	it.each([["success"], ["error"]] as const)("ignores a committed actor A-B-A stale %s and keeps the newer pending request locked", async (outcome) => {
		const stale = deferred<void>();
		const current = deferred<void>();
		const bookmarkA = bookmark();
		vi.mocked(removeBookmark).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click((await waitForRows(bookmarkA))[0]);
			await expectDelete(1, bookmarkA.id);
			act(() => useAuthStore.setState({ user: userB }));
			await waitForRows(bookmarkA);
			act(() => useAuthStore.setState({ user: userA }));
			const [newest] = await waitForRows(bookmarkA);
			fireEvent.click(newest);
			await expectDelete(2, bookmarkA.id);
			expect(newest).toBeDisabled();
			await settle(stale, outcome, undefined, new Error("stale actor"));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(newest).toBeDisabled();
			await settle(current, "success", undefined, new Error("unused"));
			await waitFor(() => expect(removeButtons()[0]).toBeEnabled());
		} finally { invalidate.mockRestore(); }
	});

	it.each([["success"], ["error"]] as const)("consumes an unmounted %s without an invalidation, toast, or unhandled rejection", async (outcome) => {
		const stale = deferred<void>();
		vi.mocked(removeBookmark).mockReturnValueOnce(stale.promise);
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click((await waitForRows(bookmark()))[0]);
			await expectDelete(1);
			view.unmount();
			await settle(stale, outcome, undefined, new Error("stale unmount"));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it("hides cached rows and controls without an authenticated actor and never starts GET or DELETE", async () => {
		const queryClient = createQueryClient();
		const noUserKey = bookmarksQueryKey("__no-user__");
		queryClient.setQueryData(noUserKey, [bookmark()]);
		queryClient.setQueryData(["messenger", "bookmarks"], [bookmark()]);
		useAuthStore.setState({ isAuthenticated: false, user: null });
		renderPanel(queryClient);
		await act(async () => Promise.resolve());
		expect(getBookmarks).not.toHaveBeenCalled();
		expect(removeBookmark).not.toHaveBeenCalled();
		expect(screen.queryByText(bookmark().channel_name!)).not.toBeInTheDocument();
		expect(screen.queryByRole("button", { name: REMOVE_LABEL })).not.toBeInTheDocument();
	});

	it("does not delete or create pending UI for a rendered bookmark with an empty id", async () => {
		const emptyId = bookmark({ id: "", channel_name: "Empty bookmark id" });
		vi.mocked(getBookmarks).mockResolvedValueOnce([emptyId]);
		renderPanel();
		const [remove] = await waitForRows(emptyId);
		fireEvent.click(remove);
		await act(async () => Promise.resolve());
		expect(removeBookmark).not.toHaveBeenCalled();
		expect(remove).toBeEnabled();
	});
});
