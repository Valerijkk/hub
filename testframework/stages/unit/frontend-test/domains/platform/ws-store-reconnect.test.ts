import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
	clearWsReconnectTimer,
	useWsStore,
	wsReconnectDelayMs,
} from "@/stores/ws-store";

type Handler = ((ev?: unknown) => void) | null;

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

	/** Симулирует успешное открытие. */
	open() {
		this.readyState = FakeWebSocket.OPEN;
		this.onopen?.({});
	}

	/** Симулирует неожиданный обрыв. */
	drop() {
		this.readyState = FakeWebSocket.CLOSED;
		this.onclose?.({});
	}
}

describe("ws-store reconnect (A10-003)", () => {
	beforeEach(() => {
		vi.useFakeTimers();
		FakeWebSocket.instances = [];
		vi.stubGlobal("WebSocket", FakeWebSocket);
		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
	});

	afterEach(() => {
		useWsStore.getState().disconnect();
		clearWsReconnectTimer();
		vi.useRealTimers();
		vi.unstubAllGlobals();
	});

	it("wsReconnectDelayMs grows with ceiling", () => {
		expect(wsReconnectDelayMs(0)).toBe(1_000);
		expect(wsReconnectDelayMs(1)).toBe(2_000);
		expect(wsReconnectDelayMs(5)).toBe(30_000);
		expect(wsReconnectDelayMs(10)).toBe(30_000);
	});

	it("schedules reconnect after unexpected close", () => {
		useWsStore.getState().connect("tok-1");
		expect(FakeWebSocket.instances).toHaveLength(1);
		FakeWebSocket.instances[0]?.open();
		expect(useWsStore.getState().connected).toBe(true);

		FakeWebSocket.instances[0]?.drop();
		expect(useWsStore.getState().connected).toBe(false);
		expect(useWsStore.getState().socket).toBeNull();

		vi.advanceTimersByTime(1_000);
		expect(FakeWebSocket.instances).toHaveLength(2);
		expect(FakeWebSocket.instances[1]?.url).toContain("tok-1");
	});

	it("does not reconnect after intentional disconnect", () => {
		useWsStore.getState().connect("tok-2");
		FakeWebSocket.instances[0]?.open();
		useWsStore.getState().disconnect();

		vi.advanceTimersByTime(60_000);
		expect(FakeWebSocket.instances).toHaveLength(1);
		expect(useWsStore.getState().connected).toBe(false);
	});

	it("clears pending reconnect timer on disconnect (no leak)", () => {
		useWsStore.getState().connect("tok-3");
		FakeWebSocket.instances[0]?.open();
		FakeWebSocket.instances[0]?.drop();

		useWsStore.getState().disconnect();
		vi.advanceTimersByTime(60_000);
		expect(FakeWebSocket.instances).toHaveLength(1);
	});
});
