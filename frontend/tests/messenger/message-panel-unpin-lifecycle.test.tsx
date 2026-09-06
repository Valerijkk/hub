import { QueryClient, QueryObserver } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import i18n from "@/i18n";
import {
	getChannelPins,
	togglePin,
	type Channel,
	type ChannelCapabilities,
	type Message,
	type PinnedMessage,
} from "@/api/messenger";
import { PinnedMessages, pinsQueryKey } from "@/components/messenger/pinned-messages";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ getChannelPins: vi.fn(), togglePin: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, getChannelPins: api.getChannelPins, togglePin: api.togglePin };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));

const CHANNEL_A = "panel-unpin-channel-a";
const CHANNEL_B = "panel-unpin-channel-b";
const PANEL_LABEL = "Закреплённые сообщения";
const UNPIN_LABEL = "Открепить";
const ERROR_EN = "Unable to update pinned message";
const ERROR_RU = "Не удалось изменить закрепление сообщения";
const userA = { id: "panel-unpin-user-a", email: "panel-a@example.test", timezone: "UTC", language: "en", system_role: "member", onboarding_completed: true };
const userB = { ...userA, id: "panel-unpin-user-b", email: "panel-b@example.test" };
const queryClients: QueryClient[] = [];

type Deferred<T> = { promise: Promise<T>; resolve: (value: T) => void; reject: (reason?: unknown) => void };

function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true, has_active_project_membership: false, can_read: true,
		can_post: true, can_reply: true, can_join: false, can_follow: false,
		can_unfollow: false, can_view_unread: true, can_type: true, can_edit: false,
		can_forward: false, can_react: false, can_delete_message: false,
		can_pin: true, can_list_pins: true, can_bookmark: false, can_promote: false,
		...overrides,
	};
}

function channel(id: string, currentCapabilities = capabilities()): Channel {
	return {
		id, name: id, channel_type: "private", is_readonly: false,
		capabilities: currentCapabilities, created_by: userA.id, is_starred: false,
		created_at: "2026-09-06T00:00:00Z", updated_at: "2026-09-06T00:00:00Z",
	};
}

function pin(overrides: Partial<PinnedMessage> = {}): PinnedMessage {
	return {
		id: "panel-unpin-pin-a", message_id: "panel-unpin-message-a", channel_id: CHANNEL_A,
		pinned_by: userA.id, pinned_at: "2026-09-06T00:00:00Z", ...overrides,
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

function installChannels(entries: Array<{ id: string; currentCapabilities: ChannelCapabilities }>) {
	useMessengerStore.getState().setChannels(entries.map(({ id, currentCapabilities }) => channel(id, currentCapabilities)));
}

function panel(channelId = CHANNEL_A, currentCapabilities = capabilities()) {
	return <PinnedMessages channelId={channelId} capabilities={currentCapabilities} />;
}

function renderPanel(channelId = CHANNEL_A, currentCapabilities = capabilities()) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return { ...renderWithProviders(panel(channelId, currentCapabilities), { queryClient }), queryClient };
}

function commitPanel(view: ReturnType<typeof renderPanel>, channelId = CHANNEL_A, currentCapabilities = capabilities()) {
	act(() => view.rerender(panel(channelId, currentCapabilities)));
}

function openPanel() {
	fireEvent.click(screen.getByRole("button", { name: PANEL_LABEL }));
}

async function openUnpin() {
	openPanel();
	return screen.findByRole("button", { name: UNPIN_LABEL });
}

function unpinButton() {
	return screen.getByRole("button", { name: UNPIN_LABEL });
}

async function expectPost(number: number, channelId = CHANNEL_A, messageId = "panel-unpin-message-a") {
	await waitFor(() => expect(togglePin).toHaveBeenNthCalledWith(number, channelId, messageId));
}

describe("panel unpin command lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(getChannelPins).mockReset().mockImplementation((channelId) =>
			Promise.resolve([pin({ channel_id: channelId })]),
		);
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: false });
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: true, user: userA });
		useMessengerStore.getState().reset();
		installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }, { id: CHANNEL_B, currentCapabilities: capabilities() }]);
	});

	afterEach(async () => {
		cleanup();
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({ accessToken: null, refreshToken: null, isAuthenticated: false, user: null });
		useMessengerStore.getState().reset();
		await i18n.changeLanguage("ru");
	});

	it("unpins with the exact identity, preserves every cache before refetch, and keeps the panel open", async () => {
		const post = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(post.promise);
		const view = renderPanel();
		const ownKey = pinsQueryKey(CHANNEL_A, userA.id);
		const otherUserKey = pinsQueryKey(CHANNEL_A, userB.id);
		const otherChannelKey = pinsQueryKey(CHANNEL_B, userA.id);
		const timelineKey = messagesQueryKey(CHANNEL_A, userA.id);
		const own = [{ ...pin(), marker: { state: "stable" } }];
		const otherUser = [{ ...pin({ id: "other-user" }), marker: { state: "private" } }];
		const otherChannel = [{ ...pin({ id: "other-channel", channel_id: CHANNEL_B }), marker: { state: "other" } }];
		const timeline = { pages: [[{ id: "panel-unpin-message-a", content_text: "timeline" } as Message]], pageParams: [undefined] };
		view.queryClient.setQueryData(ownKey, own);
		view.queryClient.setQueryData(otherUserKey, otherUser);
		view.queryClient.setQueryData(otherChannelKey, otherChannel);
		view.queryClient.setQueryData(timelineKey, timeline);
		const ownSnapshot = structuredClone(own);
		const otherUserSnapshot = structuredClone(otherUser);
		const otherChannelSnapshot = structuredClone(otherChannel);
		const timelineSnapshot = structuredClone(timeline);
		vi.mocked(getChannelPins).mockResolvedValueOnce(own);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			expect(view.queryClient.getQueryData(ownKey)).toEqual(ownSnapshot);
			expect(view.queryClient.getQueryData(otherUserKey)).toEqual(otherUserSnapshot);
			expect(view.queryClient.getQueryData(otherChannelKey)).toEqual(otherChannelSnapshot);
			expect(view.queryClient.getQueryData(timelineKey)).toEqual(timelineSnapshot);
			await settle(post, "success", { pinned: false }, new Error("unused"));
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: ownKey, exact: true }));
			expect(invalidate).toHaveBeenCalledTimes(1);
			await waitFor(() => expect(view.queryClient.getQueryData(ownKey)).toEqual([pin()]));
			expect(view.queryClient.getQueryData(otherUserKey)).toEqual(otherUserSnapshot);
			expect(view.queryClient.getQueryData(otherChannelKey)).toEqual(otherChannelSnapshot);
			expect(view.queryClient.getQueryData(timelineKey)).toEqual(timelineSnapshot);
			expect(screen.getByRole("dialog")).toBeInTheDocument();
		} finally { invalidate.mockRestore(); }
	});

	it.each([
		["en", ERROR_EN],
		["ru", ERROR_RU],
	] as const)("reports one current %s failure and permits only an explicit retry", async (language, fallback) => {
		const failure = new Error(`panel unpin ${language} failure`);
		const view = renderPanel();
		const key = pinsQueryKey(CHANNEL_A, userA.id);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			await i18n.changeLanguage(language);
			vi.mocked(togglePin).mockRejectedValueOnce(failure).mockResolvedValueOnce({ pinned: false });
			fireEvent.click(await openUnpin());
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, fallback));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(invalidate).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
			expect(unpinButton()).toBeEnabled();
			fireEvent.click(unpinButton());
			await expectPost(2);
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: key, exact: true }));
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it("claims one panel request for same-tick clicks on different pins and disables every Unpin button", async () => {
		const post = deferred<{ pinned: boolean }>();
		vi.mocked(getChannelPins).mockResolvedValueOnce([pin(), pin({ id: "panel-unpin-pin-b", message_id: "panel-unpin-message-b" })]);
		vi.mocked(togglePin).mockReturnValueOnce(post.promise);
		renderPanel();
		openPanel();
		const buttons = await screen.findAllByRole("button", { name: UNPIN_LABEL });
		act(() => { buttons[0].click(); buttons[1].click(); });
		await expectPost(1);
		expect(togglePin).toHaveBeenCalledTimes(1);
		expect(buttons[0]).toBeDisabled();
		expect(buttons[1]).toBeDisabled();
		await settle(post, "success", { pinned: false }, new Error("unused"));
	});

	it.each([
		["the authenticated actor is absent", CHANNEL_A, pin(), true],
		["the panel channel id is empty", "", pin({ channel_id: "" }), false],
		["the pin id is empty", CHANNEL_A, pin({ id: "" }), false],
		["the pinned message id is empty", CHANNEL_A, pin({ message_id: "" }), false],
		["the pin belongs to another channel", CHANNEL_A, pin({ channel_id: CHANNEL_B }), false],
	] as const)("does not submit when %s", async (_name, channelId, currentPin, loseActor) => {
		installChannels([{ id: channelId, currentCapabilities: capabilities() }]);
		vi.mocked(getChannelPins).mockResolvedValueOnce([currentPin]);
		renderPanel(channelId);
		const control = await openUnpin();
		if (loseActor) {
			act(() => { useAuthStore.setState({ user: null }); control.click(); });
		} else fireEvent.click(control);
		await act(async () => Promise.resolve());
		expect(togglePin).not.toHaveBeenCalled();
	});

	it.each([
		["the live channel is absent", null, false],
		["live list permission is false", capabilities({ can_list_pins: false }), false],
		["live pin permission is false", capabilities({ can_pin: false }), false],
		["the channel is revoked", capabilities(), true],
	] as const)("does not submit when %s", async (_name, liveCapabilities, revoke) => {
		if (liveCapabilities) installChannels([{ id: CHANNEL_A, currentCapabilities: liveCapabilities }]);
		else useMessengerStore.getState().reset();
		renderPanel();
		const control = await openUnpin();
		if (revoke) act(() => { useMessengerStore.getState().revokeChannel(CHANNEL_A); control.click(); });
		else fireEvent.click(control);
		await act(async () => Promise.resolve());
		expect(togglePin).not.toHaveBeenCalled();
	});

	it.each([
		["actor", "success"], ["actor", "error"],
		["channel", "success"], ["channel", "error"],
		["open session", "success"], ["open session", "error"],
	] as const)("does not let a committed %s A-B-A %s settlement affect a newer panel request", async (boundary, outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		const current = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			if (boundary === "actor") {
				act(() => useAuthStore.setState({ user: userB }));
				act(() => useAuthStore.setState({ user: userA }));
			} else if (boundary === "channel") {
				commitPanel(view, CHANNEL_B);
				commitPanel(view, CHANNEL_A);
			} else {
				act(() => fireEvent.keyDown(document, { key: "Escape" }));
				await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
				openPanel();
			}
			const newest = await screen.findByRole("button", { name: UNPIN_LABEL });
			fireEvent.click(newest);
			await expectPost(2);
			expect(newest).toBeDisabled();
			await settle(stale, outcome, { pinned: false }, new Error(`stale ${boundary}`));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(newest).toBeDisabled();
			await settle(current, "success", { pinned: false }, new Error("unused"));
			await waitFor(() => expect(unpinButton()).toBeEnabled());
		} finally { invalidate.mockRestore(); }
	});

	it.each([
		["prop list permission", "prop-list", "success"], ["prop list permission", "prop-list", "error"],
		["prop pin permission", "prop-pin", "success"], ["prop pin permission", "prop-pin", "error"],
		["live list permission", "live-list", "success"], ["live list permission", "live-list", "error"],
		["live pin permission", "live-pin", "success"], ["live pin permission", "live-pin", "error"],
	] as const)("consumes a late %s loss-and-restore %s settlement before explicit retry", async (_name, boundary, outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockResolvedValueOnce({ pinned: false });
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			if (boundary === "prop-list") {
				commitPanel(view, CHANNEL_A, capabilities({ can_list_pins: false }));
				await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
				commitPanel(view);
				openPanel();
			} else if (boundary === "prop-pin") {
				commitPanel(view, CHANNEL_A, capabilities({ can_pin: false }));
				commitPanel(view);
			} else {
				act(() => {
					installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities({ [boundary === "live-list" ? "can_list_pins" : "can_pin"]: false }) }]);
					installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]);
				});
			}
			await screen.findByRole("button", { name: UNPIN_LABEL });
			await settle(stale, outcome, { pinned: false }, new Error(`stale ${boundary}`));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			fireEvent.click(unpinButton());
			await expectPost(2);
			await waitFor(() => expect(invalidate).toHaveBeenCalledTimes(1));
		} finally { invalidate.mockRestore(); }
	});

	it.each([["success"], ["error"]] as const)("consumes a late revoke-and-restore %s settlement before explicit retry", async (outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockResolvedValueOnce({ pinned: false });
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
			await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
			act(() => {
				useMessengerStore.getState().restoreChannelAccess(CHANNEL_A);
				installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]);
			});
			await openUnpin();
			await settle(stale, outcome, { pinned: false }, new Error("stale revoke"));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			fireEvent.click(unpinButton());
			await expectPost(2);
		} finally { invalidate.mockRestore(); }
	});

	it.each([["success"], ["error"]] as const)("consumes an unmounted %s settlement without a toast or unhandled rejection", async (outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise);
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			view.unmount();
			await settle(stale, outcome, { pinned: false }, new Error("stale unmount"));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
			invalidate.mockRestore();
		}
	});

	it("keeps pending through a real QueryObserver refetch and never lets an old invalidation unlock the new scope", async () => {
		const stale = deferred<{ pinned: boolean }>();
		const current = deferred<{ pinned: boolean }>();
		const refetch = deferred<PinnedMessage[]>();
		vi.mocked(getChannelPins)
			.mockResolvedValueOnce([pin()])
			.mockReturnValueOnce(refetch.promise)
			.mockImplementation((channelId) => Promise.resolve([pin({ channel_id: channelId })]));
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
		const view = renderPanel();
		const key = pinsQueryKey(CHANNEL_A, userA.id);
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		const observer = new QueryObserver(view.queryClient, {
			queryKey: key, queryFn: () => refetch.promise, retry: false, enabled: false,
		});
		const unsubscribe = observer.subscribe(() => undefined);
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			await settle(stale, "success", { pinned: false }, new Error("unused"));
			await waitFor(() => expect(invalidate).toHaveBeenCalledWith({ queryKey: key, exact: true }));
			await waitFor(() => expect(observer.getCurrentResult().isFetching).toBe(true));
			expect(unpinButton()).toBeDisabled();
			act(() => useAuthStore.setState({ user: userB }));
			act(() => useAuthStore.setState({ user: userA }));
			fireEvent.click(await screen.findByRole("button", { name: UNPIN_LABEL }));
			await expectPost(2);
			expect(unpinButton()).toBeDisabled();
			await act(async () => { refetch.resolve([pin()]); await refetch.promise; });
			expect(unpinButton()).toBeDisabled();
			await settle(current, "success", { pinned: false }, new Error("unused"));
			await waitFor(() => expect(unpinButton()).toBeEnabled());
		} finally {
			unsubscribe();
			invalidate.mockRestore();
		}
	});

	it.each([["success"], ["error"]] as const)("remembers a batched actor A-B-A %s loss and requires retry after stale settlement", async (outcome) => {
		const stale = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(stale.promise).mockResolvedValueOnce({ pinned: false });
		const view = renderPanel();
		const invalidate = vi.spyOn(view.queryClient, "invalidateQueries");
		try {
			fireEvent.click(await openUnpin());
			await expectPost(1);
			act(() => { useAuthStore.setState({ user: userB }); useAuthStore.setState({ user: userA }); });
			await settle(stale, outcome, { pinned: false }, new Error("batched actor loss"));
			expect(invalidate).not.toHaveBeenCalled();
			expect(toastApiError).not.toHaveBeenCalled();
			await waitFor(() => expect(unpinButton()).toBeEnabled());
			fireEvent.click(unpinButton());
			await expectPost(2);
		} finally { invalidate.mockRestore(); }
	});
});
