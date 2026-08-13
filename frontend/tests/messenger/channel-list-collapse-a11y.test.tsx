import { fireEvent, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel } from "@/api/messenger";
import { ChannelList } from "@/components/messenger/channel-list";
import { renderWithProviders } from "../support/render";

vi.mock("@/hooks/use-channels", () => ({
	useChannels: vi.fn(),
}));

vi.mock("@/hooks/use-unread", () => ({
	useUnread: vi.fn(),
}));

vi.mock("@/hooks/use-presence", () => ({
	usePresence: () => ({ getStatus: () => "offline" }),
}));

vi.mock("@tanstack/react-router", () => ({
	useMatchRoute: () => () => false,
	Link: ({
		children,
		...props
	}: {
		children: React.ReactNode;
		to?: string;
		params?: Record<string, string>;
		className?: string;
		onClick?: (e: React.MouseEvent) => void;
		onContextMenu?: (e: React.MouseEvent) => void;
	}) => (
		<a href={props.to ?? "#"} className={props.className} onClick={props.onClick}>
			{children}
		</a>
	),
}));

vi.mock("@/components/messenger/new-dm-dialog", () => ({
	NewDmDialog: () => (
		<button type="button" aria-label="Новый DM">
			DM
		</button>
	),
}));

vi.mock("@/components/messenger/create-channel-dialog", () => ({
	CreateChannelDialog: () => null,
	CreateChannelDialogContent: () => null,
}));

vi.mock("@/components/messenger/bookmarks-panel", () => ({
	BookmarksPanel: () => null,
}));

import { useChannels } from "@/hooks/use-channels";

function makeChannel(overrides: Partial<Channel> = {}): Channel {
	return {
		id: "00000000-0000-4000-8000-000000000001",
		channel_type: "public",
		name: "General",
		is_readonly: false,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-07-01T00:00:00Z",
		updated_at: "2026-07-01T00:00:00Z",
		...overrides,
	};
}

/**
 * A3W9-3-001: section collapse — aria-expanded + aria-controls + panel id.
 */
describe("ChannelList section collapse a11y (A3W9-3-001)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("collapse button has aria-expanded and aria-controls pointing at panel id", () => {
		vi.mocked(useChannels).mockReturnValue({
			data: [makeChannel()],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		renderWithProviders(<ChannelList />);

		const toggle = screen.getByRole("button", { name: /каналы/i });
		expect(toggle).toHaveAttribute("aria-expanded", "true");
		const panelId = toggle.getAttribute("aria-controls");
		expect(panelId).toBeTruthy();
		const panel = document.getElementById(panelId!);
		expect(panel).not.toBeNull();
		expect(panel).not.toHaveAttribute("hidden");

		fireEvent.click(toggle);
		expect(toggle).toHaveAttribute("aria-expanded", "false");
		expect(panel).toHaveAttribute("hidden");
	});
});
