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

type ReactionDelta = {
	message_id: string;
	channel_id: string;
	user_id: string;
	emoji: string;
	added: boolean;
};

const channelId = "channel-1";
const threadId = "root-1";
const rootQueryKey = ["messenger", "messages", channelId] as const;
const threadQueryKey = ["messenger", "thread", channelId, threadId] as const;

function message(id: string, thread?: string): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "sender-1",
		content_text: id,
		content_json: { type: "doc", text: id },
		thread_id: thread,
		reply_count: 0,
		created_at: "2026-08-25T10:05:00Z",
	};
}

describe("reaction_toggled delta reconciliation", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
		});
		api.getMessages.mockResolvedValue([]);
		api.getThreadReplies.mockResolvedValue([]);
		useWsStore.setState({
			connected: true,
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

	function emit(delta: ReactionDelta | unknown): void {
		act(() => {
			for (const listener of [...useWsStore.getState().listeners]) {
				listener("reaction_toggled", delta);
			}
		});
	}

	async function mountRootAndThread() {
		const root = renderHook(() => useMessages(channelId), { wrapper: TestProvider });
		const thread = renderHook(() => useThreadMessages(channelId, threadId), {
			wrapper: TestProvider,
		});
		await waitFor(() => {
			expect(root.result.current.isSuccess).toBe(true);
			expect(thread.result.current.isSuccess).toBe(true);
			expect(useWsStore.getState().listeners.size).toBeGreaterThanOrEqual(2);
		});
	}

	it("applies idempotent add/remove deltas to root and thread caches without mutating previous cache references", async () => {
		await mountRootAndThread();
		const rootMessage: Message = {
			...message("root-message"),
			reactions: [
				{ emoji: "thumbsup", user_ids: ["user-a", "user-b"] },
				{ emoji: "fire", user_ids: ["user-fire"] },
			],
		};
		const threadMessage: Message = {
			...message("thread-message", threadId),
			reactions: [
				{ emoji: "sparkles", user_ids: ["thread-user"] },
				{ emoji: "eyes", user_ids: ["watcher"] },
			],
		};
		seed(rootQueryKey, [[rootMessage]]);
		seed(threadQueryKey, [[threadMessage]]);
		const rootBefore = cached(rootQueryKey);
		const threadBefore = cached(threadQueryKey);

		const addRoot: ReactionDelta = {
			message_id: "root-message",
			channel_id: channelId,
			user_id: "user-new",
			emoji: "thumbsup",
			added: true,
		};
		emit(addRoot);
		emit(addRoot);
		emit({
			message_id: "root-message",
			channel_id: channelId,
			user_id: "user-a",
			emoji: "thumbsup",
			added: false,
		});
		emit({
			message_id: "root-message",
			channel_id: channelId,
			user_id: "user-a",
			emoji: "thumbsup",
			added: false,
		});
		emit({
			message_id: "thread-message",
			channel_id: channelId,
			user_id: "thread-user",
			emoji: "sparkles",
			added: false,
		});

		expect(rootBefore?.pages).toEqual([[rootMessage]]);
		expect(threadBefore?.pages).toEqual([[threadMessage]]);
		expect(cached(rootQueryKey)?.pages[0]?.[0]?.reactions).toEqual([
			{ emoji: "thumbsup", user_ids: ["user-b", "user-new"] },
			{ emoji: "fire", user_ids: ["user-fire"] },
		]);
		expect(cached(threadQueryKey)?.pages[0]?.[0]?.reactions).toEqual([
			{ emoji: "eyes", user_ids: ["watcher"] },
		]);
	});

	it("ignores foreign, unknown, and malformed delta fields without mutating either cache", async () => {
		await mountRootAndThread();
		const rootMessage: Message = {
			...message("root-message"),
			reactions: [{ emoji: "thumbsup", user_ids: ["user-a"] }],
		};
		const threadMessage: Message = {
			...message("thread-message", threadId),
			reactions: [{ emoji: "eyes", user_ids: ["watcher"] }],
		};
		seed(rootQueryKey, [[rootMessage]]);
		seed(threadQueryKey, [[threadMessage]]);
		const rootBefore = cached(rootQueryKey);
		const threadBefore = cached(threadQueryKey);

		emit({
			message_id: "root-message",
			channel_id: "channel-foreign",
			user_id: "user-new",
			emoji: "thumbsup",
			added: true,
		});
		emit({
			message_id: "missing-message",
			channel_id: channelId,
			user_id: "user-new",
			emoji: "thumbsup",
			added: true,
		});
		emit({
			message_id: "root-message",
			channel_id: channelId,
			user_id: "",
			emoji: "thumbsup",
			added: true,
		});
		emit({
			message_id: "root-message",
			channel_id: channelId,
			user_id: "user-new",
			emoji: "",
			added: true,
		});
		emit({
			message_id: "root-message",
			channel_id: channelId,
			user_id: "user-new",
			emoji: "thumbsup",
			added: "yes",
		});


		expect(cached(rootQueryKey)).toBe(rootBefore);
		expect(cached(threadQueryKey)).toBe(threadBefore);
	});
});