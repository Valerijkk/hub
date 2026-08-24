import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message } from "@/api/messenger";
import { useMessages, useThreadMessages } from "@/hooks/use-messages";
import { useWsStore } from "@/stores/ws-store";

const api = vi.hoisted(() => ({
	getMessages: vi.fn(),
	getThreadReplies: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getMessages: api.getMessages,
		getThreadReplies: api.getThreadReplies,
	};
});

type MessageCache = {
	pages: Message[][];
	pageParams: unknown[];
};

const channelId = "channel-1";
const threadId = "root-1";
const rootQueryKey = ["messenger", "messages", channelId] as const;

function threadQueryKey(id: string) {
	return ["messenger", "thread", channelId, id] as const;
}

function cacheMessage(id: string, contentText: string, overrides: Partial<Message> = {}): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "sender-1",
		content_text: contentText,
		content_json: { type: "doc", text: contentText },
		reply_count: 0,
		created_at: "2026-08-25T10:00:00Z",
		...overrides,
	};
}

describe("useMessages canonical cache events", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		api.getMessages.mockResolvedValue([]);
		api.getThreadReplies.mockResolvedValue([]);
		useWsStore.setState({
			connected: false,
			socket: null,
			listeners: new Set(),
		});
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	function seed(key: readonly unknown[], pages: Message[][]): void {
		queryClient.setQueryData<MessageCache>(key, {
			pages,
			pageParams: pages.map(() => undefined),
		});
	}

	function cached(key: readonly unknown[]): MessageCache | undefined {
		return queryClient.getQueryData<MessageCache>(key);
	}

	function publish(event: string, payload: unknown): void {
		act(() => {
			for (const listener of [...useWsStore.getState().listeners]) {
				listener(event, payload);
			}
		});
	}

	async function mountRoot() {
		const hook = renderHook(() => useMessages(channelId), { wrapper: TestProvider });
		await waitFor(() => expect(hook.result.current.isSuccess).toBe(true));
		await waitFor(() => expect(useWsStore.getState().listeners.size).toBeGreaterThan(0));
		return hook;
	}

	async function mountThread() {
		const hook = renderHook(() => useThreadMessages(channelId, threadId), {
			wrapper: TestProvider,
		});
		await waitFor(() => expect(hook.result.current.isSuccess).toBe(true));
		await waitFor(() => expect(useWsStore.getState().listeners.size).toBeGreaterThan(0));
		return hook;
	}

	function recordingSocket(order: string[]): WebSocket {
		return {
			readyState: 1,
			send(raw: string) {
				const frame = JSON.parse(raw) as { type?: string; room?: string };
				if (
					(frame.type === "subscribe" || frame.type === "unsubscribe") &&
					frame.room
				) {
					order.push(`${frame.type}:${frame.room}`);
				}
			},
		} as unknown as WebSocket;
	}

	function recordInvalidations(order: string[]) {
		const originalInvalidate = queryClient.invalidateQueries.bind(queryClient);
		return vi
			.spyOn(queryClient, "invalidateQueries")
			.mockImplementation((filters, options) => {
				order.push(`invalidate:${JSON.stringify(filters?.queryKey ?? [])}`);
				return originalInvalidate(filters, options);
			});
	}

	it("uses only a nested canonical root message, dedupes its durable ID, and rejects malformed, legacy, foreign, and reply routes", async () => {
		await mountRoot();
		const existing = cacheMessage("root-existing", "already cached");
		seed(rootQueryKey, [[existing]]);
		const before = cached(rootQueryKey);

		const canonicalRoot: Message = {
			id: "root-rich-1",
			channel_id: channelId,
			sender_id: "sender-rich",
			content_json: {
				type: "doc",
				children: [{ type: "paragraph", text: "canonical rich root" }],
			},
			content_text: "canonical rich root",
			voice_url: "https://files.example.test/voice.ogg",
			voice_duration_ms: 812,
			attachments: [
				{
					file_id: "file-1",
					file_name: "report.pdf",
					file_size: 42,
					mime_type: "application/pdf",
					url: "https://files.example.test/report.pdf",
				},
			],
			reactions: [{ emoji: "thumbsup", user_ids: ["user-a"] }],
			mentions: ["user-a"],
			mention_groups: ["all"],
			client_nonce: "4f540250-335c-4c48-92ad-7de744a70d11",
			reply_count: 0,
			created_at: "2026-08-25T10:01:00Z",
		};

		publish("message_sent", { message: canonicalRoot });
		publish("message_sent", { message: canonicalRoot });
		publish("message_sent", { ...canonicalRoot, id: "legacy-top-level-only" });
		publish("message_sent", {});
		publish("message_sent", { message: [] });
		publish("message_sent", {
			message: { id: "nested-missing-sender", channel_id: channelId },
		});
		publish("message_sent", {
			message: { ...canonicalRoot, id: "foreign-nested", channel_id: "channel-foreign" },
		});
		publish("message_sent", {
			message: { ...canonicalRoot, id: "reply-must-not-enter-root", thread_id: threadId },
		});

		expect(before?.pages).toEqual([[existing]]);
		expect(cached(rootQueryKey)?.pages).toEqual([[existing, canonicalRoot]]);
	});

	it("routes nested replies only to their exact channel-thread cache and leaves root and other thread caches unchanged", async () => {
		await mountRoot();
		await mountThread();
		const root = cacheMessage(threadId, "thread root");
		const existingReply = cacheMessage("reply-existing", "existing reply", {
			thread_id: threadId,
		});
		const otherThreadReply = cacheMessage("other-thread-existing", "other thread reply", {
			thread_id: "root-other",
		});
		seed(rootQueryKey, [[root]]);
		seed(threadQueryKey(threadId), [[existingReply]]);
		seed(threadQueryKey("root-other"), [[otherThreadReply]]);
		const rootBefore = cached(rootQueryKey);
		const otherThreadBefore = cached(threadQueryKey("root-other"));

		const canonicalReply: Message = {
			id: "reply-rich-1",
			channel_id: channelId,
			sender_id: "reply-sender",
			content_json: { type: "doc", children: [{ type: "paragraph", text: "thread reply" }] },
			content_text: "thread reply",
			attachments: [
				{
					file_id: "reply-file",
					file_name: "reply.txt",
					file_size: 7,
					mime_type: "text/plain",
					url: "https://files.example.test/reply.txt",
				},
			],
			reactions: [{ emoji: "eyes", user_ids: ["watcher-1"] }],
			mentions: ["watcher-1"],
			client_nonce: "903da1aa-9cb4-4e86-8b94-548a68c49f0e",
			thread_id: threadId,
			reply_count: 0,
			created_at: "2026-08-25T10:02:00Z",
		};

		publish("message_sent", { message: canonicalReply });
		publish("message_sent", { message: canonicalReply });
		publish("message_sent", {
			message: { ...canonicalReply, id: "wrong-thread", thread_id: "root-other" },
		});
		publish("message_sent", {
			message: { ...canonicalReply, id: "foreign-thread", channel_id: "channel-foreign" },
		});

		expect(rootBefore?.pages).toEqual([[root]]);
		expect(otherThreadBefore?.pages).toEqual([[otherThreadReply]]);
		expect(cached(rootQueryKey)?.pages).toEqual([[root]]);
		expect(cached(threadQueryKey(threadId))?.pages).toEqual([
			[existingReply, canonicalReply],
		]);
		expect(cached(threadQueryKey("root-other"))?.pages).toEqual([
			[otherThreadReply],
		]);
	});

	it("edits and deletes root and exact thread cache entries while rejecting foreign and missing-channel deltas", async () => {
		await mountRoot();
		await mountThread();
		const rootEditable = cacheMessage("root-editable", "root before");
		const rootCollision = cacheMessage("colliding-id", "root collision");
		const threadEditable = cacheMessage("thread-editable", "thread before", {
			thread_id: threadId,
		});
		const threadCollision = cacheMessage("colliding-id", "thread collision", {
			thread_id: threadId,
		});
		seed(rootQueryKey, [[rootEditable, rootCollision]]);
		seed(threadQueryKey(threadId), [[threadEditable, threadCollision]]);
		const rootBefore = cached(rootQueryKey);
		const threadBefore = cached(threadQueryKey(threadId));

		publish("message_edited", {
			channel_id: channelId,
			message_id: "root-editable",
			content_text: "root after",
			content_json: { type: "doc", text: "root after" },
			edited_at: "2026-08-25T11:00:00Z",
		});
		publish("message_edited", {
			channel_id: channelId,
			message_id: "thread-editable",
			content_text: "thread after",
			content_json: { type: "doc", text: "thread after" },
			edited_at: "2026-08-25T11:01:00Z",
		});

		expect(rootBefore?.pages).toEqual([[rootEditable, rootCollision]]);
		expect(threadBefore?.pages).toEqual([[threadEditable, threadCollision]]);
		expect(cached(rootQueryKey)?.pages[0]?.[0]).toEqual({
			...rootEditable,
			content_text: "root after",
			content_json: { type: "doc", text: "root after" },
			edited_at: "2026-08-25T11:00:00Z",
		});
		expect(cached(threadQueryKey(threadId))?.pages[0]?.[0]).toEqual({
			...threadEditable,
			content_text: "thread after",
			content_json: { type: "doc", text: "thread after" },
			edited_at: "2026-08-25T11:01:00Z",
		});

		publish("message_deleted", {
			channel_id: channelId,
			message_id: "root-editable",
		});
		publish("message_deleted", {
			channel_id: channelId,
			message_id: "thread-editable",
		});
		expect(cached(rootQueryKey)?.pages).toEqual([[rootCollision]]);
		expect(cached(threadQueryKey(threadId))?.pages).toEqual([[threadCollision]]);
		const rootAfterDelete = cached(rootQueryKey);
		const threadAfterDelete = cached(threadQueryKey(threadId));

		publish("message_edited", {
			channel_id: "channel-foreign",
			message_id: "colliding-id",
			content_text: "must not overwrite",
		});
		publish("message_deleted", {
			channel_id: "channel-foreign",
			message_id: "colliding-id",
		});
		publish("message_edited", {
			message_id: "colliding-id",
			content_text: "missing channel must not overwrite",
		});
		publish("message_deleted", { message_id: "colliding-id" });

		expect(cached(rootQueryKey)).toBe(rootAfterDelete);
		expect(cached(threadQueryKey(threadId))).toBe(threadAfterDelete);
	});

	it("ignores a malformed cache page without throwing or mutating it", async () => {
		await mountRoot();
		const malformedCache = { pages: [[null]], pageParams: [undefined] };
		queryClient.setQueryData(rootQueryKey, malformedCache);
		const stored = queryClient.getQueryData(rootQueryKey);

		expect(() => {
			publish("message_sent", { message: cacheMessage("valid-after-malformed", "valid") });
		}).not.toThrow();
		expect(queryClient.getQueryData(rootQueryKey)).toBe(stored);
		expect(queryClient.getQueryData(rootQueryKey)).toEqual({
			pages: [[null]],
			pageParams: [undefined],
		});
	});

	it("shares one channel room for root and thread hooks, then resubscribes before exact recovery invalidations", async () => {
		const order: string[] = [];
		useWsStore.setState({ socket: recordingSocket(order) });
		const invalidateSpy = recordInvalidations(order);
		const rootHook = await mountRoot();
		const threadHook = await mountThread();
		seed(rootQueryKey, [[cacheMessage("root-recovery", "root recovery")]]);
		seed(threadQueryKey(threadId), [[cacheMessage("reply-recovery", "thread recovery", { thread_id: threadId })]]);

		order.length = 0;
		invalidateSpy.mockClear();
		act(() => {
			useWsStore.setState({ connected: true });
		});
		await waitFor(() => {
			expect(order.filter((entry) => entry === `subscribe:channel:${channelId}`)).toHaveLength(1);
		});
		expect(invalidateSpy).not.toHaveBeenCalled();

		act(() => {
			useWsStore.setState({ connected: false });
		});
		await waitFor(() => {
			expect(order).toEqual([`subscribe:channel:${channelId}`]);
		});

		order.length = 0;
		invalidateSpy.mockClear();
		act(() => {
			useWsStore.setState({ connected: true });
		});
		await waitFor(() => {
			expect(order.filter((entry) => entry === `subscribe:channel:${channelId}`)).toHaveLength(1);
			expect(invalidateSpy).toHaveBeenCalledTimes(2);
		});
		const invalidationFilters = invalidateSpy.mock.calls.map(([filters]) => filters);
		expect(invalidationFilters).toEqual(
			expect.arrayContaining([
				expect.objectContaining({ queryKey: rootQueryKey, exact: true }),
				expect.objectContaining({ queryKey: threadQueryKey(threadId), exact: true }),
			]),
		);
		const resubscribeIndex = order.indexOf(`subscribe:channel:${channelId}`);
		const invalidationIndexes = order
			.map((entry, index) => (entry.startsWith("invalidate:") ? index : -1))
			.filter((index) => index >= 0);
		expect(invalidationIndexes).toHaveLength(2);
		expect(invalidationIndexes.every((index) => resubscribeIndex < index)).toBe(true);

		order.length = 0;
		threadHook.unmount();
		expect(order).not.toContain(`unsubscribe:channel:${channelId}`);
		rootHook.unmount();
		await waitFor(() => {
			expect(order).toEqual([`unsubscribe:channel:${channelId}`]);
		});
	});

	it("gives an isolated thread hook its own room lifecycle and exact recovery invalidation", async () => {
		const order: string[] = [];
		useWsStore.setState({ socket: recordingSocket(order) });
		const invalidateSpy = recordInvalidations(order);
		const threadHook = await mountThread();
		seed(threadQueryKey(threadId), [[cacheMessage("thread-only", "thread only", { thread_id: threadId })]]);

		order.length = 0;
		invalidateSpy.mockClear();
		act(() => {
			useWsStore.setState({ connected: true });
		});
		await waitFor(() => {
			expect(order).toEqual([`subscribe:channel:${channelId}`]);
		});
		expect(invalidateSpy).not.toHaveBeenCalled();

		act(() => {
			useWsStore.setState({ connected: false });
		});
		await waitFor(() => {
			expect(order).toEqual([`subscribe:channel:${channelId}`]);
		});

		order.length = 0;
		invalidateSpy.mockClear();
		act(() => {
			useWsStore.setState({ connected: true });
		});
		await waitFor(() => {
			expect(order.filter((entry) => entry === `subscribe:channel:${channelId}`)).toHaveLength(1);
			expect(invalidateSpy).toHaveBeenCalledTimes(1);
		});
		expect(invalidateSpy.mock.calls[0]?.[0]).toEqual(
		expect.objectContaining({ queryKey: threadQueryKey(threadId), exact: true }),
		);
		expect(order.indexOf(`subscribe:channel:${channelId}`)).toBeLessThan(
		order.findIndex((entry) => entry.startsWith("invalidate:")),
	);

		order.length = 0;
		threadHook.unmount();
		await waitFor(() => {
			expect(order).toEqual([`unsubscribe:channel:${channelId}`]);
		});
	});
});