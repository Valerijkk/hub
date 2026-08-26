import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
	act,
	cleanup,
	render,
	renderHook,
	screen,
	waitFor,
} from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Message, ThreadResponse } from "@/api/messenger";
import { ThreadPanel } from "@/components/messenger/thread-panel";
import { useInvalidateMessages } from "@/hooks/use-invalidate";
import {
	messagesQueryKey,
	threadMessagesQueryKey,
	useThreadMessages,
} from "@/hooks/use-messages";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { useWsStore } from "@/stores/ws-store";

const api = vi.hoisted(() => ({
	getThread: vi.fn(),
	getThreadReplies: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getThread: api.getThread,
		getThreadReplies: api.getThreadReplies,
	};
});

vi.mock("@/components/messenger/message-item", () => ({
	MessageItem: ({ message }: { message: Message }) => (
		<div data-testid={`thread-message-${message.id}`}>{message.content_text}</div>
	),
}));

vi.mock("@/components/messenger/message-composer", () => ({
	MessageComposer: () => <div data-testid="thread-composer" />,
}));

vi.mock("@/components/messenger/thread-issue-links", () => ({
	ThreadIssueLinks: () => null,
}));

const channelId = "thread-convergence-channel";
const threadId = "thread-convergence-root";
const user = {
	id: "thread-convergence-user",
	email: "thread-convergence@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};

function message(id: string, text: string, parentId?: string): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "sender-1",
		content_text: text,
		content_json: { type: "doc", text },
		thread_id: parentId,
		reply_count: 0,
		created_at: "2026-08-26T08:00:00Z",
	};
}

function detail(parent: Message, replies: Message[]): ThreadResponse {
	return { parent, replies, follow_mode: "manual_follow" };
}

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

describe("Phase 151 ThreadPanel realtime convergence", () => {
	let queryClient: QueryClient;

	beforeEach(() => {
		queryClient = new QueryClient({
			defaultOptions: {
				queries: { retry: false },
				mutations: { retry: false },
			},
		});
		api.getThread.mockReset();
		api.getThreadReplies.mockReset();
		useAuthStore.setState({ user });
		useMessengerStore.getState().reset();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
	});

	afterEach(() => {
		cleanup();
		queryClient.clear();
		useAuthStore.setState({ user: null });
		useMessengerStore.getState().reset();
		useWsStore.setState({ connected: false, socket: null, listeners: new Set() });
		vi.clearAllMocks();
	});

	function TestProvider({ children }: { children: ReactNode }) {
		return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
	}

	function renderPanel() {
		return render(
			<ThreadPanel
				channelId={channelId}
				messageId={threadId}
				canReply
				onClose={vi.fn()}
			/>,
			{ wrapper: TestProvider },
		);
	}

	function publish(event: string, payload: unknown) {
		act(() => {
			for (const listener of [...useWsStore.getState().listeners]) {
				listener(event, payload);
			}
		});
	}

	it("renders only canonical paged replies and appends a loaded-thread WS reply", async () => {
		const parent = message(threadId, "Canonical parent");
		const staleDetailReply = message("stale-detail", "Stale detail reply", threadId);
		const canonicalReply = message("canonical-reply", "Canonical paged reply", threadId);
		const wsReply = message("ws-reply", "Realtime WS reply", threadId);
		api.getThread.mockResolvedValue(detail(parent, [staleDetailReply]));
		api.getThreadReplies.mockResolvedValue([canonicalReply]);

		renderPanel();

		expect(await screen.findByText("Canonical parent")).toBeInTheDocument();
		expect(await screen.findByText("Canonical paged reply")).toBeInTheDocument();
		expect(screen.queryByText("Stale detail reply")).not.toBeInTheDocument();
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);

		publish("message_sent", { message: wsReply });
		expect(await screen.findByText("Realtime WS reply")).toBeInTheDocument();
	});

	it("replays a WS reply that arrives while the initial paged fetch is in flight", async () => {
		const parent = message(threadId, "Race parent");
		const initialReplies = deferred<Message[]>();
		const wsReply = message("ws-race-reply", "WS during initial fetch", threadId);
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies.mockReturnValue(initialReplies.promise);

		renderPanel();
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(1));
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);
		publish("message_sent", { message: wsReply });

		await act(async () => {
			initialReplies.resolve([]);
			await initialReplies.promise;
		});

		expect(await screen.findByText("WS during initial fetch")).toBeInTheDocument();
	});

	it("retains a WS reply when a refetch replaces resident thread cache", async () => {
		const parent = message(threadId, "Resident parent");
		const initialReply = message("resident-reply", "Resident reply", threadId);
		const staleRefetchReply = message(
			"stale-refetch-reply",
			"Stale refetch reply",
			threadId,
		);
		const wsReply = message("resident-ws-reply", "WS during refetch", threadId);
		const refetchReplies = deferred<Message[]>();
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies
			.mockResolvedValueOnce([initialReply])
			.mockReturnValueOnce(refetchReplies.promise);

		renderPanel();
		expect(await screen.findByText("Resident reply")).toBeInTheDocument();
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);

		act(() => {
			void queryClient.invalidateQueries({
				queryKey: threadMessagesQueryKey(channelId, threadId, user.id),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));

		publish("message_sent", { message: wsReply });
		expect(await screen.findByText("WS during refetch")).toBeInTheDocument();

		await act(async () => {
			refetchReplies.resolve([staleRefetchReply]);
			await refetchReplies.promise;
		});

		expect(await screen.findByText("Stale refetch reply")).toBeInTheDocument();
		expect(await screen.findByText("WS during refetch")).toBeInTheDocument();
	});

	it("recovers an evicted initial-fetch WS reply through one canonical refetch", async () => {
		const parent = message(threadId, "Overflow parent");
		const initialReplies = deferred<Message[]>();
		const wsReplies = Array.from({ length: 101 }, (_, index) =>
			message(
				"overflow-reply-" + (index + 1),
				"Overflow reply " + (index + 1),
				threadId,
			),
		);
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies
			.mockReturnValueOnce(initialReplies.promise)
			.mockResolvedValueOnce(wsReplies);

		renderPanel();
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(1));
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);
		for (const wsReply of wsReplies) {
			publish("message_sent", { message: wsReply });
		}

		await act(async () => {
			initialReplies.resolve([]);
			await initialReplies.promise;
		});

		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));
		expect(await screen.findByText("Overflow reply 1")).toBeInTheDocument();
		expect(api.getThreadReplies).toHaveBeenCalledTimes(2);
	});

	it("recovers stale refetch edits and deletes with the canonical response", async () => {
		const parent = message(threadId, "Delta parent");
		const editable = message("delta-edit", "Original edit", threadId);
		const deletable = message("delta-delete", "Original delete", threadId);
		const staleRefetch = deferred<Message[]>();
		const canonicalRecovery = deferred<Message[]>();
		const canonicalEdit = message("delta-edit", "Canonical edit", threadId);
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies
			.mockResolvedValueOnce([editable, deletable])
			.mockReturnValueOnce(staleRefetch.promise)
			.mockReturnValueOnce(canonicalRecovery.promise);

		renderPanel();
		expect(await screen.findByText("Original edit")).toBeInTheDocument();
		expect(await screen.findByText("Original delete")).toBeInTheDocument();
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);

		act(() => {
			void queryClient.invalidateQueries({
				queryKey: threadMessagesQueryKey(channelId, threadId, user.id),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));
		publish("message_edited", {
			channel_id: channelId,
			thread_id: threadId,
			message_id: editable.id,
			content_text: "Realtime edit",
			content_json: { type: "doc", text: "Realtime edit" },
			edited_at: "2026-08-26T10:00:00Z",
		});
		publish("message_deleted", {
			channel_id: channelId,
			thread_id: threadId,
			message_id: deletable.id,
		});
		expect(await screen.findByText("Realtime edit")).toBeInTheDocument();
		expect(screen.queryByText("Original delete")).not.toBeInTheDocument();
		expect(api.getThreadReplies).toHaveBeenCalledTimes(2);

		await act(async () => {
			staleRefetch.resolve([editable, deletable]);
			await staleRefetch.promise;
		});

		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(3));
		expect(screen.getByText("Realtime edit")).toBeInTheDocument();
		expect(screen.queryByText("Original delete")).not.toBeInTheDocument();

		await act(async () => {
			canonicalRecovery.resolve([canonicalEdit]);
			await canonicalRecovery.promise;
		});

		expect(await screen.findByText("Canonical edit")).toBeInTheDocument();
		expect(screen.queryByText("Original delete")).not.toBeInTheDocument();
		expect(api.getThreadReplies).toHaveBeenCalledTimes(3);
	});

	it("masks an overflowed mutation journal until canonical thread recovery completes", async () => {
		const residentReplies = Array.from({ length: 101 }, (_, index) =>
			message(
				`mutation-overflow-${index + 1}`,
				`Mutation overflow ${index + 1}`,
				threadId,
			),
		);
		const staleRefetch = deferred<Message[]>();
		const canonicalRecovery = deferred<Message[]>();
		const canonicalReply = message(
			"mutation-overflow-canonical",
			"Canonical mutation recovery",
			threadId,
		);
		api.getThreadReplies
			.mockResolvedValueOnce(residentReplies)
			.mockReturnValueOnce(staleRefetch.promise)
			.mockReturnValueOnce(canonicalRecovery.promise);

		const hook = renderHook(() => useThreadMessages(channelId, threadId), {
			wrapper: TestProvider,
		});
		await waitFor(() => expect(hook.result.current.messages).toHaveLength(101));
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);

		act(() => {
			void queryClient.invalidateQueries({
				queryKey: threadMessagesQueryKey(channelId, threadId, user.id),
				exact: true,
			});
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));

		for (const reply of residentReplies) {
			publish("message_edited", {
				channel_id: channelId,
				thread_id: threadId,
				message_id: reply.id,
				content_text: `Realtime ${reply.id}`,
				content_json: { type: "doc", text: `Realtime ${reply.id}` },
				edited_at: "2026-08-26T10:02:00Z",
			});
		}

		expect(hook.result.current.data).toBeUndefined();
		expect(hook.result.current.messages).toEqual([]);
		expect(api.getThreadReplies).toHaveBeenCalledTimes(2);

		await act(async () => {
			staleRefetch.resolve(residentReplies);
			await staleRefetch.promise;
		});

		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(3));
		expect(hook.result.current.data).toBeUndefined();
		expect(hook.result.current.messages).toEqual([]);

		await act(async () => {
			canonicalRecovery.resolve([canonicalReply]);
			await canonicalRecovery.promise;
		});

		await waitFor(() =>
			expect(hook.result.current.messages.map((reply) => reply.content_text)).toEqual([
				"Canonical mutation recovery",
			]),
		);
		expect(hook.result.current.data).toBeDefined();
		expect(api.getThreadReplies).toHaveBeenCalledTimes(3);
	});

	it("reconciles an edit received before the initial thread cache exists", async () => {
		const parent = message(threadId, "Initial delta parent");
		const initialReplies = deferred<Message[]>();
		const staleInitialReply = message("initial-delta", "Stale initial edit", threadId);
		const canonicalReply = message("initial-delta", "Canonical initial edit", threadId);
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies
			.mockReturnValueOnce(initialReplies.promise)
			.mockResolvedValueOnce([canonicalReply]);

		renderPanel();
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(1));
		await waitFor(() =>
			expect(useWsStore.getState().listeners.size).toBeGreaterThan(0),
		);
		publish("message_edited", {
			channel_id: channelId,
			thread_id: threadId,
			message_id: staleInitialReply.id,
			content_text: "Realtime initial edit",
			content_json: { type: "doc", text: "Realtime initial edit" },
			edited_at: "2026-08-26T10:01:00Z",
		});

		await act(async () => {
			initialReplies.resolve([staleInitialReply]);
			await initialReplies.promise;
		});

		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));
		expect(await screen.findByText("Canonical initial edit")).toBeInTheDocument();
		expect(api.getThreadReplies).toHaveBeenCalledTimes(2);
	});

	it("shows reply loading instead of an empty state after the parent detail loads", async () => {
		const parent = message(threadId, "Loaded parent");
		const initialReplies = deferred<Message[]>();
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies.mockReturnValue(initialReplies.promise);

		renderPanel();

		expect(await screen.findByText("Loaded parent")).toBeInTheDocument();
		expect(screen.getByText(/Загрузка/)).toBeInTheDocument();
		expect(screen.queryByText(/Нет ответов/)).not.toBeInTheDocument();

		await act(async () => {
			initialReplies.resolve([]);
			await initialReplies.promise;
		});
	});

	it("masks parent, replies, and composer immediately when the channel is revoked", async () => {
		const parent = message(threadId, "Revoked parent");
		const reply = message("revoked-reply", "Revoked reply", threadId);
		api.getThread.mockResolvedValue(detail(parent, []));
		api.getThreadReplies.mockResolvedValue([reply]);

		renderPanel();
		expect(await screen.findByText("Revoked parent")).toBeInTheDocument();
		expect(await screen.findByText("Revoked reply")).toBeInTheDocument();
		expect(screen.getByTestId("thread-composer")).toBeInTheDocument();

		act(() => {
			useMessengerStore.getState().revokeChannel(channelId);
		});

		expect(screen.queryByText("Revoked parent")).not.toBeInTheDocument();
		expect(screen.queryByText("Revoked reply")).not.toBeInTheDocument();
		expect(screen.queryByTestId("thread-composer")).not.toBeInTheDocument();
	});

	it("invalidates exact root and snapshotted thread keys and fences identity changes", async () => {
		const invalidate = vi.spyOn(queryClient, "invalidateQueries");
		const hook = renderHook(() => useInvalidateMessages(channelId), {
			wrapper: TestProvider,
		});
		const callbackForUser = hook.result.current;

		act(() => {
			callbackForUser(threadId);
		});
		expect(invalidate).toHaveBeenCalledWith({
			queryKey: messagesQueryKey(channelId, user.id),
			exact: true,
		});
		expect(invalidate).toHaveBeenCalledWith({
			queryKey: threadMessagesQueryKey(channelId, threadId, user.id),
			exact: true,
		});

		invalidate.mockClear();
		act(() => {
			useAuthStore.setState({ user: { ...user, id: "later-user" } });
		});
		callbackForUser(threadId);
		expect(invalidate).not.toHaveBeenCalled();

		act(() => {
			useAuthStore.setState({ user: null });
		});
		hook.result.current(threadId);
		expect(invalidate).not.toHaveBeenCalled();
	});
});
