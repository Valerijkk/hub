import { QueryClient } from "@tanstack/react-query";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import i18n from "@/i18n";
import { deleteMessage, type Channel, type ChannelCapabilities, type Message } from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { toastApiError } from "@/lib/format-api-error";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const api = vi.hoisted(() => ({ deleteMessage: vi.fn() }));
const errors = vi.hoisted(() => ({ toastApiError: vi.fn() }));
vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, deleteMessage: api.deleteMessage };
});
vi.mock("@/lib/format-api-error", () => ({ toastApiError: errors.toastApiError }));
vi.mock("@/components/messenger/forward-dialog", () => ({ ForwardDialog: () => null }));

const CHANNEL_A = "delete-channel-a";
const CHANNEL_B = "delete-channel-b";
const DELETE_LABEL = "Удалить сообщение";
const REPLY_LABEL = "Ответить";
const DELETE_FALLBACK = "Unable to delete message";
const userA = {
	id: "delete-user-a",
	email: "delete-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = { ...userA, id: "delete-user-b", email: "delete-user-b@example.test" };
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
		id: "delete-message-a",
		channel_id: CHANNEL_A,
		sender_id: userA.id,
		content_text: "Delete lifecycle message",
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
	channelId = CHANNEL_A,
	itemCapabilities = capabilities(),
) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return renderWithProviders(
		<MessageItem message={item} channelId={channelId} capabilities={itemCapabilities} />,
		{ queryClient },
	);
}
function renderMessageItemWithoutCapabilities(item = message(), channelId = CHANNEL_A) {
	const queryClient = createQueryClient();
	queryClients.push(queryClient);
	return renderWithProviders(<MessageItem message={item} channelId={channelId} />, { queryClient });
}
function currentDeleteButton() {
	return screen.getByRole("button", { name: DELETE_LABEL });
}

describe("message delete request lifecycle", () => {
	beforeEach(async () => {
		vi.clearAllMocks();
		vi.mocked(deleteMessage).mockReset().mockResolvedValue(undefined);
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

	it.each([
		{
			name: "an own ordinary-member message",
			item: message({ id: "own-delete-message", sender_id: userA.id }),
			currentCapabilities: capabilities(),
		},
		{
			name: "a foreign message with delete-any capability",
			item: message({ id: "foreign-delete-message", sender_id: userB.id }),
			currentCapabilities: capabilities({ can_delete_message: true }),
		},
	])("issues one exact DELETE for $name", async ({ item, currentCapabilities }) => {
		installChannels([{ id: CHANNEL_A, currentCapabilities }]);
		renderMessageItem(item, CHANNEL_A, currentCapabilities);
		fireEvent.click(currentDeleteButton());
		await waitFor(() => expect(deleteMessage).toHaveBeenCalledWith(CHANNEL_A, item.id));
		expect(deleteMessage).toHaveBeenCalledTimes(1);
	});

	it("hides Delete for a foreign ordinary member and sends no request", () => {
		const currentCapabilities = capabilities();
		installChannels([{ id: CHANNEL_A, currentCapabilities }]);
		renderMessageItem(message({ sender_id: userB.id }), CHANNEL_A, currentCapabilities);
		expect(screen.queryByRole("button", { name: DELETE_LABEL })).not.toBeInTheDocument();
		expect(deleteMessage).not.toHaveBeenCalled();
	});
	it("hides Delete when the server capability projection is missing", () => {
		installChannels([{ id: CHANNEL_A, currentCapabilities: capabilities() }]);
		renderMessageItemWithoutCapabilities();
		expect(screen.queryByRole("button", { name: DELETE_LABEL })).not.toBeInTheDocument();
		expect(deleteMessage).not.toHaveBeenCalled();
	});

	it.each([
		{
			name: "the live channel capability is absent",
			arrange: () => installChannels([]),
			item: message(),
		},
		{
			name: "the rendered message belongs to another channel",
			arrange: () => installChannels([
				{ id: CHANNEL_A, currentCapabilities: capabilities({ can_delete_message: true }) },
			]),
			item: message({ id: "mismatched-channel-message", channel_id: CHANNEL_B }),
		},
		{
			name: "the authenticated actor is absent",
			arrange: () => {
				useAuthStore.setState({ isAuthenticated: false, user: null });
				installChannels([
					{ id: CHANNEL_A, currentCapabilities: capabilities({ can_delete_message: true }) },
				]);
			},
			item: message({ sender_id: userB.id }),
		},
	])("rechecks live authorization before transport when $name", async ({ arrange, item }) => {
		const projectedCapabilities = capabilities({ can_delete_message: true });
		arrange();
		renderMessageItem(item, CHANNEL_A, projectedCapabilities);
		fireEvent.click(currentDeleteButton());
		await act(async () => Promise.resolve());
		expect(deleteMessage).not.toHaveBeenCalled();
	});
	it("sends no request after the channel is live-revoked", () => {
		const currentCapabilities = capabilities({ can_delete_message: true });
		installChannels([{ id: CHANNEL_A, currentCapabilities }]);
		renderMessageItem(message({ sender_id: userB.id }), CHANNEL_A, currentCapabilities);
		act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
		expect(screen.queryByRole("button", { name: DELETE_LABEL })).not.toBeInTheDocument();
		expect(deleteMessage).not.toHaveBeenCalled();
	});

	it("deduplicates repeated native button click activation while disabling only Delete", async () => {
		const request = deferred<void>();
		const currentCapabilities = capabilities({ can_delete_message: true });
		vi.mocked(deleteMessage).mockReturnValueOnce(request.promise);
		installChannels([{ id: CHANNEL_A, currentCapabilities }]);
		renderMessageItem(message({ sender_id: userB.id }), CHANNEL_A, currentCapabilities);
		const firstDeleteButton = currentDeleteButton();
		fireEvent.click(firstDeleteButton);
		fireEvent.click(firstDeleteButton);
		fireEvent.click(firstDeleteButton);
		await waitFor(() => expect(deleteMessage).toHaveBeenCalledTimes(1));
		const pendingDeleteButton = currentDeleteButton();
		expect(pendingDeleteButton).toBeDisabled();
		expect(pendingDeleteButton).toHaveAttribute("type", "button");
		expect(screen.getByRole("button", { name: REPLY_LABEL })).toBeEnabled();
		await act(async () => {
			request.resolve();
			await request.promise;
		});
		await waitFor(() => expect(currentDeleteButton()).toBeEnabled());
	});

	it("passes the translated fallback to the canonical formatter and retries only explicitly", async () => {
		const failure = new Error("transport failed");
		const unhandled = vi.fn();
		window.addEventListener("unhandledrejection", unhandled);
		try {
			vi.mocked(deleteMessage).mockRejectedValueOnce(failure).mockResolvedValueOnce(undefined);
			const currentCapabilities = capabilities({ can_delete_message: true });
			installChannels([{ id: CHANNEL_A, currentCapabilities }]);
			renderMessageItem(message({ sender_id: userB.id }), CHANNEL_A, currentCapabilities);
			fireEvent.click(currentDeleteButton());
			await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(failure, DELETE_FALLBACK));
			expect(deleteMessage).toHaveBeenCalledTimes(1);
			expect(currentDeleteButton()).toBeEnabled();
			fireEvent.click(currentDeleteButton());
			await waitFor(() => expect(deleteMessage).toHaveBeenCalledTimes(2));
			expect(toastApiError).toHaveBeenCalledTimes(1);
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it.each([
		["user", "success"], ["user", "error"],
		["channel", "success"], ["channel", "error"],
		["message", "success"], ["message", "error"],
	] as const)("does not let an old %s A-B-A %s completion unlock a newer request", async (boundary, outcome) => {
		const stale = deferred<void>();
		const current = deferred<void>();
		const staleError = new Error(`stale ${boundary} completion`);
		const unhandled = vi.fn();
		const currentCapabilities = capabilities({ can_delete_message: true });
		const initialMessage = message({ id: "aba-message-a", sender_id: userB.id, channel_id: CHANNEL_A });
		window.addEventListener("unhandledrejection", unhandled);
		try {
			vi.mocked(deleteMessage).mockReturnValueOnce(stale.promise).mockReturnValueOnce(current.promise);
			installChannels([
				{ id: CHANNEL_A, currentCapabilities },
				{ id: CHANNEL_B, currentCapabilities },
			]);
			const view = renderMessageItem(initialMessage, CHANNEL_A, currentCapabilities);
			fireEvent.click(currentDeleteButton());
			await waitFor(() =>
				expect(deleteMessage).toHaveBeenNthCalledWith(1, CHANNEL_A, initialMessage.id),
			);
			const staleSettled = stale.promise.then(
				() => undefined,
				() => undefined,
			);
			if (boundary === "user") {
				act(() => useAuthStore.setState({ user: userB }));
				act(() => useAuthStore.setState({ user: userA }));
			} else if (boundary === "channel") {
				view.rerender(
					<MessageItem
						message={message({ id: "aba-message-b", sender_id: userB.id, channel_id: CHANNEL_B })}
						channelId={CHANNEL_B}
						capabilities={currentCapabilities}
					/>,
				);
				view.rerender(
					<MessageItem message={initialMessage} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
				);
			} else {
				view.rerender(
					<MessageItem
						message={message({ id: "aba-message-b", sender_id: userB.id, channel_id: CHANNEL_A })}
						channelId={CHANNEL_A}
						capabilities={currentCapabilities}
					/>,
				);
				view.rerender(
					<MessageItem message={initialMessage} channelId={CHANNEL_A} capabilities={currentCapabilities} />,
				);
			}
			expect(currentDeleteButton()).toBeEnabled();
			fireEvent.click(currentDeleteButton());
			await waitFor(() =>
				expect(deleteMessage).toHaveBeenNthCalledWith(2, CHANNEL_A, initialMessage.id),
			);
			expect(currentDeleteButton()).toBeDisabled();
			await act(async () => {
				if (outcome === "success") stale.resolve();
				else stale.reject(staleError);
				await staleSettled;
			});
			expect(toastApiError).not.toHaveBeenCalled();
			expect(currentDeleteButton()).toBeDisabled();
			expect(unhandled).not.toHaveBeenCalled();
			await act(async () => {
				current.resolve();
				await current.promise;
			});
			await waitFor(() => expect(currentDeleteButton()).toBeEnabled());
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});

	it.each([
		["revoke", "success"], ["revoke", "error"],
		["unmount", "success"], ["unmount", "error"],
	] as const)("silently consumes late %s completion after %s", async (boundary, outcome) => {
		const request = deferred<void>();
		const lateError = new Error(`late ${boundary} completion`);
		const unhandled = vi.fn();
		const currentCapabilities = capabilities({ can_delete_message: true });
		window.addEventListener("unhandledrejection", unhandled);
		try {
			vi.mocked(deleteMessage).mockReturnValueOnce(request.promise);
			installChannels([{ id: CHANNEL_A, currentCapabilities }]);
			const view = renderMessageItem(message({ sender_id: userB.id }), CHANNEL_A, currentCapabilities);
			fireEvent.click(currentDeleteButton());
			await waitFor(() => expect(deleteMessage).toHaveBeenCalledTimes(1));
			const requestSettled = request.promise.then(
				() => undefined,
				() => undefined,
			);
			if (boundary === "revoke") {
				act(() => useMessengerStore.getState().revokeChannel(CHANNEL_A));
				expect(screen.queryByRole("button", { name: DELETE_LABEL })).not.toBeInTheDocument();
			} else {
				view.unmount();
			}
			await act(async () => {
				if (outcome === "success") request.resolve();
				else request.reject(lateError);
				await requestSettled;
			});
			expect(toastApiError).not.toHaveBeenCalled();
			expect(unhandled).not.toHaveBeenCalled();
		} finally {
			window.removeEventListener("unhandledrejection", unhandled);
		}
	});
});
