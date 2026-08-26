import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, render, renderHook, waitFor } from "@testing-library/react";
import type { ComponentType, ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel, Message } from "@/api/messenger";
import { useMessengerChannelsRealtime } from "@/hooks/use-messenger-channels-realtime";
import {
	messagesQueryKey as rootMessagesQueryKey,
	useMessages,
} from "@/hooks/use-messages";
import { Route } from "@/routes/_authenticated/messenger/$channelId";
import { useAuthStore } from "@/stores/auth-store";
import {
	emptyUnreadSnapshot,
	useMessengerStore,
} from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

const api = vi.hoisted(() => ({
	getChannels: vi.fn(),
	getMessages: vi.fn(),
}));

const navigation = vi.hoisted(() => ({
	navigate: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannels: api.getChannels,
		getMessages: api.getMessages,
	};
});

vi.mock("@tanstack/react-router", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-router")>();
	return {
		...actual,
		useNavigate: () => navigation.navigate,
	};
});

vi.mock("@/components/messenger/chat-view", () => ({
	ChatView: () => <div data-testid="chat-view" />,
}));

type RevokeBarrierState = ReturnType<typeof useMessengerStore.getState> & {
	revokedChannelIds: ReadonlySet<string>;
	revokeChannel: (channelId: string) => void;
	isChannelRevoked: (channelId: string) => boolean;
};

type MessageCache = {
	pages: Message[][];
	pageParams: unknown[];
};

const revokedChannelId = "channel-revoked";
const retainedChannelId = "channel-retained";
const messagesQueryKey = rootMessagesQueryKey(revokedChannelId, "user-1");

function messengerState(): RevokeBarrierState {
	return useMessengerStore.getState() as RevokeBarrierState;
}

function channel(id: string): Channel {
	return {
		id,
		name: id,
		channel_type: "private",
		is_readonly: false,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
	};
}

function message(id: string, channelId = revokedChannelId): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "user-2",
		content_text: id,
		content_json: { type: "doc", text: id },
		reply_count: 0,
		created_at: "2026-08-26T00:00:00Z",
	};
}

function publish(event: string, payload: unknown): void {
	act(() => {
		for (const listener of [...useWsStore.getState().listeners]) {
			listener(event, payload);
		}
	});
}

describe("P151 channel_removed revoke barrier", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		api.getChannels.mockResolvedValue([channel(revokedChannelId)]);
		api.getMessages.mockResolvedValue([]);
		navigation.navigate.mockReset();
		const realtimeSocket = {
			readyState: WebSocket.OPEN,
			send: vi.fn(),
		} as unknown as WebSocket;
		useWsStore.setState({
			connected: true,
			socket: realtimeSocket,
			listeners: new Set(),
		});
		useAuthStore.setState({
			accessToken: "revoke-test-token",
			refreshToken: "revoke-test-refresh",
			isAuthenticated: true,
			user: {
				id: "user-1",
				email: "user-1@hub.local",
				timezone: "UTC",
				language: "en",
				system_role: "member",
				onboarding_completed: true,
			},
		});
		useMessengerStore.getState().reset();
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null, isAuthenticated: false });
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
		useMessengerStore.getState().reset();
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	it("revokes all channel-bound store state in one atomic transition", () => {
		const reply = message("reply-to-revoked");
		useMessengerStore.setState({
			activeChannelId: revokedChannelId,
			channels: [channel(revokedChannelId), channel(retainedChannelId)],
			unreadSnapshot: {
				channels: { [revokedChannelId]: 3, [retainedChannelId]: 1 },
				dms: { [revokedChannelId]: 2 },
				mentions: { [revokedChannelId]: 1 },
				threads: { "root-revoked": 4 },
				any_unread: true,
			},
			activeThread: { channelId: revokedChannelId, rootId: "root-revoked" },
			replyToMessage: reply,
			typingUsers: {
				[revokedChannelId]: [
					{ userId: "user-2", displayName: "User Two", timestamp: 1 },
				],
				[retainedChannelId]: [
					{ userId: "user-3", displayName: "User Three", timestamp: 1 },
				],
			},
		});
		const subscriber = vi.fn();
		const unsubscribe = useMessengerStore.subscribe(subscriber);

		act(() => {
			messengerState().revokeChannel(revokedChannelId);
		});

		expect(subscriber).toHaveBeenCalledTimes(1);
		expect(messengerState().revokedChannelIds.has(revokedChannelId)).toBe(true);
		expect(messengerState().isChannelRevoked(revokedChannelId)).toBe(true);
		expect(messengerState().channels.map((item) => item.id)).toEqual([
			retainedChannelId,
		]);
		expect(messengerState().activeChannelId).toBeNull();
		expect(messengerState().activeThread).toBeNull();
		expect(messengerState().replyToMessage).toBeNull();
		expect(messengerState().typingUsers).toEqual({
			[retainedChannelId]: [
				{ userId: "user-3", displayName: "User Three", timestamp: 1 },
			],
		});
		expect(messengerState().unreadSnapshot).toEqual(emptyUnreadSnapshot());

		unsubscribe();
	});

	it("cancels a late fetch, removes protected cache keys, and ignores a queued message event", async () => {
		let resolveMessages: (messages: Message[]) => void = () => undefined;
		const lateMessages = new Promise<Message[]>((resolve) => {
			resolveMessages = resolve;
		});
		api.getMessages.mockImplementationOnce(() => lateMessages);
		const cancelSpy = vi.spyOn(queryClient, "cancelQueries");
		const removeSpy = vi.spyOn(queryClient, "removeQueries");

		renderHook(
			() => {
				useMessengerChannelsRealtime();
				return useMessages(revokedChannelId);
			},
			{ wrapper: TestProvider },
		);

		await waitFor(() => {
			expect(api.getMessages).toHaveBeenCalledTimes(1);
			expect(useWsStore.getState().listeners.size).toBeGreaterThanOrEqual(2);
		});

		publish("channel_removed", { channel_id: revokedChannelId });

		expect(messengerState().isChannelRevoked(revokedChannelId)).toBe(true);
		expect(cancelSpy).toHaveBeenCalledWith(
			expect.objectContaining({ predicate: expect.any(Function) }),
		);
		expect(removeSpy).toHaveBeenCalledWith(
			expect.objectContaining({ predicate: expect.any(Function) }),
		);

		await act(async () => {
			resolveMessages([message("late-fetch")]);
			await lateMessages;
		});
		await waitFor(() => {
			expect(queryClient.getQueryData(messagesQueryKey)).toBeUndefined();
		});

		const staleCache: MessageCache = {
			pages: [[message("stale-before-queued-event")]],
			pageParams: [undefined],
		};
		queryClient.setQueryData(messagesQueryKey, staleCache);
		publish("message_sent", { message: message("queued-after-revoke") });

		expect(queryClient.getQueryData(messagesQueryKey)).toBe(staleCache);
	});

	it("redirects an active /messenger/:channelId route to the channel index after revoke", async () => {
		const paramsSpy = vi
			.spyOn(Route, "useParams")
			.mockReturnValue({ channelId: revokedChannelId } as never);
		const ChannelPage = Route.options.component as ComponentType;

		render(<ChannelPage />);
		await waitFor(() => {
			expect(messengerState().activeChannelId).toBe(revokedChannelId);
		});

		act(() => {
			messengerState().revokeChannel(revokedChannelId);
		});

		await waitFor(() => {
			expect(navigation.navigate).toHaveBeenCalledWith({
				to: "/messenger",
				replace: true,
			});
		});
		paramsSpy.mockRestore();
	});
	it("clears the local revoke barrier only after an authorized channel_created event", async () => {
		renderHook(() => useMessengerChannelsRealtime(), { wrapper: TestProvider });

		await waitFor(() => {
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0);
		});

		act(() => {
			messengerState().revokeChannel(revokedChannelId);
			messengerState().setChannels([channel(revokedChannelId)]);
		});
		expect(messengerState().channels).toEqual([]);
		expect(messengerState().isChannelRevoked(revokedChannelId)).toBe(true);

		publish("channel_created", { channel_id: revokedChannelId });

		expect(messengerState().isChannelRevoked(revokedChannelId)).toBe(false);
		act(() => {
			messengerState().setChannels([channel(revokedChannelId)]);
		});
		expect(messengerState().channels.map((item) => item.id)).toEqual([
			revokedChannelId,
		]);
	});
});
