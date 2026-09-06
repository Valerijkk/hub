import { QueryClient } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import {
	toggleReaction,
	type Channel,
	type ChannelCapabilities,
	type Message,
} from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { messagesQueryKey } from "@/hooks/use-messages";
import { toastApiError } from "@/lib/format-api-error";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ toggleReaction: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, toggleReaction: api.toggleReaction };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@/components/messenger/forward-dialog", () => ({ ForwardDialog: () => null }));
vi.mock("@/components/messenger/link-preview", () => ({ LinkPreview: () => null }));
vi.mock("@/components/messenger/message-attachments", () => ({ MessageAttachments: () => null }));
vi.mock("@/components/messenger/poll-message", () => ({ PollMessage: () => null }));
vi.mock("@/components/messenger/render-content", () => ({ RenderContent: () => null }));

const CHANNEL_A = "reaction-channel-a";
const CHANNEL_B = "reaction-channel-b";
const REACTION_LABEL = "Реакция";
const THUMBS_UP = "👍";
const FIRE = "🔥";
const REPLY_LABEL = "Ответить";
const DELETE_LABEL = "Удалить сообщение";
const ERROR_EN = "Unable to update reaction";
const ERROR_RU = "Не удалось изменить реакцию";

const userA = {
	id: "reaction-user-a",
	email: "reaction-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = { ...userA, id: "reaction-user-b", email: "reaction-user-b@example.test" };
type Timeline = { pages: Message[][]; pageParams: unknown[] };
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
		can_react: true,
		can_delete_message: true,
		can_pin: false,
		can_list_pins: false,
		can_bookmark: false,
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
		id: "reaction-message-a",
		channel_id: CHANNEL_A,
		sender_id: userA.id,
		content_text: "Reaction lifecycle message",
		reply_count: 0,
		created_at: "2026-09-06T00:00:00Z",
		...overrides,
	};
}

function timeline(items: Message[]): Timeline {
	return { pages: [items], pageParams: [undefined] };
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

function reactionTrigger() {
	return screen.getByRole("button", { name: REACTION_LABEL });
}

function pickerEmoji(emoji = THUMBS_UP) {
	return screen.getByRole("button", { name: emoji });
}

function reactionPill(emoji: string) {
	return screen.getByRole("button", { name: new RegExp(`^${emoji}\\s*\\d+$`) });
}

function openPicker() {
	fireEvent.click(reactionTrigger());
	return pickerEmoji();
}

describe("message reaction request lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(toggleReaction).mockReset().mockResolvedValue({ added: true });
		vi.mocked(toastApiError).mockReset();
		await i18n.changeLanguage("en");
		await i18n.loadNamespaces("messenger");
		useAIStore.setState({ loaded: true, botLoadError: false });
		useAuthStore.setState({
			accessToken: null,
			refreshToken: null,
			isAuthenticated: true,
			user: userA,
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
			accessToken: null,
			refreshToken: null,
			isAuthenticated: false,
			user: null,
		});
		useMessengerStore.getState().reset();
		useAIStore.setState({ loaded: false, botLoadError: false });
		await i18n.changeLanguage("ru");
	});

	it("sends the exact picker request, closes its current session, and leaves the server cache authoritative", async () => {
		const item = message({ id: "picker-success-message" });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		const snapshot = structuredClone(origin);
		const { queryClient } = renderMessageItem(item);
		queryClient.setQueryData(originKey, origin);

		fireEvent.click(openPicker());
		await waitFor(() => {
			expect(toggleReaction).toHaveBeenCalledWith(CHANNEL_A, item.id, THUMBS_UP);
		});
		expect(toggleReaction).toHaveBeenCalledTimes(1);
		await waitFor(() => expect(screen.queryByRole("button", { name: THUMBS_UP })).toBeNull());
		expect(queryClient.getQueryData(originKey)).toBe(origin);
		expect(queryClient.getQueryData(originKey)).toEqual(snapshot);
	});

	it("sends an existing reaction pill from another author through the same message row", async () => {
		const item = message({
			id: "foreign-pill-message",
			sender_id: userB.id,
			reactions: [{ emoji: FIRE, user_ids: [userB.id] }],
		});
		renderMessageItem(item);

		fireEvent.click(reactionPill(FIRE));
		await waitFor(() => {
			expect(toggleReaction).toHaveBeenCalledWith(CHANNEL_A, item.id, FIRE);
		});
		expect(toggleReaction).toHaveBeenCalledTimes(1);
	});

	it.each([
		{ entrypoint: "picker", language: "en", fallback: ERROR_EN },
		{ entrypoint: "picker", language: "ru", fallback: ERROR_RU },
		{ entrypoint: "pill", language: "en", fallback: ERROR_EN },
		{ entrypoint: "pill", language: "ru", fallback: ERROR_RU },
	] as const)("reports one current $language rejection from the $entrypoint entrypoint", async ({ entrypoint, language, fallback }) => {
		const failure = new Error(`${entrypoint} ${language} failure`);
		const item = message({
			id: `rejected-${entrypoint}-${language}`,
			sender_id: userB.id,
			reactions: entrypoint === "pill" ? [{ emoji: FIRE, user_ids: [userB.id] }] : undefined,
		});
		vi.mocked(toggleReaction).mockRejectedValueOnce(failure);
		await i18n.changeLanguage(language);
		await i18n.loadNamespaces("messenger");
		renderMessageItem(item);

		if (entrypoint === "picker") fireEvent.click(openPicker());
		else fireEvent.click(reactionPill(FIRE));

		await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, fallback));
		expect(toastApiError).toHaveBeenCalledTimes(1);
		expect(toggleReaction).toHaveBeenCalledTimes(1);
		if (entrypoint === "picker") {
			expect(pickerEmoji()).toBeInTheDocument();
		}
	});

	it("keeps a current failed picker session open and retries only after a later explicit click", async () => {
		const failure = new Error("retryable reaction failure");
		vi.mocked(toggleReaction)
			.mockRejectedValueOnce(failure)
			.mockResolvedValueOnce({ added: true });
		renderMessageItem(message({ id: "manual-retry-message" }));

		fireEvent.click(openPicker());
		await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, ERROR_EN));
		expect(toggleReaction).toHaveBeenCalledTimes(1);
		expect(pickerEmoji()).toBeInTheDocument();
		fireEvent.click(pickerEmoji());
		await waitFor(() => expect(toggleReaction).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(screen.queryByRole("button", { name: THUMBS_UP })).toBeNull());
	});

	it("uses one synchronous row lock for repeated and cross-entrypoint clicks while leaving other actions enabled", async () => {
		const request = deferred<{ added: boolean }>();
		const item = message({
			id: "same-tick-message",
			reactions: [{ emoji: FIRE, user_ids: [userB.id] }],
		});
		vi.mocked(toggleReaction).mockReturnValueOnce(request.promise);
		renderMessageItem(item);
		openPicker();

		act(() => {
			fireEvent.click(reactionPill(FIRE));
			fireEvent.click(pickerEmoji());
			fireEvent.click(reactionPill(FIRE));
		});

		expect(toggleReaction).toHaveBeenCalledTimes(1);
		expect(toggleReaction).toHaveBeenCalledWith(CHANNEL_A, item.id, FIRE);
		expect(reactionPill(FIRE)).toBeDisabled();
		expect(pickerEmoji()).toBeDisabled();
		expect(screen.getByRole("button", { name: REPLY_LABEL })).toBeEnabled();
		expect(screen.getByRole("button", { name: DELETE_LABEL })).toBeEnabled();
		await act(async () => {
			request.resolve({ added: true });
			await request.promise;
		});
	});

	it.each([
		"the authenticated actor is absent",
		"the prop capability denies reactions",
		"the live channel projection is absent",
		"the live channel capability denies reactions",
		"the channel is revoked",
		"the rendered message belongs to another channel",
	] as const)("does not send a reaction when %s", async (name) => {
		let item = message({
			id: `denied-${name}`,
			reactions: [{ emoji: FIRE, user_ids: [userB.id] }],
		});
		let itemCapabilities = capabilities();
		if (name === "the authenticated actor is absent") {
			useAuthStore.setState({ isAuthenticated: false, user: null });
		} else if (name === "the prop capability denies reactions") {
			itemCapabilities = capabilities({ can_react: false });
		} else if (name === "the live channel projection is absent") {
			useMessengerStore.getState().reset();
		} else if (name === "the live channel capability denies reactions") {
			installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities({ can_react: false }) }]);
		} else if (name === "the rendered message belongs to another channel") {
			item = message({ id: "mismatched-reaction-message", channel_id: CHANNEL_B });
		}
		renderMessageItem(item, CHANNEL_A, itemCapabilities);

		if (name === "the channel is revoked") {
			expect(reactionPill(FIRE)).toBeEnabled();
			act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
			expect(screen.queryByRole("button", { name: REACTION_LABEL })).toBeNull();
			expect(screen.queryByRole("button", { name: new RegExp(`^${FIRE}\\s*\\d+$`) })).toBeNull();
		} else if (name === "the prop capability denies reactions") {
			expect(reactionPill(FIRE)).toBeDisabled();
			fireEvent.click(reactionPill(FIRE));
		} else {
			fireEvent.click(openPicker());
		}
		await act(async () => Promise.resolve());
		expect(toggleReaction).not.toHaveBeenCalled();
	});

	it.each([
		["actor", "success"], ["actor", "rejection"],
		["channel", "success"], ["channel", "rejection"],
		["message", "success"], ["message", "rejection"],
	] as const)("keeps a newer request and picker session intact after old %s A-B-A %s", async (boundary, outcome) => {
		const stale = deferred<{ added: boolean }>();
		const current = deferred<{ added: boolean }>();
		const staleError = new Error(`stale ${boundary} reaction`);
		const initial = message({ id: "aba-reaction-message-a" });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([initial]);
		const snapshot = structuredClone(origin);
		vi.mocked(toggleReaction)
			.mockReturnValueOnce(stale.promise)
			.mockReturnValueOnce(current.promise);
		const view = renderMessageItem(initial);
		view.queryClient.setQueryData(originKey, origin);

		fireEvent.click(openPicker());
		await waitFor(() => expect(toggleReaction).toHaveBeenNthCalledWith(
			1, CHANNEL_A, initial.id, THUMBS_UP,
		));
		if (boundary === "actor") {
			act(() => useAuthStore.setState({ user: userB }));
			act(() => useAuthStore.setState({ user: userA }));
		} else if (boundary === "channel") {
			view.rerender(
				<MessageItem message={message({ id: "aba-reaction-message-b", channel_id: CHANNEL_B })} channelId={CHANNEL_B} capabilities={capabilities()} />,
			);
			view.rerender(<MessageItem message={initial} channelId={CHANNEL_A} capabilities={capabilities()} />);
		} else {
			view.rerender(
				<MessageItem message={message({ id: "aba-reaction-message-b" })} channelId={CHANNEL_A} capabilities={capabilities()} />,
			);
			view.rerender(<MessageItem message={initial} channelId={CHANNEL_A} capabilities={capabilities()} />);
		}

		expect(pickerEmoji()).toBeEnabled();
		fireEvent.click(pickerEmoji());
		await waitFor(() => expect(toggleReaction).toHaveBeenNthCalledWith(
			2, CHANNEL_A, initial.id, THUMBS_UP,
		));
		expect(pickerEmoji()).toBeDisabled();
		if (outcome === "success") {
			await act(async () => {
				stale.resolve({ added: true });
				await stale.promise;
			});
		} else {
			await act(async () => {
				stale.reject(staleError);
				await Promise.resolve();
				await Promise.resolve();
			});
		}
		expect(toastApiError).not.toHaveBeenCalled();
		expect(pickerEmoji()).toBeDisabled();
		expect(view.queryClient.getQueryData(originKey)).toBe(origin);
		expect(view.queryClient.getQueryData(originKey)).toEqual(snapshot);
		await act(async () => {
			current.resolve({ added: true });
			await current.promise;
		});
	});

	it.each([
		"capability loss and restore",
		"channel revocation",
		"unmount",
	] as const)("consumes a late rejection quietly after %s", async (name) => {
		const request = deferred<{ added: boolean }>();
		const failure = new Error(`late ${name} reaction`);
		const item = message({ id: `late-${name}` });
		const originKey = messagesQueryKey(CHANNEL_A, userA.id);
		const origin = timeline([item]);
		const snapshot = structuredClone(origin);
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		vi.mocked(toggleReaction).mockReturnValueOnce(request.promise);
		const view = renderMessageItem(item);
		view.queryClient.setQueryData(originKey, origin);
		try {
			fireEvent.click(openPicker());
			await waitFor(() => expect(toggleReaction).toHaveBeenCalledTimes(1));
			if (name === "capability loss and restore") {
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities({ can_react: false }) }]));
				act(() => installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]));
			} else if (name === "channel revocation") {
				act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
			} else {
				view.unmount();
			}
			await act(async () => {
				request.reject(failure);
				await Promise.resolve();
				await Promise.resolve();
			});
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
			expect(view.queryClient.getQueryData(originKey)).toBe(origin);
			expect(view.queryClient.getQueryData(originKey)).toEqual(snapshot);
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it("does not let an old success close a separately closed and reopened picker session", async () => {
		const request = deferred<{ added: boolean }>();
		vi.mocked(toggleReaction).mockReturnValueOnce(request.promise);
		renderMessageItem(message({ id: "reopened-picker-message" }));

		fireEvent.click(openPicker());
		await waitFor(() => expect(toggleReaction).toHaveBeenCalledTimes(1));
		expect(reactionTrigger()).toBeEnabled();
		fireEvent.click(reactionTrigger());
		expect(screen.queryByRole("button", { name: THUMBS_UP })).toBeNull();
		fireEvent.click(reactionTrigger());
		expect(pickerEmoji()).toBeInTheDocument();
		await act(async () => {
			request.resolve({ added: true });
			await request.promise;
		});
		expect(pickerEmoji()).toBeInTheDocument();
	});

	it("allows independent mounted message rows to send while another row is pending", async () => {
		const first = deferred<{ added: boolean }>();
		const second = deferred<{ added: boolean }>();
		const left = message({
			id: "independent-left-message",
			reactions: [{ emoji: FIRE, user_ids: [userB.id] }],
		});
		const right = message({
			id: "independent-right-message",
			reactions: [{ emoji: THUMBS_UP, user_ids: [userB.id] }],
		});
		vi.mocked(toggleReaction)
			.mockReturnValueOnce(first.promise)
			.mockReturnValueOnce(second.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		renderWithProviders(
			<>
				<MessageItem message={left} channelId={CHANNEL_A} capabilities={capabilities()} />
				<MessageItem message={right} channelId={CHANNEL_A} capabilities={capabilities()} />
			</>,
			{ queryClient },
		);

		act(() => {
			fireEvent.click(reactionPill(FIRE));
			fireEvent.click(reactionPill(THUMBS_UP));
		});
		expect(toggleReaction).toHaveBeenNthCalledWith(1, CHANNEL_A, left.id, FIRE);
		expect(toggleReaction).toHaveBeenNthCalledWith(2, CHANNEL_A, right.id, THUMBS_UP);
		await act(async () => {
			first.resolve({ added: true });
			second.resolve({ added: true });
			await Promise.all([first.promise, second.promise]);
		});
	});
});
