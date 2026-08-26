import { cleanup, screen, within } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { SidebarNav } from "@/components/layout/sidebar-nav";
import { useMessengerStore } from "@/stores/messenger-store";
import { renderWithProviders } from "../support/render";

vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
		<a href={to} {...props}>
			{children}
		</a>
	),
	useMatchRoute: () => () => false,
}));

type UnreadSnapshot = {
	channels: Record<string, number>;
	dms: Record<string, number>;
	mentions: Record<string, number>;
	threads: Record<string, number>;
	any_unread: boolean;
};

const emptyUnreadSnapshot: UnreadSnapshot = {
	channels: {},
	dms: {},
	mentions: {},
	threads: {},
	any_unread: false,
};

function setUnreadSnapshot(unreadSnapshot: UnreadSnapshot) {
	// Phase 151 replaces the legacy flat `unreadCounts` map with this server
	// snapshot. Keep the RED test type-safe before that source change lands.
	useMessengerStore.setState({ unreadSnapshot } as never);
}

function renderSidebar() {
	return renderWithProviders(
		<SidebarNav
			collapsed={false}
			t={(key) => (key === "nav.messenger" ? "Messenger" : key)}
		/>,
	);
}

describe("Phase 151 Messenger sidebar unread indicator", () => {
	afterEach(() => {
		cleanup();
		setUnreadSnapshot(emptyUnreadSnapshot);
	});

	it.each([
		["channel", { ...emptyUnreadSnapshot, channels: { "channel-1": 1 }, any_unread: true }],
		["direct message", { ...emptyUnreadSnapshot, dms: { "dm-1": 1 }, any_unread: true }],
		["mention", { ...emptyUnreadSnapshot, mentions: { "channel-1": 1 }, any_unread: true }],
		["thread", { ...emptyUnreadSnapshot, threads: { "thread-1": 1 }, any_unread: true }],
	] satisfies readonly [string, UnreadSnapshot][])(
		"shows one Messenger indicator when a %s category is unread",
		(_category, unreadSnapshot) => {
			setUnreadSnapshot(unreadSnapshot);
			renderSidebar();

			expect(screen.getAllByTestId("sidebar-unread-dot")).toHaveLength(1);
		},
	);

	it("does not sum overlapping unread categories into a numeric badge", () => {
		setUnreadSnapshot({
			channels: { "channel-1": 2 },
			dms: { "dm-1": 3 },
			mentions: { "channel-1": 2 },
			threads: { "thread-1": 4 },
			any_unread: true,
		});
		renderSidebar();

		const messengerLink = screen.getByRole("link", { name: "Messenger" });
		expect(within(messengerLink).getAllByTestId("sidebar-unread-dot")).toHaveLength(1);
		expect(messengerLink).not.toHaveTextContent("11");
	});

	it("does not show a Messenger indicator when every category is empty", () => {
		setUnreadSnapshot(emptyUnreadSnapshot);
		renderSidebar();

		expect(screen.queryByTestId("sidebar-unread-dot")).not.toBeInTheDocument();
	});
});
