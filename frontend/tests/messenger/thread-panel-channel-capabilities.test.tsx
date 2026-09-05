import type { QueryClient } from "@tanstack/react-query";
import { act, cleanup, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel, ChannelCapabilities } from "@/api/messenger";
import { ChatView } from "@/components/messenger/chat-view";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const channelId = "channel-thread-capabilities";
const otherChannelId = "channel-other";
const rootMessageId = "root-thread-capabilities";
const api = vi.hoisted(() => ({ getChannel: vi.fn(), markRead: vi.fn() }));
const threadPanelWitness = vi.hoisted(() => ({
	capabilities: [] as Array<ChannelCapabilities | undefined>,
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, getChannel: api.getChannel };
});
vi.mock("@/hooks/use-messages", () => ({
	useMessages: () => ({ messages: [], loadMore: () => undefined, hasMore: false, isLoading: false }),
}));
vi.mock("@/hooks/use-typing", () => ({ useTypingIndicator: () => undefined }));
vi.mock("@/hooks/use-unread", () => ({ useMarkRead: () => api.markRead }));
vi.mock("@/components/calls/CallButton", () => ({ CallButton: () => null }));
vi.mock("@/components/messenger/invite-link-dialog", () => ({ InviteLinkDialog: () => null }));
vi.mock("@/components/messenger/message-composer", () => ({ MessageComposer: () => null }));
vi.mock("@/components/messenger/message-item", () => ({ MessageItem: () => null }));
vi.mock("@/components/messenger/pinned-messages", () => ({ PinnedMessages: () => null }));
vi.mock("@/components/messenger/search-dialog", () => ({ SearchDialog: () => null }));
vi.mock("@/components/messenger/thread-panel", () => ({
	ThreadPanel: ({ capabilities }: { capabilities?: ChannelCapabilities }) => {
		threadPanelWitness.capabilities.push(capabilities);
		return (
			<output
				data-testid="thread-panel-observer"
				data-capabilities={capabilities === undefined ? "undefined" : capabilities.can_delete_message ? "allowed" : "denied"}
			/>
		);
	},
}));

function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
		can_read: true,
		can_post: false,
		can_reply: false,
		can_join: false,
		can_follow: false,
		can_unfollow: false,
		can_view_unread: false,
		can_type: false,
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

function channel(name: string, projection?: ChannelCapabilities): Channel {
	return {
		id: channelId,
		name,
		channel_type: "private",
		is_readonly: false,
		capabilities: projection,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-09-06T00:00:00Z",
		updated_at: "2026-09-06T00:00:00Z",
	};
}

function deferred<T>() {
	let resolve!: (value: T) => void;
	const promise = new Promise<T>((resolvePromise) => { resolve = resolvePromise; });
	return { promise, resolve };
}

function latestThreadPanelCapabilities() {
	return threadPanelWitness.capabilities[threadPanelWitness.capabilities.length - 1];
}

describe("ChatView thread channel-capability boundary", () => {
	let queryClient: QueryClient | undefined;
	beforeEach(() => {
		api.getChannel.mockReset();
		api.markRead.mockReset().mockResolvedValue(undefined);
		threadPanelWitness.capabilities.length = 0;
		useMessengerStore.getState().reset();
	});
	afterEach(() => {
		cleanup();
		queryClient?.clear();
		queryClient = undefined;
		threadPanelWitness.capabilities.length = 0;
		useMessengerStore.getState().reset();
		vi.clearAllMocks();
	});
	function renderChatView() {
		const result = renderWithProviders(<ChatView channelId={channelId} />);
		queryClient = result.queryClient;
		return result;
	}

	it("keeps a pending projection undefined before forwarding the resolved matching snapshot", async () => {
		const pending = deferred<Channel>();
		const projection = capabilities({ can_delete_message: true });
		api.getChannel.mockReturnValueOnce(pending.promise);
		useMessengerStore.getState().openThread(channelId, rootMessageId);
		renderChatView();
		await screen.findByTestId("thread-panel-observer");
		await waitFor(() => expect(api.getChannel).toHaveBeenCalledWith(channelId));
		expect(latestThreadPanelCapabilities()).toBeUndefined();
		await act(async () => {
			pending.resolve(channel("Resolved thread channel", projection));
			await pending.promise;
		});
		await screen.findByRole("heading", { name: "Resolved thread channel" });
		await waitFor(() => expect(latestThreadPanelCapabilities()).toBe(projection));
		expect(screen.getByTestId("thread-panel-observer")).toHaveAttribute("data-capabilities", "allowed");
	});

	it("does not fabricate a projection after a channel without one resolves", async () => {
		api.getChannel.mockResolvedValueOnce(channel("Thread channel without projection"));
		useMessengerStore.getState().openThread(channelId, rootMessageId);
		renderChatView();
		await screen.findByRole("heading", { name: "Thread channel without projection" });
		expect(screen.getByTestId("thread-panel-observer")).toHaveAttribute("data-capabilities", "undefined");
		expect(latestThreadPanelCapabilities()).toBeUndefined();
	});

	it("does not open a thread owned by another channel", async () => {
		api.getChannel.mockResolvedValueOnce(channel("Current channel"));
		useMessengerStore.getState().openThread(otherChannelId, rootMessageId);
		renderChatView();
		await screen.findByRole("heading", { name: "Current channel" });
		expect(screen.queryByTestId("thread-panel-observer")).toBeNull();
	});

	it("hides an already observed thread when live access is revoked", async () => {
		const projection = capabilities({ can_delete_message: true });
		api.getChannel.mockResolvedValueOnce(channel("Revocable channel", projection));
		useMessengerStore.getState().openThread(channelId, rootMessageId);
		renderChatView();
		await screen.findByRole("heading", { name: "Revocable channel" });
		await waitFor(() => expect(latestThreadPanelCapabilities()).toBe(projection));
		act(() => useMessengerStore.getState().revokeChannel(channelId));
		expect(screen.queryByTestId("thread-panel-observer")).toBeNull();
	});
});
