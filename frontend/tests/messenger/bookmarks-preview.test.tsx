import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { BookmarksPanel } from "@/components/messenger/bookmarks-panel";
import { renderWithProviders } from "../support/render";

const CHANNEL_UUID = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee";
const MESSAGE_UUID = "11111111-2222-4333-8444-555555555555";

vi.mock("@tanstack/react-query", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-query")>(
		"@tanstack/react-query",
	);
	return {
		...actual,
		useQuery: vi.fn(),
	};
});

vi.mock("@tanstack/react-router", () => ({
	useNavigate: () => vi.fn(),
}));

import { useQuery } from "@tanstack/react-query";

describe("BookmarksPanel preview labels (A3W9-2-001 / TASK-173)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows channel_name + content_preview, not raw UUIDs as primary label", async () => {
		vi.mocked(useQuery).mockReturnValue({
			data: [
				{
					id: "bm-1",
					message_id: MESSAGE_UUID,
					channel_id: CHANNEL_UUID,
					channel_name: "general",
					content_preview: "Привет, команда!",
					created_at: "2026-07-29T10:00:00Z",
				},
			],
			isLoading: false,
			isError: false,
			error: null,
			refetch: vi.fn(),
		} as unknown as ReturnType<typeof useQuery>);

		renderWithProviders(<BookmarksPanel />);

		await waitFor(() => {
			expect(screen.getByText("general")).toBeInTheDocument();
		});
		expect(screen.getByText("Привет, команда!")).toBeInTheDocument();

		const openBtn = screen.getByRole("button", {
			name: /general:\s*Привет, команда!/i,
		});
		expect(openBtn).toBeInTheDocument();
		expect(openBtn.textContent).not.toContain(CHANNEL_UUID);
		expect(openBtn.textContent).not.toContain(MESSAGE_UUID);
		expect(screen.queryByText(CHANNEL_UUID)).not.toBeInTheDocument();
		expect(screen.queryByText(MESSAGE_UUID)).not.toBeInTheDocument();
	});

	it("falls back to «Канал»/«Сообщение» when preview fields empty — still no UUID primary", async () => {
		vi.mocked(useQuery).mockReturnValue({
			data: [
				{
					id: "bm-2",
					message_id: MESSAGE_UUID,
					channel_id: CHANNEL_UUID,
					created_at: "2026-07-29T10:00:00Z",
				},
			],
			isLoading: false,
			isError: false,
			error: null,
			refetch: vi.fn(),
		} as unknown as ReturnType<typeof useQuery>);

		renderWithProviders(<BookmarksPanel />);

		await waitFor(() => {
			expect(
				screen.getByRole("button", { name: /канал:\s*сообщение/i }),
			).toBeInTheDocument();
		});
		expect(screen.queryByText(CHANNEL_UUID)).not.toBeInTheDocument();
		expect(screen.queryByText(MESSAGE_UUID)).not.toBeInTheDocument();
	});
});
