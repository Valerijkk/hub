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

import type { Message } from "@/api/messenger";
import { ThreadPanel } from "@/components/messenger/thread-panel";
import { isStaleMessengerIdentityQuery } from "@/hooks/use-messenger-channels-realtime";
import { useThreadMessages } from "@/hooks/use-messages";
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

const channelId = "shared-thread-channel";
const threadId = "shared-thread-root";
const userA = {
	id: "thread-user-a",
	email: "thread-user-a@example.test",
	timezone: "UTC",
	language: "en",
	system_role: "member",
	onboarding_completed: true,
};
const userB = {
	...userA,
	id: "thread-user-b",
	email: "thread-user-b@example.test",
};

function message(id: string, text: string, thread?: string): Message {
	return {
		id,
		channel_id: channelId,
		sender_id: "sender-1",
		content_text: text,
		content_json: { type: "doc", text },
		thread_id: thread,
		reply_count: 0,
		created_at: "2026-08-26T07:00:00Z",
	};
}

function deferred<T>() {
	let resolve: (value: T) => void = () => undefined;
	const promise = new Promise<T>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

describe("Phase 151 thread query identity boundary", () => {
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
		useAuthStore.setState({ user: userA });
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

	it("keeps delayed paged replies for A out of the same thread rendered for B", async () => {
		const fromA = deferred<Message[]>();
		const fromB = deferred<Message[]>();
		const replyA = message("reply-a", "Private reply A", threadId);
		const replyB = message("reply-b", "Private reply B", threadId);
		api.getThreadReplies
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);

		const hook = renderHook(() => useThreadMessages(channelId, threadId), {
			wrapper: TestProvider,
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(1));

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() => expect(api.getThreadReplies).toHaveBeenCalledTimes(2));

		await act(async () => {
			fromA.resolve([replyA]);
			await fromA.promise;
		});
		expect(hook.result.current.messages).not.toContainEqual(replyA);

		await act(async () => {
			fromB.resolve([replyB]);
			await fromB.promise;
		});
		await waitFor(() => expect(hook.result.current.messages).toEqual([replyB]));

		expect(
			queryClient.getQueryData<{ pages: Message[][] }>([
				"messenger",
				"thread-messages",
				channelId,
				threadId,
				userA.id,
			])?.pages.flat(),
		).toEqual([replyA]);
		expect(
			queryClient.getQueryData<{ pages: Message[][] }>([
				"messenger",
				"thread-messages",
				channelId,
				threadId,
				userB.id,
			])?.pages.flat(),
		).toEqual([replyB]);
	});

	it("does not fetch or expose paged replies from legacy/sentinel caches without identity", async () => {
		const privateReply = message("private-reply", "Private cached reply", threadId);
		queryClient.setQueryData(
			["messenger", "thread-messages", channelId, threadId],
			{ pages: [[privateReply]], pageParams: [undefined] },
		);
		queryClient.setQueryData(
			["messenger", "thread-messages", channelId, threadId, "__no-user__"],
			{ pages: [[privateReply]], pageParams: [undefined] },
		);
		useAuthStore.setState({ user: null });

		const hook = renderHook(() => useThreadMessages(channelId, threadId), {
			wrapper: TestProvider,
		});
		await act(async () => Promise.resolve());

		expect(api.getThreadReplies).not.toHaveBeenCalled();
		expect(hook.result.current.data).toBeUndefined();
		expect(hook.result.current.messages).toEqual([]);
	});

	it("keeps delayed production thread detail for A out of the panel rendered for B", async () => {
		const fromA = deferred<{ parent: Message }>();
		const fromB = deferred<{ parent: Message }>();
		const parentA = message("parent-a", "Private parent A");
		const parentB = message("parent-b", "Private parent B");
		api.getThreadReplies.mockResolvedValue([]);
		api.getThread
			.mockImplementationOnce(() => fromA.promise)
			.mockImplementationOnce(() => fromB.promise);

		render(
			<ThreadPanel channelId={channelId} messageId={threadId} onClose={vi.fn()} />,
			{ wrapper: TestProvider },
		);
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(1));

		act(() => {
			useAuthStore.setState({ user: userB });
		});
		await waitFor(() => expect(api.getThread).toHaveBeenCalledTimes(2));

		await act(async () => {
			fromA.resolve({ parent: parentA });
			await fromA.promise;
		});
		expect(screen.queryByText("Private parent A")).not.toBeInTheDocument();
		expect(screen.queryByText("Private detail reply A")).not.toBeInTheDocument();

		await act(async () => {
			fromB.resolve({ parent: parentB });
			await fromB.promise;
		});
		expect(await screen.findByText("Private parent B")).toBeInTheDocument();
		expect(screen.queryByText("Private detail reply A")).not.toBeInTheDocument();
		expect(screen.queryByText("Private detail reply B")).not.toBeInTheDocument();
	});

	it("does not fetch or render production thread detail caches without identity", async () => {
		const privateParent = message("private-parent", "Private cached parent");
		const privateReply = message("private-detail-reply", "Private cached detail reply", threadId);
		const privateThread = { parent: privateParent, replies: [privateReply] };
		queryClient.setQueryData(
			["messenger", "thread", channelId, threadId],
			privateThread,
		);
		queryClient.setQueryData(
			["messenger", "thread", channelId, threadId, "__no-user__"],
			privateThread,
		);
		useAuthStore.setState({ user: null });

		render(
			<ThreadPanel channelId={channelId} messageId={threadId} onClose={vi.fn()} />,
			{ wrapper: TestProvider },
		);
		await act(async () => Promise.resolve());

		expect(api.getThread).not.toHaveBeenCalled();
		expect(screen.queryByText("Private cached parent")).not.toBeInTheDocument();
		expect(screen.queryByText("Private cached detail reply")).not.toBeInTheDocument();
	});

	it("preserves only current-identity Messenger cache schemas during async identity cleanup", () => {
		const currentIdentityKeys = [
			["messenger", "channels", userB.id],
			["messenger", "channels", "discover", userB.id],
			["messenger", "unread", userB.id],
			["messenger", "bookmarks", userB.id],
			["messenger", "messages", channelId, userB.id],
			["messenger", "pins", channelId, userB.id],
			["messenger", "thread", channelId, threadId, userB.id],
			["messenger", "thread-messages", channelId, threadId, userB.id],
		] as const;
		for (const queryKey of currentIdentityKeys) {
			expect(isStaleMessengerIdentityQuery(queryKey, userB.id)).toBe(false);
		}

		expect(
			isStaleMessengerIdentityQuery(
				["messenger", "thread-messages", channelId, threadId, userA.id],
				userB.id,
			),
		).toBe(true);
		expect(
			isStaleMessengerIdentityQuery(
				["messenger", "thread", channelId, threadId, userA.id],
				userB.id,
			),
		).toBe(true);
		expect(
			isStaleMessengerIdentityQuery(
				["messenger", "channel", channelId],
				userB.id,
			),
		).toBe(true);
		expect(
			isStaleMessengerIdentityQuery(["projects"], userB.id),
		).toBe(false);
	});
});
