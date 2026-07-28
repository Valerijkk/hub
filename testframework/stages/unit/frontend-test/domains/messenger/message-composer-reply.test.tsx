import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Message } from "@/api/messenger";
import { MessageComposer } from "@/components/messenger/message-composer";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "@testframework/render";

const sendMessage = vi.fn();

vi.mock("@/api/messenger", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/messenger")>();
	return {
		...actual,
		sendMessage: (...args: unknown[]) => sendMessage(...args),
		sendTypingIndicator: vi.fn(),
	};
});

vi.mock("@/api/files", () => ({
	uploadFile: vi.fn(),
}));

const parent: Message = {
	id: "reply-parent-id",
	channel_id: "ch-1",
	sender_id: "u-2",
	content_text: "parent",
	reply_count: 0,
	created_at: "2026-07-28T12:00:00Z",
};

describe("FE-P1-MESSENGER · reply thread_id (W9-A3-002)", () => {
	beforeEach(() => {
		sendMessage.mockReset();
		sendMessage.mockResolvedValue({ id: "new-msg" });
		useMessengerStore.setState({ replyToMessage: parent });
	});

	it("sendMessage includes replyToMessage.id as thread_id when not in thread panel", async () => {
		renderWithProviders(<MessageComposer channelId="ch-1" />);

		const input = screen.getByRole("textbox");
		fireEvent.change(input, { target: { value: "Ответ на сообщение" } });
		fireEvent.click(screen.getByRole("button", { name: /отправ/i }));

		await waitFor(() => {
			expect(sendMessage).toHaveBeenCalled();
		});

		const [, payload] = sendMessage.mock.calls[0] as [
			string,
			{ thread_id?: string },
		];
		expect(payload.thread_id).toBe("reply-parent-id");
	});

	it("mutation: without reply id message would go without thread (guard)", async () => {
		useMessengerStore.setState({ replyToMessage: null });
		renderWithProviders(<MessageComposer channelId="ch-1" />);

		fireEvent.change(screen.getByRole("textbox"), {
			target: { value: "Обычное" },
		});
		fireEvent.click(screen.getByRole("button", { name: /отправ/i }));

		await waitFor(() => expect(sendMessage).toHaveBeenCalled());
		const [, payload] = sendMessage.mock.calls[0] as [
			string,
			{ thread_id?: string },
		];
		expect(payload.thread_id).toBeUndefined();
	});
});
