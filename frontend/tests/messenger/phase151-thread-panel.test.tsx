import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message, ThreadResponse } from "@/api/messenger";
import { ThreadPanel } from "@/components/messenger/thread-panel";
import { useAuthStore } from "@/stores/auth-store";

const getThread = vi.fn();
const getThreadReplies = vi.fn();
const getChannels = vi.fn();
const legacyTimelineFetch = vi.fn();

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		getChannels: (...args: unknown[]) => getChannels(...args),
		getThread: (...args: unknown[]) => getThread(...args),
		getThreadReplies: (...args: unknown[]) => getThreadReplies(...args),
		getMessages: (...args: unknown[]) => legacyTimelineFetch(...args),
	};
});

vi.mock("@/components/messenger/message-item", () => ({
	MessageItem: ({ message }: { message: Message }) => (
		<div data-testid="thread-message">{message.id}</div>
	),
}));

vi.mock("@/components/messenger/message-composer", () => ({
	MessageComposer: ({ threadId }: { threadId?: string }) => (
		<div data-testid="thread-composer">{threadId}</div>
	),
}));

vi.mock("@/components/messenger/thread-issue-links", () => ({
	ThreadIssueLinks: () => null,
}));

function message(id: string, threadId?: string): Message {
	return {
		id,
		channel_id: "ch-1",
		sender_id: "user-1",
		content_text: id,
		thread_id: threadId,
		reply_count: 0,
		created_at: "2026-08-26T00:00:00Z",
	};
}

describe("Phase 151 exact thread panel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		useAuthStore.setState({
			user: {
				id: "user-1",
				email: "user-1@example.test",
				timezone: "UTC",
				language: "en",
				system_role: "member",
				onboarding_completed: true,
			},
		});
		const response: ThreadResponse = {
			parent: message("root-1"),
			replies: [],
			follow_mode: "manual_follow",
		};
		getThread.mockResolvedValue(response);
		getThreadReplies.mockResolvedValue([message("reply-1", "root-1")]);
		getChannels.mockResolvedValue([]);
		legacyTimelineFetch.mockRejectedValue(
			new Error("timeline fallback is forbidden"),
		);
	});

	afterEach(() => {
		useAuthStore.setState({ user: null });
	});

	function wrapper({ children }: { children: ReactNode }) {
		return (
			<QueryClientProvider
				client={
					new QueryClient({
						defaultOptions: { queries: { retry: false } },
					})
				}
			>
				{children}
			</QueryClientProvider>
		);
	}

	it("renders the root detail and canonical paged replies", async () => {
		render(
			<ThreadPanel
				channelId="ch-1"
				messageId="root-1"
				onClose={() => undefined}
			/>,
			{ wrapper },
		);

		await waitFor(() => {
						expect(getThread).toHaveBeenCalledWith("ch-1", "root-1");
			expect(getThreadReplies).toHaveBeenCalledWith(
				"ch-1",
				"root-1",
				expect.objectContaining({ limit: 30 }),
			);
		});
		const messages = await screen.findAllByTestId("thread-message");
		expect(messages).toHaveLength(2);
		expect(messages[0]).toHaveTextContent("root-1");
		expect(messages[1]).toHaveTextContent("reply-1");
		expect(screen.getByTestId("thread-composer")).toHaveTextContent("root-1");
		expect(legacyTimelineFetch).not.toHaveBeenCalled();
	});

	it("moves focus into the panel and returns it to the opener on close", async () => {
		const opener = document.createElement("button");
		opener.textContent = "open thread";
		document.body.appendChild(opener);
		opener.focus();
		const onClose = vi.fn();

		render(
			<ThreadPanel
				channelId="ch-1"
				messageId="root-1"
				onClose={onClose}
			/>,
			{ wrapper },
		);

		const close = await screen.findByRole("button", { name: /закрыть тред/i });
		await waitFor(() => expect(close).toHaveFocus());
		fireEvent.click(close);

		expect(onClose).toHaveBeenCalledTimes(1);
		expect(opener).toHaveFocus();
		opener.remove();
	});
});
