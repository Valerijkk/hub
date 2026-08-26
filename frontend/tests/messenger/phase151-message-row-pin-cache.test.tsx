import { QueryClient } from "@tanstack/react-query";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
	togglePin,
	type Channel,
	type ChannelCapabilities,
	type Message,
} from "@/api/messenger";
import { MessageItem } from "@/components/messenger/message-item";
import { pinsQueryKey } from "@/components/messenger/pinned-messages";
import { useAIStore } from "@/stores/ai-store";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, togglePin: vi.fn() };
});

vi.mock("@/components/messenger/forward-dialog", () => ({
	ForwardDialog: () => null,
}));

const channelId = "row-pin-channel";
const userA = {
	id: "row-pin-user-a",
	email: "row-pin-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = {
	...userA,
	id: "row-pin-user-b",
	email: "row-pin-b@example.test",
};
const message: Message = {
	id: "row-pin-message",
	channel_id: channelId,
	sender_id: userA.id,
	content_text: "Pin this row",
	reply_count: 0,
	created_at: "2026-08-26T00:00:00Z",
};

function capabilities(
	overrides: Partial<ChannelCapabilities> = {},
): ChannelCapabilities {
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
		can_pin: true,
		can_list_pins: true,
		can_bookmark: false,
		can_promote: false,
		...overrides,
	};
}

function channel(currentCapabilities: ChannelCapabilities): Channel {
	return {
		id: channelId,
		name: "Row pin channel",
		channel_type: "private",
		is_readonly: false,
		capabilities: currentCapabilities,
		created_by: userA.id,
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
	};
}

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

function createQueryClient() {
	return new QueryClient({
		defaultOptions: {
			queries: { retry: false },
			mutations: { retry: false },
		},
	});
}

function renderMessageItem(
	queryClient: QueryClient,
	currentCapabilities: ChannelCapabilities,
) {
	return renderWithProviders(
		<MessageItem
			message={message}
			channelId={channelId}
			capabilities={currentCapabilities}
		/>,
		{ queryClient },
	);
}

function clickPin() {
	fireEvent.click(screen.getByRole("button", { name: "Закрепить сообщение" }));
}

describe("Phase 151 message-row pin cache synchronization", () => {
	const queryClients: QueryClient[] = [];

	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(togglePin).mockReset().mockResolvedValue({ pinned: true });
		useAIStore.setState({ loaded: true, botLoadError: false });
		useAuthStore.setState({ user: userA });
		useMessengerStore.getState().reset();
		useMessengerStore.getState().setChannels([channel(capabilities())]);
	});

	afterEach(() => {
		for (const queryClient of queryClients.splice(0)) queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		useAIStore.setState({ loaded: false });
	});

	it("invalidates only the exact current-user channel pins key after success", async () => {
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		renderMessageItem(queryClient, capabilities());

		clickPin();

		await waitFor(() =>
			expect(togglePin).toHaveBeenCalledWith(channelId, message.id),
		);
		await waitFor(() =>
			expect(invalidate).toHaveBeenCalledWith({
				queryKey: pinsQueryKey(channelId, userA.id),
				exact: true,
			}),
		);
		expect(invalidate).toHaveBeenCalledTimes(1);
	});

	it("makes no pin request or cache invalidation without an authenticated identity", async () => {
		useAuthStore.setState({ user: null });
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		renderMessageItem(queryClient, capabilities());

		clickPin();
		await act(async () => Promise.resolve());

		expect(togglePin).not.toHaveBeenCalled();
		expect(invalidate).not.toHaveBeenCalled();
	});

	it("does not invalidate A pins when identity changes during the request", async () => {
		const pending = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(pending.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		renderMessageItem(queryClient, capabilities());

		clickPin();
		await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
		act(() => useAuthStore.setState({ user: userB }));
		await act(async () => {
			pending.resolve({ pinned: true });
			await pending.promise;
		});

		expect(invalidate).not.toHaveBeenCalled();
	});

	it("does not invalidate when the latest DTO loses can_pin during the request", async () => {
		const initialCapabilities = capabilities();
		const pending = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(pending.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const { rerender } = renderMessageItem(queryClient, initialCapabilities);

		clickPin();
		await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
		rerender(
			<MessageItem
				message={message}
				channelId={channelId}
				capabilities={capabilities({ can_pin: false })}
			/>,
		);
		await act(async () => {
			pending.resolve({ pinned: true });
			await pending.promise;
		});

		expect(invalidate).not.toHaveBeenCalled();
	});

	it("does not invalidate when current store can_pin is lost during the request", async () => {
		const pending = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(pending.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		renderMessageItem(queryClient, capabilities());

		clickPin();
		await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
		act(() => {
			useMessengerStore.getState().updateChannel(channelId, {
				capabilities: capabilities({ can_pin: false }),
			});
		});
		await act(async () => {
			pending.resolve({ pinned: true });
			await pending.promise;
		});

		expect(invalidate).not.toHaveBeenCalled();
	});

	it("does not invalidate when the channel is revoked during the request", async () => {
		const pending = deferred<{ pinned: boolean }>();
		vi.mocked(togglePin).mockReturnValueOnce(pending.promise);
		const queryClient = createQueryClient();
		queryClients.push(queryClient);
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		renderMessageItem(queryClient, capabilities());

		clickPin();
		await waitFor(() => expect(togglePin).toHaveBeenCalledTimes(1));
		act(() => useMessengerStore.getState().revokeChannel(channelId));
		await act(async () => {
			pending.resolve({ pinned: true });
			await pending.promise;
		});

		expect(invalidate).not.toHaveBeenCalled();
	});
});
