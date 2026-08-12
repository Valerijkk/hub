import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel } from "@/api/messenger";
import { ChannelList } from "@/components/messenger/channel-list";
import { renderWithProviders } from "../../render";

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

describe("ChannelList mute/star a11y (A10-013)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("star button exposes aria-label matching title pattern", () => {
		vi.mocked(useChannels).mockReturnValue({
			data: [makeChannel({ is_starred: false })],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		renderWithProviders(<ChannelList />);

		expect(
			screen.getByRole("button", { name: "Закрепить канал" }),
		).toBeInTheDocument();
	});

	it("starred channel uses unstar aria-label", () => {
		vi.mocked(useChannels).mockReturnValue({
			data: [makeChannel({ is_starred: true })],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		renderWithProviders(<ChannelList />);

		expect(
			screen.getByRole("button", { name: "Убрать из закреплённых" }),
		).toBeInTheDocument();
	});

	it("muted channel exposes aria-label on mute indicator", () => {
		const future = new Date(Date.now() + 3_600_000).toISOString();
		vi.mocked(useChannels).mockReturnValue({
			data: [makeChannel({ muted_until: future })],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		renderWithProviders(<ChannelList />);

		expect(screen.getByRole("img", { name: "Канал замьючен" })).toBeInTheDocument();
	});
});
