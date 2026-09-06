import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import {
	togglePin,
	type Channel,
	type ChannelCapabilities,
	type Message,
} from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { pinsQueryKey } from "@/components/messenger/pinned-messages";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";
const api = vi.hoisted(() => ({ togglePin: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));
vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, togglePin: api.togglePin };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@/components/messenger/forward-dialog", () => ({ ForwardDialog: () => null }));
const CHANNEL_A = "pin-channel-a";
const CHANNEL_B = "pin-channel-b";
const PIN_LABEL = "Закрепить сообщение";
const REPLY_LABEL = "Ответить";
const ERROR_EN = "Unable to update pinned message";
const ERROR_RU = "Не удалось изменить закрепление сообщения";
const userA = { id: "pin-user-a", email: "pin-user-a@example.test", timezone: "UTC", language: "en", system_role: "member", onboarding_completed: true };
const userB = { ...userA, id: "pin-user-b", email: "pin-user-b@example.test" };
const queryClients: QueryClient[] = [];
type Deferred<T> = { promise: Promise<T>; reject: (reason?: unknown) => void; resolve: (value: T) => void };
function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true, has_active_project_membership: false, can_read: true,
		can_post: true, can_reply: true, can_join: false, can_follow: false,
		can_unfollow: false, can_view_unread: true, can_type: true, can_edit: false,
		can_forward: false, can_react: false, can_delete_message: false,
		can_pin: true, can_list_pins: false, can_bookmark: false, can_promote: false,
		...overrides,
	};
}
function channel(id: string, currentCapabilities: ChannelCapabilities): Channel {
	return {
		id, name: id, channel_type: "private", is_readonly: false,
		capabilities: currentCapabilities, created_by: userA.id, is_starred: false,
		created_at: "2026-09-06T00:00:00Z", updated_at: "2026-09-06T00:00:00Z",
	};
}
function message(overrides: Partial<Message> = {}): Message {
	return {
		id: "pin-message-a", channel_id: CHANNEL_A, sender_id: userA.id,
		content_text: "Pin lifecycle message", reply_count: 0,
		created_at: "2026-09-06T00:00:00Z", ...overrides,
	};
}
function deferred<T>(): Deferred<T> {
	let resolve: (value: T) => void = () => undefined;
	let reject: (reason?: unknown) => void = () => undefined;
	const promise = new Promise<T>((onResolve, onReject) => {
		resolve = onResolve;
		reject = onReject;
	});
	return { promise, reject, resolve };
}
function observe<T>(request: Deferred<T>) {
	return request.promise.then(() => undefined, () => undefined);
}
async function settle<T>(
	request: Deferred<T>,
	observed: Promise<void>,
	outcome: "success" | "error",
	value: T,
	error: Error,
) {
	await act(async () => {
		if (outcome === "success") request.resolve(value);
		else request.reject(error);
		await observed;
	});
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
function renderMessageItem(item = message(), rowChannelId = CHANNEL_A, itemCapabilities = capabilities()) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return renderWithProviders(
		<MessageItem message={item} channelId={rowChannelId} capabilities={itemCapabilities} />,
		{ queryClient },
	);
}
function commitRow(
	view: ReturnType<typeof renderMessageItem>,
	item: Message,
	rowChannelId = CHANNEL_A,
	itemCapabilities = capabilities(),
) {
	act(() => view.rerender(
		<MessageItem message={item} channelId={rowChannelId} capabilities={itemCapabilities} />,
	));
}
function pinButton() {
	return screen.getByRole("button", { name: PIN_LABEL });
}
describe("message row pin command lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: true });
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAIStore.setState({ loaded: true, botLoadError: false });
		useAuthStore.setState({
			accessToken: null, refreshToken: null, isAuthenticated: true, user: userA,
		});
		useMessengerStore.getState().reset();
		installChannels([
			{ id: CHANNEL_A, currentCapabilities: capabilities() },
			{ id: CHANNEL_B, currentCapabilities: capabilities() },
		]);
	});
	afterEach(async () => {
		cleanup();
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({
			accessToken: null, refreshToken: null, isAuthenticated: false, user: null,
		});
		useMessengerStore.getState().reset();
		useAIStore.setState({ loaded: false, botLoadError: false });
		await i18n.changeLanguage("ru");
	});
	it("pins a foreign row without list permission and leaves every cache authoritative", async () => {
		const post = deferred<{ pinned: boolean }>();
		const item = message({ id: "foreign-pin-message", sender_id: userB.id });
		const ownKey = pinsQueryKey(CHANNEL_A, userA.id);
		const otherUserKey = pinsQueryKey(CHANNEL_A, userB.id);
		const otherChannelKey = pinsQueryKey(CHANNEL_B, userA.id);
		const timelineKey = messagesQueryKey(CHANNEL_A, userA.id);
		const ownPins = [{ id: "own-pin", nested: { state: "stable" } }];
		const otherUserPins = [{ id: "other-user-pin" }];
		const otherChannelPins = [{ id: "other-channel-pin" }];
		const timeline = { pages: [[item]], pageParams: [undefined] };
		const ownSnapshot = structuredClone(ownPins);
		const otherUserSnapshot = structuredClone(otherUserPins);
		const otherChannelSnapshot = structuredClone(otherChannelPins);
		const timelineSnapshot = structuredClone(timeline);
		const rowCapabilities = capabilities({ can_list_pins: false });
		vi.mocked(togglePin).mockReturnValueOnce(post.promise);
		installChannels([
			{ id: CHANNEL_A, currentCapabilities: rowCapabilities },
			{ id: CHANNEL_B, currentCapabilities: rowCapabilities },
		]);
		const { queryClient } = renderMessageItem(item, CHANNEL_A, rowCapabilities);
		queryClient.setQueryData(ownKey, ownPins);
		queryClient.setQueryData(otherUserKey, otherUserPins);
		queryClient.setQueryData(otherChannelKey, otherChannelPins);
		queryClient.setQueryData(timelineKey, timeline);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		try {
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenCalledWith(CHANNEL_A, item.id));
			expect(togglePin).toHaveBeenCalledTimes(1);
			expect(invalidate).not.toHaveBeenCalled();
			expect(queryClient.getQueryData(ownKey)).toBe(ownPins);
			expect(queryClient.getQueryData(ownKey)).toEqual(ownSnapshot);
			expect(queryClient.getQueryData(otherUserKey)).toEqual(otherUserSnapshot);
			expect(queryClient.getQueryData(otherChannelKey)).toEqual(otherChannelSnapshot);
			expect(queryClient.getQueryData(timelineKey)).toEqual(timelineSnapshot);
			await act(async () => {
				post.resolve({ pinned: true });
				await post.promise;
			});
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({
				queryKey: ownKey, exact: true,
			}));
			expect(invalidate).toHaveBeenCalledTimes(1);
			expect(queryClient.getQueryData(ownKey)).toEqual(ownSnapshot);
			expect(queryClient.getQueryData(otherUserKey)).toEqual(otherUserSnapshot);
			expect(queryClient.getQueryData(otherChannelKey)).toEqual(otherChannelSnapshot);
			expect(queryClient.getQueryData(timelineKey)).toEqual(timelineSnapshot);
			expect(toastApiError).not.toHaveBeenCalled();
		} finally {
			invalidate.mockRestore();
		}
	});
	it.each([
		{ language: "en", fallback: ERROR_EN },
		{ language: "ru", fallback: ERROR_RU },
	] as const)("reports one current $language rejection and retries only after an explicit click", async ({ language, fallback }) => {
		const failure = new Error("pin " + language + " failure");
		const item = message({ id: "failed-pin-" + language });
		const { queryClient } = renderMessageItem(item);
		const key = pinsQueryKey(CHANNEL_A, userA.id);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			await i18n.changeLanguage(language);
			await i18n.loadNamespaces("messenger");
			vi.mocked(togglePin)
				.mockRejectedValueOnce(failure)
				.mockResolvedValueOnce({ pinned: true });
			fireEvent.click(pinButton());
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, fallback));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(togglePin).toHaveBeenCalledWith(CHANNEL_A, item.id);
			expect(togglePin).toHaveBeenCalledTimes(1);
			expect(invalidate).not.toHaveBeenCalled();
			expect(pinButton()).toBeEnabled();
			expect(unhandled).not.toHaveBeenCalled();
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenNthCalledWith(2, CHANNEL_A, item.id));
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: key, exact: true }));
			expect(togglePin).toHaveBeenCalledTimes(2);
			expect(invalidate).toHaveBeenCalledTimes(1);
			expect(toastApiError).toHaveBeenCalledTimes(1);
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});
	it("deduplicates same-tick clicks while an independent row and Reply remain usable", async () => {
		const first = deferred<{ pinned: boolean }>();
		const second = deferred<{ pinned: boolean }>();
		const left = message({ id: "first-pending-pin" });
		const right = message({ id: "second-pending-pin" });
		vi.mocked(togglePin).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		renderWithProviders(
			<>
				<MessageItem message={left} channelId={CHANNEL_A} capabilities={capabilities()} />
				<MessageItem message={right} channelId={CHANNEL_A} capabilities={capabilities()} />
			</>,
			{ queryClient },
		);
		const [leftPin, rightPin] = screen.getAllByRole("button", { name: PIN_LABEL });
		act(() => {
			leftPin.click();
			leftPin.click();
			rightPin.click();
		});
		await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(2));
		expect(togglePin).toHaveBeenNthCalledWith(1, CHANNEL_A, left.id);
		expect(togglePin).toHaveBeenNthCalledWith(2, CHANNEL_A, right.id);
		expect(screen.getAllByRole("button", { name: PIN_LABEL })[0]).toBeDisabled();
		expect(screen.getAllByRole("button", { name: PIN_LABEL })[1]).toBeDisabled();
		expect(screen.getAllByRole("button", { name: REPLY_LABEL })[0]).toBeEnabled();
		await act(async () => {
			first.resolve({ pinned: true });
			second.resolve({ pinned: true });
			await Promise.all([first.promise, second.promise]);
		});
	});
	it.each([
		{ name: "the authenticated actor is absent", item: message(), prop: capabilities(), live: capabilities(), actor: null, revoked: false },
		{ name: "the message id is empty", item: message({ id: "" }), prop: capabilities(), live: capabilities(), actor: userA, revoked: false },
		{ name: "the message belongs to another channel", item: message({ channel_id: CHANNEL_B }), prop: capabilities(), live: capabilities(), actor: userA, revoked: false },
		{ name: "the rendered capability denies pinning", item: message(), prop: capabilities({ can_pin: false }), live: capabilities(), actor: userA, revoked: false },
		{ name: "the live channel capability is absent", item: message(), prop: capabilities(), live: null, actor: userA, revoked: false },
		{ name: "the live channel capability denies pinning", item: message(), prop: capabilities(), live: capabilities({ can_pin: false }), actor: userA, revoked: false },
		{ name: "the channel is revoked", item: message(), prop: capabilities(), live: capabilities(), actor: userA, revoked: true },
	] as const)("does not submit a row pin when $name", async ({ item, prop, live, actor, revoked }) => {
		useAuthStore.setState({ isAuthenticated: actor !== null, user: actor });
		installChannels(live ? [{ id: CHANNEL_A, currentCapabilities: live }] : []);
		renderMessageItem(item, CHANNEL_A, prop);
		if (revoked) {
			act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
			expect(screen.queryByRole("button", { name: PIN_LABEL })).toBeNull();
		} else {
			const control = screen.queryByRole("button", { name: PIN_LABEL });
			if (control) fireEvent.click(control);
		}
		await act(async () => Promise.resolve());
		expect(togglePin).not.toHaveBeenCalled();
	});
	it.each([
		["actor", "success"], ["actor", "error"],
		["channel", "success"], ["channel", "error"],
		["message", "success"], ["message", "error"],
	] as const)("does not let an old %s A-B-A %s settlement affect a newer pin", async (boundary, outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		const current = deferred<{ pinned: boolean }>();
		const failure = new Error("stale " + boundary + " pin");
		const initial = message({ id: "aba-pin-message-a" });
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		const view = renderMessageItem(initial);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenNthCalledWith(1, CHANNEL_A, initial.id));
			const staleObserved = observe(stale);
			if (boundary === "actor") {
				act(() => useAuthStore.setState({ user: userB }));
				act(() => useAuthStore.setState({ user: userA }));
			} else if (boundary === "channel") {
				commitRow(view, message({ id: initial.id, channel_id: CHANNEL_B }), CHANNEL_B);
				commitRow(view, initial);
			} else {
				commitRow(view, message({ id: "aba-pin-message-b" }));
				commitRow(view, initial);
			}
			expect(pinButton()).toBeEnabled();
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenNthCalledWith(2, CHANNEL_A, initial.id));
			expect(pinButton()).toBeDisabled();
			await settle(stale, staleObserved, outcome, { pinned: true }, failure);
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(pinButton()).toBeDisabled();
			expect(unhandled).not.toHaveBeenCalled();
			await act(async () => {
				current.resolve({ pinned: true });
				await current.promise;
			});
			await waitFor(() => expect(pinButton()).toBeEnabled());
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
	] as const)("consumes a late settlement after %s (%s) without current pin effects", async (boundary, outcome) => {
		const request = deferred<{ pinned: boolean }>();
		const failure = new Error("late " + boundary + " pin");
		const item = message({ id: "late-pin-" + boundary + "-" + outcome });
		vi.mocked(togglePin).mockReturnValueOnce(request.promise);
		const view = renderMessageItem(item);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
			const requestObserved = observe(request);
			if (boundary === "prop loss and restore") {
				commitRow(view, item, CHANNEL_A, capabilities({ can_pin: false }));
				commitRow(view, item);
				expect(pinButton()).toBeEnabled();
			} else if (boundary === "live loss and restore") {
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities({ can_pin: false }) }]));
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]));
				expect(pinButton()).toBeEnabled();
			} else if (boundary === "revoke") {
				act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
				expect(screen.queryByRole("button", { name: PIN_LABEL })).toBeNull();
			} else {
				view.unmount();
			}
			await settle(request, requestObserved, outcome, { pinned: true }, failure);
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});
	it("keeps pending through a real observer refetch and ignores obsolete invalidation settlement", async () => {
		const stalePost = deferred<{ pinned: boolean }>();
		const currentPost = deferred<{ pinned: boolean }>();
		const refetch = deferred<unknown>();
		const item = message({ id: "observer-pin-message" });
		const key = pinsQueryKey(CHANNEL_A, userA.id);
		vi.mocked(togglePin).mockReturnValueOnce(stalePost.promise).mockReturnValueOnce(currentPost.promise);
		const view = renderMessageItem(item);
		view.queryClient.setQueryData(key, [{ id: "cached-pin" }]);
		const observer = new QueryObserver(view.queryClient, {
			queryKey: key, queryFn: () => refetch.promise, retry: false,
			staleTime: Number.POSITIVE_INFINITY,
		});
		const unsubscribe = observer.subscribe(() => undefined);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
			await act(async () => {
				stalePost.resolve({ pinned: true });
				await stalePost.promise;
			});
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: key, exact: true }));
			await waitFor(() => expect(observer.getCurrentResult().isFetching).toBe(true));
			expect(pinButton()).toBeDisabled();
			act(() => useAuthStore.setState({ user: userB }));
			act(() => useAuthStore.setState({ user: userA }));
			expect(pinButton()).toBeEnabled();
			fireEvent.click(pinButton());
			await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(2));
			expect(pinButton()).toBeDisabled();
			await act(async () => {
				refetch.resolve([]);
				await refetch.promise;
			});
			expect(pinButton()).toBeDisabled();
			await act(async () => {
				currentPost.resolve({ pinned: true });
				await currentPost.promise;
			});
			await waitFor(() => expect(pinButton()).toBeEnabled());
		} finally {
			unsubscribe();
			invalidate.mockRestore();
		}
	});
});
