import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel, ChannelCapabilities, Message } from "@/api/messenger";
import { ForwardDialog } from "@/components/messenger/forward-dialog";
import { MessageActions } from "@/components/messenger/message-actions";
import { useMessengerStore } from "@/stores/messenger-store";

const api = vi.hoisted(() => ({
	forwardMessage: vi.fn(),
}));

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return { ...actual, forwardMessage: api.forwardMessage };
});

function capabilities(
	overrides: Partial<ChannelCapabilities> = {},
): ChannelCapabilities {
	return {
		is_member: true,
		has_active_project_membership: false,
		can_read: false,
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

function channel(
	id: string,
	overrides: Partial<ChannelCapabilities> = {},
): Channel {
	return {
		id,
		name: id,
		channel_type: "private",
		is_readonly: false,
		capabilities: capabilities(overrides),
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-08-26T00:00:00Z",
		updated_at: "2026-08-26T00:00:00Z",
	};
}

const message: Message = {
	id: "message-1",
	channel_id: "source",
	sender_id: "user-1",
	content_text: "forward me",
	content_json: { type: "doc" },
	reply_count: 0,
	created_at: "2026-08-26T00:00:00Z",
};

const actionCallbacks = {
	onReactionPickerChange: vi.fn(),
	onReply: vi.fn(),
	onOpenThread: vi.fn(),
	onPin: vi.fn(),
	onBookmark: vi.fn(),
	onForward: vi.fn(),
	onEdit: vi.fn(),
	onDelete: vi.fn(),
};

describe("P151 message action capabilities", () => {
	beforeEach(() => {
		api.forwardMessage.mockReset();
		api.forwardMessage.mockResolvedValue(message);
		useMessengerStore.getState().reset();
		vi.clearAllMocks();
	});

	afterEach(() => {
		cleanup();
		useMessengerStore.getState().reset();
	});

	it("fails closed when the server capability snapshot denies every action", () => {
		render(
			<MessageActions
				message={message}
				channelId="source"
				isOwn={false}
				capabilities={capabilities()}
				showReactionPicker={false}
				{...actionCallbacks}
			/>,
		);

		expect(screen.queryByRole("button", { name: "Ответить" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Реакция" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Открыть тред" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Закрепить сообщение" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Сохранить в закладки" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Переслать сообщение" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Редактировать сообщение" })).toBeNull();
		expect(screen.queryByRole("button", { name: "Удалить сообщение" })).toBeNull();
	});

	it("allows an ordinary member to delete only their own message", () => {
		render(
			<MessageActions
				message={message}
				channelId="source"
				isOwn
				capabilities={capabilities({ is_member: true })}
				showReactionPicker={false}
				{...actionCallbacks}
			/>,
		);

		expect(screen.getByRole("button", { name: "Удалить сообщение" })).toBeTruthy();
		expect(screen.queryByRole("button", { name: "Редактировать сообщение" })).toBeNull();
	});
	it("uses can_delete_message directly instead of inferring ownership or role", () => {
		render(
			<MessageActions
				message={message}
				channelId="source"
				isOwn={false}
				capabilities={capabilities({ can_delete_message: true })}
				showReactionPicker={false}
				{...actionCallbacks}
			/>,
		);

		expect(screen.getByRole("button", { name: "Удалить сообщение" })).toBeTruthy();
		expect(screen.queryByRole("button", { name: "Редактировать сообщение" })).toBeNull();
	});

	it("lists only non-revoked targets with a fresh can_post capability", () => {
		useMessengerStore.setState({
			channels: [
				channel("source", { can_forward: true }),
				channel("allowed", { can_post: true }),
				channel("denied", { can_forward: true }),
				channel("revoked", { can_post: true }),
			],
		});
		useMessengerStore.getState().revokeChannel("revoked");

		render(
			<ForwardDialog
				message={message}
				channelId="source"
				canForward
				open
				onOpenChange={vi.fn()}
			/>,
		);

		expect(screen.getByRole("button", { name: "allowed" })).toBeTruthy();
		expect(screen.queryByRole("button", { name: "denied" })).toBeNull();
		expect(screen.queryByRole("button", { name: "revoked" })).toBeNull();
	});

	it("does not complete a forward after the source is revoked in flight", async () => {
		let resolveForward: (value: Message) => void = () => undefined;
		api.forwardMessage.mockImplementationOnce(
			() =>
				new Promise<Message>((resolve) => {
					resolveForward = resolve;
				}),
		);
		useMessengerStore.setState({
			channels: [
				channel("source", { can_forward: true }),
				channel("allowed", { can_post: true }),
			],
		});
		const onOpenChange = vi.fn();
		render(
			<ForwardDialog
				message={message}
				channelId="source"
				canForward
				open
				onOpenChange={onOpenChange}
			/>,
		);

		fireEvent.click(screen.getByRole("button", { name: "allowed" }));
		expect(api.forwardMessage).toHaveBeenCalledTimes(1);
		act(() => useMessengerStore.getState().revokeChannel("source"));
		await act(async () => resolveForward(message));

		await waitFor(() => {
			expect(onOpenChange).not.toHaveBeenCalledWith(false);
		});
	});
});
