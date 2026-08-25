import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { UserResponse } from "@/api/auth";
import type { Channel, Message } from "@/api/messenger";
import { CHANNELS_QUERY_KEY } from "@/hooks/use-channels";
import { queryClient } from "@/lib/query-client";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { clearWsReconnectTimer, useWsStore } from "@/stores/ws-store";

const testUser: UserResponse = {
	id: "user-1",
	email: "user-1@hub.local",
	name: "User One",
	display_name: "User One",
	timezone: "UTC",
	language: "ru",
	system_role: "member",
	onboarding_completed: true,
};

const testChannel: Channel = {
	id: "channel-1",
	channel_type: "public",
	name: "Identity state",
	is_readonly: false,
	created_by: testUser.id,
	is_starred: false,
	created_at: "2026-08-25T00:00:00Z",
	updated_at: "2026-08-25T00:00:00Z",
};

const testMessage: Message = {
	id: "message-1",
	channel_id: testChannel.id,
	sender_id: testUser.id,
	content_text: "Reply target",
	reply_count: 0,
	created_at: "2026-08-25T00:00:00Z",
};

type WebSocketHandler = ((event?: unknown) => void) | null;

class FakeWebSocket {
	static CONNECTING = 0;
	static OPEN = 1;
	static CLOSING = 2;
	static CLOSED = 3;
	static instances: FakeWebSocket[] = [];

	readyState = FakeWebSocket.CONNECTING;
	onopen: WebSocketHandler = null;
	onclose: WebSocketHandler = null;
	onerror: WebSocketHandler = null;
	onmessage: WebSocketHandler = null;

	constructor(
		readonly url: string,
		readonly protocols?: string | string[],
	) {
		FakeWebSocket.instances.push(this);
	}

	send() {}

	close() {
		this.readyState = FakeWebSocket.CLOSED;
		this.onclose?.({});
	}

	open() {
		this.readyState = FakeWebSocket.OPEN;
		this.onopen?.({});
	}

	drop() {
		this.readyState = FakeWebSocket.CLOSED;
		this.onclose?.({});
	}
}
function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

function resetIdentityDerivedState(): void {
	queryClient.clear();
	const messenger = useMessengerStore.getState();
	messenger.setChannels([]);
	messenger.setUnreadCounts({});
	messenger.setActiveChannel(null);
	messenger.closeThread();
	messenger.setReplyTo(null);
	useMessengerStore.setState({ typingUsers: {} });
	useWsStore.setState({
		connected: false,
		socket: null,
		listeners: new Set(),
	});
}

function seedIdentityDerivedState() {
	const close = vi.fn();
	const listener = vi.fn();
	const socket = { close } as unknown as WebSocket;

	queryClient.setQueryData(CHANNELS_QUERY_KEY, [testChannel]);
	queryClient.setQueryData(["identity", testUser.id, "messenger"], {
		private: true,
	});

	const messenger = useMessengerStore.getState();
	messenger.setChannels([testChannel]);
	messenger.setUnreadCounts({ [testChannel.id]: 4 });
	messenger.setActiveChannel(testChannel.id);
	messenger.openThread(testMessage.id);
	messenger.setReplyTo(testMessage);
	messenger.addTypingUser(testChannel.id, "user-2", "Teammate");

	useWsStore.setState({
		connected: true,
		socket,
		listeners: new Set([listener]),
	});

	return { close, listener, socket };
}

function expectIdentityDerivedStateCleared(
	close: ReturnType<typeof vi.fn>,
): void {
	expect(queryClient.getQueryData(CHANNELS_QUERY_KEY)).toBeUndefined();
	expect(
		queryClient.getQueryData(["identity", testUser.id, "messenger"]),
	).toBeUndefined();

	const messenger = useMessengerStore.getState();
	expect(messenger.channels).toEqual([]);
	expect(messenger.unreadCounts).toEqual({});
	expect(messenger.activeChannelId).toBeNull();
	expect(messenger.activeThreadMessageId).toBeNull();
	expect(messenger.replyToMessage).toBeNull();
	expect(messenger.typingUsers).toEqual({});

	const ws = useWsStore.getState();
	expect(close).toHaveBeenCalledTimes(1);
	expect(ws.connected).toBe(false);
	expect(ws.socket).toBeNull();
	expect(ws.listeners.size).toBe(0);
	const auth = useAuthStore.getState();
	expect(auth.accessToken).toBeNull();
	expect(auth.refreshToken).toBeNull();
	expect(auth.user).toBeNull();
	expect(auth.isAuthenticated).toBe(false);
}

describe("auth-store", () => {
	beforeEach(() => {
		resetAuthStore();
		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		resetIdentityDerivedState();
		vi.restoreAllMocks();
	});

	afterEach(() => {
		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		vi.useRealTimers();
		vi.unstubAllGlobals();
		vi.restoreAllMocks();
	});

	it("setTokens stores tokens and sets isAuthenticated", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");

		const state = useAuthStore.getState();
		expect(state.accessToken).toBe("access-abc");
		expect(state.refreshToken).toBe("refresh-xyz");
		expect(state.isAuthenticated).toBe(true);
	});

	it("refreshing tokens keeps existing identity-derived cache, messenger and WS state", () => {
		const { close, listener, socket } = seedIdentityDerivedState();
		useAuthStore.setState({
			accessToken: "access-before-refresh",
			refreshToken: "refresh-before-refresh",
			user: testUser,
			isAuthenticated: true,
		});

		useAuthStore.getState().setTokens("access-refreshed", "refresh-refreshed");

		expect(useAuthStore.getState()).toMatchObject({
			accessToken: "access-refreshed",
			refreshToken: "refresh-refreshed",
			user: testUser,
			isAuthenticated: true,
		});
		expect(queryClient.getQueryData(CHANNELS_QUERY_KEY)).toEqual([testChannel]);
		expect(
			queryClient.getQueryData(["identity", testUser.id, "messenger"]),
		).toEqual({ private: true });

		const messenger = useMessengerStore.getState();
		expect(messenger.channels).toEqual([testChannel]);
		expect(messenger.unreadCounts).toEqual({ [testChannel.id]: 4 });
		expect(messenger.activeChannelId).toBe(testChannel.id);
		expect(messenger.activeThreadMessageId).toBe(testMessage.id);
		expect(messenger.replyToMessage).toEqual(testMessage);
		expect(messenger.typingUsers).toMatchObject({
			[testChannel.id]: [{ userId: "user-2", displayName: "Teammate" }],
		});

		const ws = useWsStore.getState();
		expect(ws.connected).toBe(true);
		expect(ws.socket).toBe(socket);
		expect(ws.listeners).toEqual(new Set([listener]));
		expect(close).not.toHaveBeenCalled();
	});

	it("does not persist JWT into sessionStorage (A2W11-2-001)", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");
		useAuthStore.getState().setUser({
			id: "user-1",
			email: "dev@hub.local",
			name: "Dev",
			display_name: "Dev",
			timezone: "UTC",
			language: "ru",
			system_role: "member",
			onboarding_completed: true,
		});

		expect(sessionStorage.getItem("hub-auth")).toBeNull();
		const raw = sessionStorage.getItem("hub-auth");
		expect(raw).toBeNull();
		// Defence: nowhere in sessionStorage should JWT substrings appear.
		for (let i = 0; i < sessionStorage.length; i++) {
			const key = sessionStorage.key(i);
			if (!key) continue;
			const val = sessionStorage.getItem(key) ?? "";
			expect(val).not.toContain("access-abc");
			expect(val).not.toContain("refresh-xyz");
		}
	});

	it("setUser stores user profile", () => {
		const user = {
			id: "user-1",
			email: "dev@hub.local",
			name: "Dev",
			display_name: "Dev",
			timezone: "UTC",
			language: "ru",
			system_role: "member",
			onboarding_completed: true,
		};

		useAuthStore.getState().setUser(user);

		expect(useAuthStore.getState().user).toEqual(user);
	});

	it("logout clears tokens and isAuthenticated", () => {
		useAuthStore.getState().setTokens("access-abc", "refresh-xyz");
		useAuthStore.getState().setUser({
			id: "user-1",
			email: "dev@hub.local",
			name: "Dev",
			display_name: "Dev",
			timezone: "UTC",
			language: "ru",
			system_role: "admin",
			onboarding_completed: true,
		});

		useAuthStore.getState().logout();

		const state = useAuthStore.getState();
		expect(state.accessToken).toBeNull();
		expect(state.refreshToken).toBeNull();
		expect(state.user).toBeNull();
		expect(state.isAuthenticated).toBe(false);
	});

	it("clearSessionLocal clears without BroadcastChannel postMessage (P2-FE-09)", () => {
		const postSpy = vi.spyOn(BroadcastChannel.prototype, "postMessage");

		useAuthStore.setState({
			accessToken: "access-remote",
			refreshToken: "refresh-remote",
			user: null,
			isAuthenticated: true,
		});
		postSpy.mockClear();

		useAuthStore.getState().clearSessionLocal();

		const state = useAuthStore.getState();
		expect(state.isAuthenticated).toBe(false);
		expect(state.accessToken).toBeNull();
		expect(state.refreshToken).toBeNull();
		expect(postSpy).not.toHaveBeenCalled();

		useAuthStore.setState({
			accessToken: "a",
			refreshToken: "r",
			isAuthenticated: true,
		});
		postSpy.mockClear();
		useAuthStore.getState().logout();
		expect(postSpy).toHaveBeenCalled();
	});

	it("logout synchronously clears all identity-derived client state and broadcasts", () => {
		const postSpy = vi.spyOn(BroadcastChannel.prototype, "postMessage");
		const { close } = seedIdentityDerivedState();
		useAuthStore.setState({
			accessToken: "access-logout",
			refreshToken: "refresh-logout",
			user: testUser,
			isAuthenticated: true,
		});
		postSpy.mockClear();

		useAuthStore.getState().logout();

		expectIdentityDerivedStateCleared(close);
		expect(postSpy).toHaveBeenCalledWith({ type: "logout" });
	});

	it("clearSessionLocal synchronously clears identity-derived client state without broadcasting", () => {
		const postSpy = vi.spyOn(BroadcastChannel.prototype, "postMessage");
		const { close } = seedIdentityDerivedState();
		useAuthStore.setState({
			accessToken: "access-remote",
			refreshToken: "refresh-remote",
			user: testUser,
			isAuthenticated: true,
		});
		postSpy.mockClear();

		useAuthStore.getState().clearSessionLocal();

		expectIdentityDerivedStateCleared(close);
		expect(postSpy).not.toHaveBeenCalled();

		useAuthStore.getState().clearSessionLocal();
		expectIdentityDerivedStateCleared(close);
		expect(postSpy).not.toHaveBeenCalled();
	});

	it("logout resets identity-derived state before broadcasting", () => {
		const postSpy = vi.spyOn(BroadcastChannel.prototype, "postMessage");
		const { close } = seedIdentityDerivedState();
		let stateAtBroadcast: Record<string, unknown> | undefined;
		useAuthStore.setState({
			accessToken: "access-logout",
			refreshToken: "refresh-logout",
			user: testUser,
			isAuthenticated: true,
		});
		postSpy.mockClear();
		postSpy.mockImplementation(() => {
			const messenger = useMessengerStore.getState();
			const ws = useWsStore.getState();
			const auth = useAuthStore.getState();
			stateAtBroadcast = {
				channelsQuery: queryClient.getQueryData(CHANNELS_QUERY_KEY),
				identityQuery: queryClient.getQueryData([
					"identity",
					testUser.id,
					"messenger",
				]),
				channels: messenger.channels,
				unreadCounts: messenger.unreadCounts,
				activeChannelId: messenger.activeChannelId,
				activeThreadMessageId: messenger.activeThreadMessageId,
				replyToMessage: messenger.replyToMessage,
				typingUsers: messenger.typingUsers,
				connected: ws.connected,
				socket: ws.socket,
				listenerCount: ws.listeners.size,
				accessToken: auth.accessToken,
				refreshToken: auth.refreshToken,
				user: auth.user,
				isAuthenticated: auth.isAuthenticated,
			};
		});

		useAuthStore.getState().logout();

		expectIdentityDerivedStateCleared(close);
		expect(stateAtBroadcast).toEqual({
			channelsQuery: undefined,
			identityQuery: undefined,
			channels: [],
			unreadCounts: {},
			activeChannelId: null,
			activeThreadMessageId: null,
			replyToMessage: null,
			typingUsers: {},
			connected: false,
			socket: null,
			listenerCount: 0,
			accessToken: null,
			refreshToken: null,
			user: null,
			isAuthenticated: false,
		});
		expect(postSpy).toHaveBeenCalledWith({ type: "logout" });
	});

	it.each(["logout", "clearSessionLocal"] as const)(
		"%s cancels a pending public reconnect before it can open another socket",
		async (sessionResetAction) => {
			vi.useFakeTimers();
			FakeWebSocket.instances = [];
			vi.stubGlobal("WebSocket", FakeWebSocket);
			vi.stubGlobal(
				"fetch",
				vi.fn(async () =>
					Response.json({ ticket: "ticket-auth-reset", expires_in: 30 }),
				),
			);
			useAuthStore.setState({
				accessToken: "access-reconnect",
				refreshToken: "refresh-reconnect",
				user: testUser,
				isAuthenticated: true,
			});

			useWsStore.getState().connect("access-reconnect");
			await vi.runAllTimersAsync();
			expect(FakeWebSocket.instances).toHaveLength(1);
			FakeWebSocket.instances[0]?.open();
			FakeWebSocket.instances[0]?.drop();

			useAuthStore.getState()[sessionResetAction]();
			await vi.advanceTimersByTimeAsync(60_000);

			expect(FakeWebSocket.instances).toHaveLength(1);
			expect(useWsStore.getState()).toMatchObject({
				connected: false,
				socket: null,
			});
		},
	);
});
