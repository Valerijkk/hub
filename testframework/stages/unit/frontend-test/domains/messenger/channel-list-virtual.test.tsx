import { screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Channel } from "@/api/messenger";
import { ChannelList } from "@/components/messenger/channel-list";
import { renderWithProviders } from "../../render";

/** Mirrors `CHANNEL_VIRTUALIZE_THRESHOLD` in channel-list.tsx (A13-018). */
const CHANNEL_VIRTUALIZE_THRESHOLD = 40;

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

function makeChannel(i: number): Channel {
	return {
		id: `00000000-0000-4000-8000-${String(i).padStart(12, "0")}`,
		channel_type: "public",
		name: `Channel ${i}`,
		is_readonly: false,
		created_by: "user-1",
		is_starred: false,
		created_at: "2026-07-01T00:00:00Z",
		updated_at: "2026-07-01T00:00:00Z",
	};
}

/** jsdom reports 0 scroll height — stub so @tanstack/react-virtual mounts rows. */
function stubScrollMetrics() {
	Object.defineProperty(HTMLElement.prototype, "clientHeight", {
		configurable: true,
		get() {
			return 480;
		},
	});
	Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
		configurable: true,
		get() {
			return 52;
		},
	});
}

describe("ChannelList virtualization (A13-018)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		stubScrollMetrics();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("renders all channels when below virtualize threshold", () => {
		const channels = Array.from({ length: 5 }, (_, i) => makeChannel(i + 1));
		vi.mocked(useChannels).mockReturnValue({
			data: channels,
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		const { container } = renderWithProviders(<ChannelList />);

		expect(screen.getByText("Channel 1")).toBeInTheDocument();
		expect(screen.getByText("Channel 5")).toBeInTheDocument();
		expect(container.querySelectorAll("[data-index]")).toHaveLength(0);
	});

	it("mounts virtual path when flat rows >= CHANNEL_VIRTUALIZE_THRESHOLD", () => {
		// One section header ("Каналы") + N channels → N+1 flat rows.
		// Use N = threshold so rows.length === threshold + 1 > threshold.
		const channelCount = CHANNEL_VIRTUALIZE_THRESHOLD;
		const channels = Array.from({ length: channelCount }, (_, i) =>
			makeChannel(i + 1),
		);
		vi.mocked(useChannels).mockReturnValue({
			data: channels,
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useChannels>);

		const { container } = renderWithProviders(<ChannelList />);

		const list = screen.getByTestId("messenger-channel-list");
		expect(list).toBeInTheDocument();

		const virtualRows = container.querySelectorAll("[data-index]");
		expect(virtualRows.length).toBeGreaterThan(0);
		// Flat rows = 1 header + channelCount; DOM must mount fewer than that.
		expect(virtualRows.length).toBeLessThan(channelCount + 1);

		const spacer = Array.from(
			container.querySelectorAll<HTMLElement>("div.relative.w-full"),
		).find((el) => el.style.height && Number.parseFloat(el.style.height) > 0);
		expect(spacer).toBeTruthy();
		expect(Number.parseFloat(spacer?.style.height ?? "0")).toBeGreaterThan(
			channelCount * 40,
		);

		expect(
			within(list).queryByText(`Channel ${channelCount}`),
		).not.toBeInTheDocument();
		expect(screen.getByText("Channel 1")).toBeInTheDocument();
	});
});
