import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { clearWsReconnectTimer, useWsStore } from "@/stores/ws-store";

type Handler = ((event?: unknown) => void) | null;

type Deferred<T> = {
	promise: Promise<T>;
	resolve: (value: T) => void;
	reject: (reason?: unknown) => void;
};

function deferred<T>(): Deferred<T> {
	let resolve!: (value: T) => void;
	let reject!: (reason?: unknown) => void;
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise;
		reject = rejectPromise;
	});
	return { promise, resolve, reject };
}

class FakeWebSocket {
	static CONNECTING = 0;
	static OPEN = 1;
	static CLOSING = 2;
	static CLOSED = 3;
	static instances: FakeWebSocket[] = [];

	readyState = FakeWebSocket.CONNECTING;
	onopen: Handler = null;
	onclose: Handler = null;
	onerror: Handler = null;
	onmessage: Handler = null;

	constructor(
		readonly url: string,
		readonly protocols?: string | string[],
	) {
		FakeWebSocket.instances.push(this);
	}

	close() {
		this.readyState = FakeWebSocket.CLOSED;
		this.onclose?.({});
	}

	open() {
		this.readyState = FakeWebSocket.OPEN;
		this.onopen?.({});
	}

	send() {}
}

async function flushMicrotasks(): Promise<void> {
	for (let attempt = 0; attempt < 5; attempt += 1) {
		await Promise.resolve();
	}
}

describe("ws-store connect generation", () => {
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

	it("does not create or overwrite the current socket when a superseded ticket resolves", async () => {
		const ticketA = deferred<Response>();
		const ticketB = deferred<Response>();
		vi.stubGlobal(
			"fetch",
			vi.fn((_: RequestInfo | URL, init?: RequestInit) => {
				const authorization = new Headers(init?.headers).get("Authorization");
				return authorization === "Bearer token-a"
					? ticketA.promise
					: ticketB.promise;
			}),
		);

		useWsStore.getState().connect("token-a");
		useWsStore.getState().connect("token-b");

		ticketB.resolve(Response.json({ ticket: "ticket-b" }));
		await flushMicrotasks();

		expect(FakeWebSocket.instances).toHaveLength(1);
		const currentSocket = FakeWebSocket.instances[0];
		currentSocket?.open();
		expect(useWsStore.getState()).toMatchObject({
			connected: true,
			socket: currentSocket,
		});

		ticketA.resolve(Response.json({ ticket: "ticket-a" }));
		await flushMicrotasks();
		await vi.advanceTimersByTimeAsync(60_000);

		expect(FakeWebSocket.instances).toEqual([currentSocket]);
		expect(useWsStore.getState()).toMatchObject({
			connected: true,
			socket: currentSocket,
		});
	});

	it("does not schedule a reconnect when a pending ticket fails after disconnect", async () => {
		const ticket = deferred<Response>();
		const fetchSpy = vi.fn(() => ticket.promise);
		vi.stubGlobal("fetch", fetchSpy);

		useWsStore.getState().connect("token-a");
		useWsStore.getState().disconnect();
		ticket.reject(new Error("ticket request cancelled"));
		await flushMicrotasks();
		await vi.advanceTimersByTimeAsync(60_000);

		expect(fetchSpy).toHaveBeenCalledOnce();
		expect(FakeWebSocket.instances).toHaveLength(0);
		expect(useWsStore.getState()).toMatchObject({
			connected: false,
			socket: null,
		});
	});
});
