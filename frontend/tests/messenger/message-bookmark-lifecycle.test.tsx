import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import {
	addBookmark,
	type Channel,
	type ChannelCapabilities,
	type Message,
} from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { bookmarksQueryKey } from "@/components/messenger/bookmarks-panel";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ addBookmark: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, addBookmark: api.addBookmark };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@/components/messenger/forward-dialog", () => ({ ForwardDialog: () => null }));

const CHANNEL_A = "bookmark-channel-a";
const CHANNEL_B = "bookmark-channel-b";
const BOOKMARK_LABEL = "Сохранить в закладки";
const REPLY_LABEL = "Ответить";
const RU_FAILURE = "Не удалось сохранить закладку";
const EN_FAILURE = "Unable to save bookmark";
const userA = {
	id: "bookmark-user-a",
	email: "bookmark-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = { ...userA, id: "bookmark-user-b", email: "bookmark-user-b@example.test" };
const queryClients: QueryClient[] = [];

function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
		can_read: true,
		can_post: true,
		can_reply: true,
		can_join: false,
		can_follow: false,
		can_unfollow: false,
		can_view_unread: true,
		can_type: true,
		can_edit: false,
		can_forward: false,
		can_react: false,
		can_delete_message: false,
		can_pin: false,
		can_list_pins: false,
		can_bookmark: true,
		can_promote: false,
		...overrides,
	};
}
function channel(id: string, currentCapabilities: ChannelCapabilities): Channel {
	return {
		id,
		name: id,
		channel_type: "private",
		is_readonly: false,
		capabilities: currentCapabilities,
		created_by: userA.id,
		is_starred: false,
		created_at: "2026-09-06T00:00:00Z",
		updated_at: "2026-09-06T00:00:00Z",
	};
}
function message(overrides: Partial<Message> = {}): Message {
	return {
		id: "bookmark-message-a",
		channel_id: CHANNEL_A,
		sender_id: userA.id,
		content_text: "Bookmark lifecycle message",
		reply_count: 0,
		created_at: "2026-09-06T00:00:00Z",
		...overrides,
	};
}
function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((onResolve, onReject) => {
		resolve = onResolve;
		reject = onReject;
	});
	return { promise, reject, resolve };
}
function createQueryClient() {
	return new QueryClient({
		defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
	});
}
function installChannels(entries: Array<{ id: string; currentCapabilities: ChannelCapabilities }>) {
	useMessengerStore.getState().setChannels(
		entries.map(({ id, currentCapabilities }) => channel(id, currentCapabilities)),
	);
}
function renderMessageItem(
	item = message(),
	rowChannelId = CHANNEL_A,
	itemCapabilities = capabilities(),
) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return renderWithProviders(
		<MessageItem message={item} channelId={rowChannelId} capabilities={itemCapabilities} />,
		{ queryClient },
	);
}
function bookmarkButton() {
	return screen.getByRole("button", { name: BOOKMARK_LABEL });
}

describe("message bookmark command lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(addBookmark).mockReset().mockResolvedValue(undefined);
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAIStore.setState({ loaded: true, botLoadError: false });
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: true, user: userA });
		useMessengerStore.getState().reset();
		installChannels([
			{ id: CHANNEL_A, currentCapabilities: capabilities() },
			{ id: CHANNEL_B, currentCapabilities: capabilities() },
		]);
	});
	afterEach(async () => {
		cleanup();
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: false, user: null });
		useMessengerStore.getState().reset();
		useAIStore.setState({ loaded: false, botLoadError: false });
		await i18n.changeLanguage("ru");
	});

	it("posts captured message then channel and exact-invalidates only the current user bookmarks", async () => {
		const currentKey = bookmarksQueryKey(userA.id);
		const otherKey = bookmarksQueryKey(userB.id);
		const timelineKey = messagesQueryKey(CHANNEL_A, userA.id);
		const post = deferred<void>();
		const foreign = message({ id: "save-current-a", sender_id: userB.id });
		const currentSeed = [{ id: "current-user-bookmark", nested: { value: "stable" } }];
		const currentSnapshot = structuredClone(currentSeed);
		vi.mocked(addBookmark).mockReturnValueOnce(post.promise);
		const { queryClient } = renderMessageItem(foreign);
		queryClient.setQueryData(currentKey, currentSeed);
		queryClient.setQueryData(otherKey, ["other-user-bookmark"]);
		queryClient.setQueryData(timelineKey, { pages: [[message()]], pageParams: [undefined] });
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		try {
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenCalledWith("save-current-a", CHANNEL_A));
			expect(queryClient.getQueryData(currentKey)).toEqual(currentSnapshot);
			expect(invalidate).not.toHaveBeenCalled();
			await act(async () => { post.resolve(undefined); await post.promise; });
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: currentKey, exact: true }));
			expect(invalidate).toHaveBeenCalledTimes(1);
			expect(queryClient.getQueryData(currentKey)).toEqual(currentSnapshot);
			expect(queryClient.getQueryData(otherKey)).toEqual(["other-user-bookmark"]);
			expect(queryClient.getQueryData(timelineKey)).toEqual({ pages: [[message()]], pageParams: [undefined] });
		} finally {
			invalidate.mockRestore();
		}
	});

	it.each([
		{ language: "en", fallback: EN_FAILURE },
		{ language: "ru", fallback: RU_FAILURE },
	])("reports the current $language failure once and permits only explicit retry", async ({ language, fallback }) => {
		const failure = new Error(`bookmark ${language} failure`);
		const { queryClient } = renderMessageItem(message({ id: `failed-${language}` }));
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			await i18n.changeLanguage(language);
			await i18n.loadNamespaces("messenger");
			vi.mocked(addBookmark).mockRejectedValueOnce(failure).mockResolvedValueOnce(undefined);
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, fallback));
			expect(invalidate).not.toHaveBeenCalled();
			expect(bookmarkButton()).toBeEnabled();
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenCalledTimes(2));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it("deduplicates two same-tick native clicks while another row can submit and Reply stays enabled", async () => {
		const first = deferred<void>();
		const second = deferred<void>();
		vi.mocked(addBookmark).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
		const firstMessage = message({ id: "double-click-first" });
		const secondMessage = message({ id: "double-click-second" });
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		renderWithProviders(
			<><MessageItem message={firstMessage} channelId={CHANNEL_A} capabilities={capabilities()} /><MessageItem message={secondMessage} channelId={CHANNEL_A} capabilities={capabilities()} /></>,
			{ queryClient },
		);
		const [firstButton, secondButton] = screen.getAllByRole("button", { name: BOOKMARK_LABEL });
		act(() => { firstButton.click(); firstButton.click(); secondButton.click(); });
		await waitFor(() => expect(addBookmark).toHaveBeenCalledTimes(2));
		expect(addBookmark).toHaveBeenNthCalledWith(1, firstMessage.id, CHANNEL_A);
		expect(addBookmark).toHaveBeenNthCalledWith(2, secondMessage.id, CHANNEL_A);
		expect(screen.getAllByRole("button", { name: BOOKMARK_LABEL })).toEqual([
			expect.objectContaining({ disabled: true }),
			expect.objectContaining({ disabled: true }),
		]);
		expect(screen.getAllByRole("button", { name: REPLY_LABEL })[0]).toBeEnabled();
		await act(async () => { first.resolve(undefined); second.resolve(undefined); await Promise.all([first.promise, second.promise]); });
	});

	it.each([
		{ name: "the actor is absent", item: message(), prop: capabilities(), live: capabilities(), actor: null, revoke: false },
		{ name: "the message id is empty", item: message({ id: "" }), prop: capabilities(), live: capabilities(), actor: userA, revoke: false },
		{ name: "the message belongs to another channel", item: message({ channel_id: CHANNEL_B }), prop: capabilities(), live: capabilities(), actor: userA, revoke: false },
		{ name: "the rendered capability denies bookmarking", item: message(), prop: capabilities({ can_bookmark: false }), live: capabilities(), actor: userA, revoke: false },
		{ name: "the live channel is absent", item: message(), prop: capabilities(), live: null, actor: userA, revoke: false },
		{ name: "the live capability denies bookmarking", item: message(), prop: capabilities(), live: capabilities({ can_bookmark: false }), actor: userA, revoke: false },
		{ name: "the channel is revoked", item: message(), prop: capabilities(), live: capabilities(), actor: userA, revoke: true },
	])("fails closed before POST when $name", async ({ item, prop, live, actor, revoke }) => {
		useAuthStore.setState({ isAuthenticated: Boolean(actor), user: actor });
		installChannels(live ? [{ id: CHANNEL_A, currentCapabilities: live }] : []);
		renderMessageItem(item, CHANNEL_A, prop);
		if (revoke) act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
		const button = screen.queryByRole("button", { name: BOOKMARK_LABEL });
		if (button) fireEvent.click(button);
		await act(async () => Promise.resolve());
		expect(addBookmark).not.toHaveBeenCalled();
	});

	it.each([
		["user", "success"], ["user", "error"], ["channel", "success"],
		["channel", "error"], ["message", "success"], ["message", "error"],
	] as const)("keeps a later command isolated from an old %s A-B-A %s completion", async (boundary, outcome) => {
		const stale = deferred<void>();
		const current = deferred<void>();
		const failure = new Error(`stale ${boundary} bookmark`);
		const initial = message({ id: "aba-bookmark-a" });
		vi.mocked(addBookmark).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		const view = renderMessageItem(initial);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenNthCalledWith(1, initial.id, CHANNEL_A));
			const staleSettled = stale.promise.then(() => undefined, () => undefined);
			if (boundary === "user") {
				act(() => useAuthStore.setState({ user: userB }));
				act(() => useAuthStore.setState({ user: userA }));
			} else if (boundary === "channel") {
				view.rerender(<MessageItem message={message({ id: "aba-bookmark-b", channel_id: CHANNEL_B })} channelId={CHANNEL_B} capabilities={capabilities()} />);
				view.rerender(<MessageItem message={initial} channelId={CHANNEL_A} capabilities={capabilities()} />);
			} else {
				view.rerender(<MessageItem message={message({ id: "aba-bookmark-b" })} channelId={CHANNEL_A} capabilities={capabilities()} />);
				view.rerender(<MessageItem message={initial} channelId={CHANNEL_A} capabilities={capabilities()} />);
			}
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenNthCalledWith(2, initial.id, CHANNEL_A));
			expect(bookmarkButton()).toBeDisabled();
			await act(async () => { if (outcome === "success") stale.resolve(undefined); else stale.reject(failure); await staleSettled; });
			expect(bookmarkButton()).toBeDisabled();
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
			await act(async () => { current.resolve(undefined); await current.promise; });
			await waitFor(() => expect(bookmarkButton()).toBeEnabled());
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it.each([
		["prop loss and restore", "success"], ["prop loss and restore", "error"],
		["live loss and restore", "success"], ["live loss and restore", "error"],
		["revoke", "success"], ["revoke", "error"],
		["unmount", "success"], ["unmount", "error"],
	] as const)("consumes a late %s after %s without current side effects", async (boundary, outcome) => {
		const request = deferred<void>();
		const failure = new Error(`late bookmark ${boundary}`);
		vi.mocked(addBookmark).mockReturnValueOnce(request.promise);
		const item = message({ id: `late-${boundary}-${outcome}` });
		const view = renderMessageItem(item);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenCalledTimes(1));
			const settled = request.promise.then(() => undefined, () => undefined);
			if (boundary === "prop loss and restore") {
				view.rerender(<MessageItem message={item} channelId={CHANNEL_A} capabilities={capabilities({ can_bookmark: false })} />);
				view.rerender(<MessageItem message={item} channelId={CHANNEL_A} capabilities={capabilities()} />);
			} else if (boundary === "live loss and restore") {
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities({ can_bookmark: false }) }]));
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]));
			} else if (boundary === "revoke") {
				act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
			} else {
				view.unmount();
			}
			await act(async () => { if (outcome === "success") request.resolve(undefined); else request.reject(failure); await settled; });
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it("holds current pending through real exact invalidation and an obsolete completion cannot unlock a newer scope", async () => {
		const stalePost = deferred<void>();
		const currentPost = deferred<void>();
		const invalidation = deferred<unknown>();
		const key = bookmarksQueryKey(userA.id);
		vi.mocked(addBookmark).mockReturnValueOnce(stalePost.promise).mockReturnValueOnce(currentPost.promise);
		const view = renderMessageItem(message({ id: "invalidation-aba" }));
		view.queryClient.setQueryData(key, ["cached"]);
		const observer = new QueryObserver(view.queryClient, {
			queryKey: key,
			queryFn: () => invalidation.promise,
			staleTime: Number.POSITIVE_INFINITY,
			retry: false,
		});
		const unsubscribe = observer.subscribe(() => undefined);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenCalledTimes(1));
			await act(async () => { stalePost.resolve(undefined); await stalePost.promise; });
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: key, exact: true }));
			expect(bookmarkButton()).toBeDisabled();
			act(() => useAuthStore.setState({ user: userB }));
			act(() => useAuthStore.setState({ user: userA }));
			fireEvent.click(bookmarkButton());
			await waitFor(() => expect(addBookmark).toHaveBeenCalledTimes(2));
			expect(bookmarkButton()).toBeDisabled();
			await act(async () => { invalidation.resolve([]); await invalidation.promise; });
			expect(bookmarkButton()).toBeDisabled();
			await act(async () => { currentPost.resolve(undefined); await currentPost.promise; });
			await waitFor(() => expect(bookmarkButton()).toBeEnabled());
		} finally {
			unsubscribe();
			invalidate.mockRestore();
		}
	});
});
