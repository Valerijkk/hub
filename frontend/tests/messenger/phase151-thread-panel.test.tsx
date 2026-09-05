import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ChannelCapabilities, Message, ThreadResponse } from "@/api/messenger";
import { ThreadPanel } from "@/components/messenger/thread-panel";
import { useAuthStore } from "@/stores/auth-store";

const getThread = vi.fn();
const getThreadReplies = vi.fn();
const getChannels = vi.fn();
const legacyTimelineFetch = vi.fn();
const deliveredCapabilities = new Map<string, ChannelCapabilities | undefined>();

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
	MessageItem: ({ message, capabilities }: { message: Message; capabilities?: ChannelCapabilities }) => {
		deliveredCapabilities.set(message.id, capabilities);
		return <div data-testid="thread-message">{message.id}</div>;
	},
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

function capabilities(overrides: Partial<ChannelCapabilities> = {}): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
		can_read: true,
		can_post: false,
		can_reply: false,
		can_join: false,
		can_follow: false,
		can_unfollow: false,
		can_view_unread: false,
		can_type: false,
		can_edit: false,
		can_forward: false,
		can_react: false,
		can_delete_message: false,
		can_pin: false,
		can_list_pins: false,
		can_bookmark: false,
		can_promote: false,
		...overrides,
	};
}

async function expectBothMessageBoundaries() {
	await waitFor(() => {
		expect(screen.getAllByTestId("thread-message")).toHaveLength(2);
		expect(deliveredCapabilities.has("root-1")).toBe(true);
		expect(deliveredCapabilities.has("reply-1")).toBe(true);
	});
}

describe("Phase 151 exact thread panel", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		deliveredCapabilities.clear();
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
		deliveredCapabilities.clear();
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

	it("delivers the exact channel projection to loaded root and replies", async () => {
		const projection = capabilities({ can_delete_message: true });
		const capabilityProps = { capabilities: projection };
		render(
			<ThreadPanel channelId="ch-1" messageId="root-1" onClose={() => undefined} {...capabilityProps} />,
			{ wrapper },
		);
		await expectBothMessageBoundaries();
		expect(deliveredCapabilities.get("root-1")).toBe(projection);
		expect(deliveredCapabilities.get("reply-1")).toBe(projection);
	});

	it("leaves loaded root and replies without a projection undefined", async () => {
		render(
			<ThreadPanel channelId="ch-1" messageId="root-1" onClose={() => undefined} />,
			{ wrapper },
		);
		await expectBothMessageBoundaries();
		expect(deliveredCapabilities.get("root-1")).toBeUndefined();
		expect(deliveredCapabilities.get("reply-1")).toBeUndefined();
	});

	it("updates both loaded message boundaries for downgrade and withdrawal", async () => {
		const allowed = capabilities({ can_delete_message: true });
		const denied = capabilities();
		const allowedProps = { capabilities: allowed };
		const deniedProps = { capabilities: denied };
		const absentProps: { capabilities?: ChannelCapabilities } = { capabilities: undefined };
		const panelProps = { channelId: "ch-1", messageId: "root-1", onClose: () => undefined };
		const { rerender } = render(<ThreadPanel {...panelProps} {...allowedProps} />, { wrapper });
		await expectBothMessageBoundaries();
		await waitFor(() => {
			expect(deliveredCapabilities.get("root-1")).toBe(allowed);
			expect(deliveredCapabilities.get("reply-1")).toBe(allowed);
		});
		rerender(<ThreadPanel {...panelProps} {...deniedProps} />);
		await waitFor(() => {
			expect(deliveredCapabilities.get("root-1")).toBe(denied);
			expect(deliveredCapabilities.get("reply-1")).toBe(denied);
		});
		rerender(<ThreadPanel {...panelProps} {...absentProps} />);
		await waitFor(() => {
			expect(deliveredCapabilities.get("root-1")).toBeUndefined();
			expect(deliveredCapabilities.get("reply-1")).toBeUndefined();
		});
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
