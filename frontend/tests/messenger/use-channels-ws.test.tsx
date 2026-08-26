import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { StrictMode, type ReactNode } from "react";
import {
	afterEach,
	beforeEach,
	describe,
	expect,
	it,
	type MockInstance,
	vi,
} from "vitest";
import type { Channel } from "@/api/messenger";
import {
	channelsQueryKey,
	useChannels,
} from "@/hooks/use-channels";
import { useMessengerChannelsRealtime } from "@/hooks/use-messenger-channels-realtime";
import { unreadQueryKey } from "@/hooks/use-unread";
import { useAuthStore } from "@/stores/auth-store";
import {
	emptyUnreadSnapshot,
	useMessengerStore,
} from "@/stores/messenger-store";
import { clearWsReconnectTimer, useWsStore } from "@/stores/ws-store";

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannels: vi.fn(),
	};
});

import { getChannels } from "@/api/messenger";

type Handler = ((ev?: unknown) => void) | null;

/** FakeWS — тот же паттерн, что A10-003 (`ws-store-reconnect.test.ts`). */
class FakeWebSocket {
	static CONNECTING = 0;
	static OPEN = 1;
	static CLOSING = 2;
	static CLOSED = 3;
	static instances: FakeWebSocket[] = [];

	readyState = FakeWebSocket.CONNECTING;
	url: string;
	onopen: Handler = null;
	onclose: Handler = null;
	onerror: Handler = null;
	onmessage: Handler = null;
	sent: string[] = [];

	constructor(url: string) {
		this.url = url;
		FakeWebSocket.instances.push(this);
	}

	send(data: string) {
		this.sent.push(data);
	}

	close() {
		this.readyState = FakeWebSocket.CLOSED;
		this.onclose?.({});
	}

	open() {
		this.readyState = FakeWebSocket.OPEN;
		this.onopen?.({});
	}
}

function makeChannel(
	partial: Partial<Channel> & Pick<Channel, "id" | "name">,
): Channel {
	return {
		channel_type: "public",
		is_readonly: false,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-07-01T00:00:00Z",
		updated_at: "2026-07-01T00:00:00Z",
		...partial,
	};
}

function emitWsEvent(event: string, payload: unknown) {
	const ws = FakeWebSocket.instances[0];
	expect(ws?.onmessage).toBeTypeOf("function");
	act(() => {
		ws?.onmessage?.({
			data: JSON.stringify({ type: "event", event, payload }),
		});
	});
}

describe("messenger channel realtime ownership (P149)", () => {
	let queryClient: QueryClient;
	let setQueryDataSpy: MockInstance<QueryClient["setQueryData"]>;
	let invalidateQueriesSpy: MockInstance<QueryClient["invalidateQueries"]>;
	let cancelQueriesSpy: MockInstance<QueryClient["cancelQueries"]>;
	let removeQueriesSpy: MockInstance<QueryClient["removeQueries"]>;
	const existing = makeChannel({ id: "ch-existing", name: "Existing" });

	beforeEach(() => {
		FakeWebSocket.instances = [];
		vi.stubGlobal("WebSocket", FakeWebSocket);
		vi.stubGlobal(
			"fetch",
			vi.fn(async () =>
				Response.json({ ticket: "ticket-channels", expires_in: 30 }),
			),
		);

		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		useWsStore.setState({
			connected: false,
			socket: null,
			listeners: new Set(),
		});

		sessionStorage.clear();
		useAuthStore.setState({
			accessToken: "tok-a10-001",
			refreshToken: "ref-a10-001",
			isAuthenticated: true,
			user: {
				id: "user-1",
				email: "u@hub.local",
				timezone: "UTC",
				language: "en",
				system_role: "member",
				onboarding_completed: true,
			},
		});
		useMessengerStore.setState({ channels: [], unreadSnapshot: emptyUnreadSnapshot() });

		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		setQueryDataSpy = vi.spyOn(queryClient, "setQueryData");
		invalidateQueriesSpy = vi.spyOn(queryClient, "invalidateQueries");
		cancelQueriesSpy = vi.spyOn(queryClient, "cancelQueries");
		removeQueriesSpy = vi.spyOn(queryClient, "removeQueries");

		vi.mocked(getChannels).mockResolvedValue([existing]);
	});

	afterEach(() => {
		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		queryClient.clear();
		vi.clearAllMocks();
		vi.unstubAllGlobals();
	});

	function wrapper({ children }: { children: ReactNode }) {
		return (
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		);
	}

	function strictWrapper({ children }: { children: ReactNode }) {
		return (
			<StrictMode>
				<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
			</StrictMode>
		);
	}

	async function mountRealtimeOwner({ strict = false } = {}) {
		await act(async () => {
			useWsStore.getState().connect("tok-a10-001");
			await Promise.resolve();
			await Promise.resolve();
		});
		await waitFor(() => {
			expect(FakeWebSocket.instances.length).toBeGreaterThan(0);
		});
		act(() => {
			FakeWebSocket.instances[0]?.open();
		});

		const hook = renderHook(
			() => {
				const query = useChannels();
				useMessengerChannelsRealtime();
				return query;
			},
			{ wrapper: strict ? strictWrapper : wrapper },
		);

		await waitFor(() => {
			expect(hook.result.current.isSuccess).toBe(true);
			expect(hook.result.current.data?.map((c) => c.id)).toEqual([
				"ch-existing",
			]);
		});
		// Listener регистрируется в useEffect — дождёмся подписки.
		await waitFor(() => {
			expect(useWsStore.getState().listeners.size).toBe(1);
		});

		return hook;
	}

	it("keeps a shared useChannels observer detached from websocket ownership", async () => {
		await act(async () => {
			useWsStore.getState().connect("tok-a10-001");
			await Promise.resolve();
			await Promise.resolve();
		});
		await waitFor(() => {
			expect(FakeWebSocket.instances.length).toBeGreaterThan(0);
		});
		act(() => {
			FakeWebSocket.instances[0]?.open();
		});
		const observer = renderHook(() => useChannels(), { wrapper });

		await waitFor(() => {
			expect(observer.result.current.isSuccess).toBe(true);
		});

		expect(FakeWebSocket.instances[0]?.sent).toEqual([]);
		expect(useWsStore.getState().listeners.size).toBe(0);
	});

	it("subscribes only to the exact authenticated user room without an initial refetch", async () => {
		await mountRealtimeOwner();

		expect(FakeWebSocket.instances[0]?.sent).toEqual([
			JSON.stringify({ type: "subscribe", room: "user:user-1" }),
		]);
		expect(FakeWebSocket.instances[0]?.sent.join(" ")).not.toContain(
			"user:user-1:channels",
		);
		expect(invalidateQueriesSpy).not.toHaveBeenCalled();
		expect(getChannels).toHaveBeenCalledTimes(1);
	});

	it("does not invalidate the initial connected room during StrictMode setup cleanup setup", async () => {
		await mountRealtimeOwner({ strict: true });

		const sent = FakeWebSocket.instances[0]?.sent ?? [];
		expect(sent).toContain(
			JSON.stringify({ type: "subscribe", room: "user:user-1" }),
		);
		expect(sent).not.toContain(
			JSON.stringify({ type: "subscribe", room: "user:user-1:channels" }),
		);
		expect(invalidateQueriesSpy).not.toHaveBeenCalled();
	});

	it("clears the previous identity before subscribing the same socket to the next user", async () => {
		await mountRealtimeOwner();
		const socket = FakeWebSocket.instances[0];
		expect(socket).toBeDefined();
		const sendSpy = vi.spyOn(socket as FakeWebSocket, "send");
		invalidateQueriesSpy.mockClear();
		cancelQueriesSpy.mockClear();
		removeQueriesSpy.mockClear();
		vi.mocked(getChannels).mockClear();
		const channelB = makeChannel({ id: "ch-user-b", name: "User B channel" });
		vi.mocked(getChannels).mockResolvedValueOnce([channelB]);

		const userA = useAuthStore.getState().user;
		if (!userA) throw new Error("authenticated user A is required");
		act(() => {
			useAuthStore.setState({ user: { ...userA, id: "user-2" } });
		});

		await waitFor(() => {
			expect(sendSpy).toHaveBeenCalledTimes(2);
		});
		expect(sendSpy).toHaveBeenNthCalledWith(
			1,
			JSON.stringify({ type: "unsubscribe", room: "user:user-1" }),
		);
		expect(sendSpy).toHaveBeenNthCalledWith(
			2,
			JSON.stringify({ type: "subscribe", room: "user:user-2" }),
		);
		expect(invalidateQueriesSpy).not.toHaveBeenCalled();

		const cancelOrder = cancelQueriesSpy.mock.invocationCallOrder[0];
		const removeOrder = removeQueriesSpy.mock.invocationCallOrder[0];
		const subscribeBOrder = sendSpy.mock.invocationCallOrder[1];
		expect(cancelOrder).toBeLessThan(removeOrder ?? Number.POSITIVE_INFINITY);
		expect(removeOrder).toBeLessThan(subscribeBOrder ?? Number.POSITIVE_INFINITY);
		expect(useMessengerStore.getState().activeChannelId).toBeNull();
		expect(useWsStore.getState().socket).toBe(socket as unknown as WebSocket);
		expect(FakeWebSocket.instances).toHaveLength(1);
		await waitFor(() => {
			expect(getChannels).toHaveBeenCalledTimes(1);
			expect(useMessengerStore.getState().channels).toEqual([channelB]);
		});
	});

	it.each([
		["channel_created", { channel_id: "ch-created" }],
		["channel_updated", { channel_id: "ch-existing" }],
		["channel_removed", { channel_id: "ch-removed" }],
		[
			"new_message_notification",
			{ channel_id: "ch-existing", message_id: "msg-1" },
		],
	] as const)(
		"%s invalidates and refetches the authoritative channel list without local patches",
		async (event, payload) => {
			await mountRealtimeOwner();
			setQueryDataSpy.mockClear();
			invalidateQueriesSpy.mockClear();
			vi.mocked(getChannels).mockClear();

			emitWsEvent(event, payload);

			expect(invalidateQueriesSpy).toHaveBeenCalledTimes(
				event === "channel_removed" || event === "new_message_notification" ? 2 : 1,
			);
			expect(invalidateQueriesSpy).toHaveBeenCalledWith({
				queryKey: channelsQueryKey("user-1"),
				exact: true,
			});
			expect(setQueryDataSpy).not.toHaveBeenCalled();
			expect(useMessengerStore.getState().unreadSnapshot).toEqual(emptyUnreadSnapshot());

			await waitFor(() => {
				expect(getChannels).toHaveBeenCalledTimes(1);
			});
			expect(
				queryClient.getQueryData<Channel[]>(channelsQueryKey("user-1"))?.map((channel) =>
					channel.id,
				),
			).toEqual(["ch-existing"]);
		},
	);

	it.each([
		["channel_created", { channel_id: "ch-authoritative" }],
		["channel_updated", { channel_id: "ch-stale" }],
		["channel_removed", { channel_id: "ch-stale" }],
		[
			"new_message_notification",
			{ channel_id: "ch-stale", message_id: "msg-deferred" },
		],
	] as const)(
		"%s uses immediate revoke only for removal, then accepts authoritative refetch",
		async (event, payload) => {
			await mountRealtimeOwner();
			setQueryDataSpy.mockClear();
			invalidateQueriesSpy.mockClear();
			vi.mocked(getChannels).mockClear();
			const stale = makeChannel({ id: "ch-stale", name: "Stale local state" });
			const authoritative = makeChannel({
				id: "ch-authoritative",
				name: "Authoritative server state",
			});
			useMessengerStore.setState({ channels: [stale] });

			let resolveRefetch: (channels: Channel[]) => void = () => undefined;
			const refetch = new Promise<Channel[]>((resolve) => {
				resolveRefetch = resolve;
			});
			vi.mocked(getChannels).mockImplementationOnce(() => refetch);

			emitWsEvent(event, payload);

			expect(invalidateQueriesSpy).toHaveBeenCalledTimes(
				event === "channel_removed" || event === "new_message_notification" ? 2 : 1,
			);
			expect(invalidateQueriesSpy).toHaveBeenCalledWith({
				queryKey: channelsQueryKey("user-1"),
				exact: true,
			});
			await waitFor(() => {
				expect(getChannels).toHaveBeenCalledTimes(1);
			});
			expect(useMessengerStore.getState().channels).toEqual(
				event === "channel_removed" ? [] : [stale],
			);
			expect(setQueryDataSpy).not.toHaveBeenCalled();

			await act(async () => {
				resolveRefetch([authoritative]);
				await refetch;
			});

			await waitFor(() => {
				expect(useMessengerStore.getState().channels).toEqual([authoritative]);
			});
		},
	);

	it("handles duplicate notifications through the authoritative refetch path without unread arithmetic", async () => {
		await mountRealtimeOwner();
		setQueryDataSpy.mockClear();
		invalidateQueriesSpy.mockClear();
		vi.mocked(getChannels).mockClear();

		emitWsEvent("new_message_notification", {
			channel_id: "ch-existing",
			message_id: "msg-duplicate",
		});
		emitWsEvent("new_message_notification", {
			channel_id: "ch-existing",
			message_id: "msg-duplicate",
		});

		expect(invalidateQueriesSpy).toHaveBeenCalledTimes(4);
		expect(invalidateQueriesSpy).toHaveBeenNthCalledWith(1, {
			queryKey: channelsQueryKey("user-1"),
			exact: true,
		});
		expect(invalidateQueriesSpy).toHaveBeenNthCalledWith(2, {
			queryKey: unreadQueryKey("user-1"),
			exact: true,
		});
		expect(invalidateQueriesSpy).toHaveBeenNthCalledWith(3, {
			queryKey: channelsQueryKey("user-1"),
			exact: true,
		});
		expect(invalidateQueriesSpy).toHaveBeenNthCalledWith(4, {
			queryKey: unreadQueryKey("user-1"),
			exact: true,
		});
		expect(setQueryDataSpy).not.toHaveBeenCalled();
		expect(useMessengerStore.getState().unreadSnapshot).toEqual(emptyUnreadSnapshot());

		await waitFor(() => {
			expect(getChannels).toHaveBeenCalledTimes(2);
		});
	});

	it("resubscribes before the exact-key reconnect invalidation after false to true", async () => {
		await mountRealtimeOwner();
		invalidateQueriesSpy.mockClear();

		act(() => {
			useWsStore.setState({ connected: false, socket: null });
		});

		const reconnected = new FakeWebSocket("ws://hub.test/ws");
		reconnected.readyState = FakeWebSocket.OPEN;
		const sendSpy = vi.spyOn(reconnected, "send");

		act(() => {
			useWsStore.setState({
				connected: true,
				socket: reconnected as unknown as WebSocket,
			});
		});

		await waitFor(() => {
			expect(sendSpy).toHaveBeenCalledWith(
				JSON.stringify({ type: "subscribe", room: "user:user-1" }),
			);
			expect(invalidateQueriesSpy).toHaveBeenCalledWith(
				{
					queryKey: channelsQueryKey("user-1"),
					exact: true,
				},
				{ cancelRefetch: false },
			);
		});

		const subscribeOrder = sendSpy.mock.invocationCallOrder[0];
		const invalidateOrder = invalidateQueriesSpy.mock.invocationCallOrder[0];
		expect(subscribeOrder).toBeLessThan(invalidateOrder ?? Number.POSITIVE_INFINITY);
	});

	it("does not let an extra useChannels observer unsubscribe the realtime owner", async () => {
		const owner = await mountRealtimeOwner();
		const socket = FakeWebSocket.instances[0];
		if (!socket) throw new Error("an open websocket is required");
		const sendSpy = vi.spyOn(socket, "send");
		sendSpy.mockClear();

		const extraObserver = renderHook(() => useChannels(), { wrapper });
		await waitFor(() => {
			expect(extraObserver.result.current.isSuccess).toBe(true);
		});
		extraObserver.unmount();

		expect(sendSpy).not.toHaveBeenCalledWith(
			JSON.stringify({ type: "unsubscribe", room: "user:user-1" }),
		);
		expect(useWsStore.getState().listeners.size).toBe(1);

		owner.unmount();
		expect(sendSpy).toHaveBeenCalledTimes(1);
		expect(sendSpy).toHaveBeenCalledWith(
			JSON.stringify({ type: "unsubscribe", room: "user:user-1" }),
		);
		expect(useWsStore.getState().listeners.size).toBe(0);
	});

	it("unsubscribes from the exact room and removes its event handler only when the realtime owner unmounts", async () => {
		const hook = await mountRealtimeOwner();
		setQueryDataSpy.mockClear();
		invalidateQueriesSpy.mockClear();

		hook.unmount();

		expect(FakeWebSocket.instances[0]?.sent).toContain(
			JSON.stringify({ type: "unsubscribe", room: "user:user-1" }),
		);
		expect(useWsStore.getState().listeners.size).toBe(0);

		emitWsEvent("channel_updated", { channel_id: "ch-existing" });
		expect(invalidateQueriesSpy).not.toHaveBeenCalled();
		expect(setQueryDataSpy).not.toHaveBeenCalled();
	});
});
