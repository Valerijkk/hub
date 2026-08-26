import { fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Message } from "@/api/messenger";
import { MessageComposer } from "@/components/messenger/message-composer";
import { useAuthStore } from "@/stores/auth-store";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

const sendMessage = vi.fn();

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		sendMessage: (...args: unknown[]) => sendMessage(...args),
		sendTypingIndicator: vi.fn(),
	};
});

vi.mock("@/api/files", () => ({ uploadFile: vi.fn() }));

const nestedReply: Message = {
	id: "reply-2",
	channel_id: "ch-1",
	sender_id: "user-2",
	content_text: "nested reply target",
	thread_id: "root-1",
	reply_count: 0,
	created_at: "2026-08-26T00:00:00Z",
};

describe("Phase 151 direct-root composer", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		sendMessage.mockResolvedValue({ id: "sent-1" });
		useAuthStore.setState({
			user: {
				id: "composer-direct-root-user",
				email: "composer-direct-root@example.test",
				timezone: "UTC",
				language: "en",
				system_role: "member",
				onboarding_completed: true,
			},
		});
		useMessengerStore.getState().reset();
		useMessengerStore.getState().setReplyTo(nestedReply);
	});

	afterEach(() => {
		useAuthStore.setState({ user: null });
	});

	it("normalizes reply-to-reply to the canonical root", async () => {
		renderWithProviders(<MessageComposer canSend canAlsoSendToChannel channelId="ch-1" />);
		fireEvent.change(screen.getByRole("textbox"), {
			target: { value: "direct reply" },
		});
		fireEvent.click(screen.getByRole("button", { name: /отправить/i }));

		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, payload] = sendMessage.mock.calls[0] as [
			string,
			{ thread_id?: string },
		];
		expect(payload.thread_id).toBe("root-1");
		expect(payload.thread_id).not.toBe("reply-2");
	});

	it("sends the explicit also-send flag only when selected", async () => {
		renderWithProviders(<MessageComposer canSend canAlsoSendToChannel channelId="ch-1" />);
		fireEvent.click(
			screen.getByRole("checkbox", { name: /также отправить в канал/i }),
		);
		fireEvent.change(screen.getByRole("textbox"), {
			target: { value: "visible in both places" },
		});
		fireEvent.click(screen.getByRole("button", { name: /отправить/i }));

		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, payload] = sendMessage.mock.calls[0] as [
			string,
			{ also_send_to_channel?: boolean },
		];
		expect(payload.also_send_to_channel).toBe(true);
	});

	it("omits the also-send field when the option is not selected", async () => {
		renderWithProviders(<MessageComposer canSend canAlsoSendToChannel channelId="ch-1" />);
		fireEvent.change(screen.getByRole("textbox"), {
			target: { value: "thread only" },
		});
		fireEvent.click(screen.getByRole("button", { name: /отправить/i }));

		await waitFor(() => expect(sendMessage).toHaveBeenCalledTimes(1));
		const [, payload] = sendMessage.mock.calls[0] as [
			string,
			{ also_send_to_channel?: boolean },
		];
		expect(payload).not.toHaveProperty("also_send_to_channel");
	});

	it("blocks click and keyboard submission without an explicit capability", () => {
		renderWithProviders(
			<MessageComposer canSend={false} channelId="ch-1" />,
		);
		const textbox = screen.getByRole("textbox");
		fireEvent.change(textbox, { target: { value: "must not send" } });
		fireEvent.keyDown(textbox, { key: "Enter", shiftKey: false });
		fireEvent.click(screen.getByRole("button", { name: /отправить/i }));
		expect(sendMessage).not.toHaveBeenCalled();
	});
});
