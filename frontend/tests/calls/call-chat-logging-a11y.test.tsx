import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CallChatMessage } from "@/api/calls";
import { CallChat } from "@/components/calls/CallChat";
import { useCallStore } from "@/stores/call-store";
import { renderWithProviders } from "../support/render";

vi.mock("@/api/calls", () => ({
	listChatMessages: vi.fn(async () => [] as CallChatMessage[]),
	sendChatMessage: vi.fn(),
}));

describe("CallChat logging toggle a11y (A3W9-1-005)", () => {
	beforeEach(() => {
		useCallStore.setState({ chatLoggingEnabled: true });
	});

	it("exposes aria-label and aria-pressed for logging toggle", async () => {
		renderWithProviders(<CallChat callId="call-log-1" />);

		const toggle = await screen.findByRole("button", {
			name: /выключить логирование чата/i,
		});
		expect(toggle).toHaveAttribute("aria-pressed", "true");

		fireEvent.click(toggle);

		await waitFor(() => {
			const off = screen.getByRole("button", {
				name: /включить логирование чата/i,
			});
			expect(off).toHaveAttribute("aria-pressed", "false");
		});
	});
});
