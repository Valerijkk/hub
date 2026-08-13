import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel } from "@/api/messenger";
import { CHANNELS_QUERY_KEY, useChannels } from "@/hooks/use-channels";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
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

describe("useChannels WS → Query cache (A10-001)", () => {
	let queryClient: QueryClient;
	let setQueryDataSpy: ReturnType<typeof vi.spyOn>;
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
		useMessengerStore.setState({ channels: [], unreadCounts: {} });

		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		setQueryDataSpy = vi.spyOn(queryClient, "setQueryData");

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

	async function mountUseChannels() {
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

		const hook = renderHook(() => useChannels(), { wrapper });

		await waitFor(() => {
			expect(hook.result.current.isSuccess).toBe(true);
			expect(hook.result.current.data?.map((c) => c.id)).toEqual([
				"ch-existing",
			]);
		});
		// Listener регистрируется в useEffect — дождёмся подписки.
		await waitFor(() => {
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0);
		});

		return hook;
	}

	it("channel_created WS path calls setQueryData and prepends channel", async () => {
		const { result } = await mountUseChannels();
		setQueryDataSpy.mockClear();

		const created = makeChannel({ id: "ch-new", name: "New Room" });
		emitWsEvent("channel_created", created);

		expect(setQueryDataSpy).toHaveBeenCalledWith(
			CHANNELS_QUERY_KEY,
			expect.any(Function),
		);

		await waitFor(() => {
			expect(
				queryClient.getQueryData<Channel[]>(CHANNELS_QUERY_KEY)?.map((c) => c.id),
			).toEqual(["ch-new", "ch-existing"]);
			expect(result.current.data?.map((c) => c.id)).toEqual([
				"ch-new",
				"ch-existing",
			]);
		});
	});

	it("channel_updated WS path calls setQueryData and patches name", async () => {
		const { result } = await mountUseChannels();
		setQueryDataSpy.mockClear();

		emitWsEvent("channel_updated", {
			id: "ch-existing",
			name: "Renamed",
			updated_at: "2026-07-28T00:00:00Z",
		});

		expect(setQueryDataSpy).toHaveBeenCalledWith(
			CHANNELS_QUERY_KEY,
			expect.any(Function),
		);

		await waitFor(() => {
			expect(
				queryClient.getQueryData<Channel[]>(CHANNELS_QUERY_KEY)?.[0]?.name,
			).toBe("Renamed");
			expect(result.current.data?.[0]?.name).toBe("Renamed");
		});
	});
});
