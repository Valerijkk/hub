import { screen, waitFor } from "@testing-library/react";
import { createRef } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as messengerApi from "@/api/messenger";
import { MentionAutocomplete } from "@/components/messenger/mention-autocomplete";
import { renderWithProviders } from "../../render";

describe("MentionAutocomplete a11y (A10-010)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
		vi.spyOn(messengerApi, "getChannelMembers").mockResolvedValue([
			{
				id: "m1",
				user_id: "u1",
				user_name: "Alice",
				role: "member",
				joined_at: "2026-07-01T00:00:00Z",
			},
			{
				id: "m2",
				user_id: "u2",
				user_name: "Bob",
				role: "admin",
				joined_at: "2026-07-01T00:00:00Z",
			},
		]);
	});

	it("exposes listbox options with aria-selected", async () => {
		const textareaRef = createRef<HTMLTextAreaElement>();
		renderWithProviders(
			<>
				<textarea ref={textareaRef} defaultValue="@A" />
				<MentionAutocomplete
					channelId="ch-1"
					text="@A"
					cursorPosition={2}
					textareaRef={textareaRef}
					onSelect={vi.fn()}
				/>
			</>,
		);

		const listbox = await screen.findByRole("listbox", {
			name: /участники для упоминания/i,
		});
		expect(listbox).toBeInTheDocument();

		await waitFor(() => {
			const options = screen.getAllByRole("option");
			expect(options.length).toBeGreaterThan(0);
			expect(options[0]).toHaveAttribute("aria-selected", "true");
		});
	});
});
