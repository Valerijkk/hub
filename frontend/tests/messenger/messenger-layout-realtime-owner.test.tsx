import { fireEvent, render, screen } from "@testing-library/react";
import { type ComponentType, type ReactNode, useEffect } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const realtimeLifecycle = vi.hoisted(() => ({
	active: 0,
	cleanups: 0,
	mounts: 0,
}));

vi.mock("@/hooks/use-messenger-channels-realtime", () => ({
	useMessengerChannelsRealtime: () => {
		useEffect(() => {
			realtimeLifecycle.active += 1;
			realtimeLifecycle.mounts += 1;
			return () => {
				realtimeLifecycle.active -= 1;
				realtimeLifecycle.cleanups += 1;
			};
		}, []);
	},
}));

vi.mock("@tanstack/react-router", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-router")>();
	const { useChannelDrawer } = await import(
		"@/components/messenger/channel-drawer-context"
	);
	return {
		...actual,
		createFileRoute: () => (options: { component: ComponentType }) => ({
			options,
		}),
		Outlet: () => {
			const { setOpen } = useChannelDrawer();
			return (
				<button type="button" onClick={() => setOpen(true)}>
					Open mobile channel drawer
				</button>
			);
		},
	};
});

vi.mock("@/components/messenger/channel-list", () => ({
	ChannelList: ({ onChannelSelect }: { onChannelSelect?: () => void }) =>
		onChannelSelect ? (
			<button type="button" onClick={onChannelSelect}>
				Select mobile channel
			</button>
		) : (
			<div>Desktop channel list</div>
		),
}));

vi.mock("@/components/route-error-boundary", () => ({
	RouteErrorBoundary: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

vi.mock("@/components/ui/sheet", () => ({
	Sheet: ({ open, children }: { open: boolean; children: ReactNode }) =>
		open ? <>{children}</> : null,
	SheetContent: ({ children }: { children: ReactNode }) => <>{children}</>,
	SheetTitle: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

import { Route } from "@/routes/_authenticated/messenger";

describe("MessengerLayout realtime owner", () => {
	beforeEach(() => {
		realtimeLifecycle.active = 0;
		realtimeLifecycle.cleanups = 0;
		realtimeLifecycle.mounts = 0;
	});

	it("keeps one realtime owner alive while the mobile Sheet opens and closes", () => {
		const MessengerLayout = Route.options.component;
		if (!MessengerLayout) {
			throw new Error("messenger route must define a layout component");
		}
		const view = render(<MessengerLayout />);

		expect(screen.getByText("Desktop channel list")).toBeInTheDocument();
		expect(realtimeLifecycle).toEqual({ active: 1, cleanups: 0, mounts: 1 });

		fireEvent.click(
			screen.getByRole("button", { name: "Open mobile channel drawer" }),
		);
		expect(
			screen.getByRole("button", { name: "Select mobile channel" }),
		).toBeInTheDocument();
		expect(realtimeLifecycle).toEqual({ active: 1, cleanups: 0, mounts: 1 });

		fireEvent.click(
			screen.getByRole("button", { name: "Select mobile channel" }),
		);
		expect(
			screen.queryByRole("button", { name: "Select mobile channel" }),
		).not.toBeInTheDocument();
		expect(realtimeLifecycle).toEqual({ active: 1, cleanups: 0, mounts: 1 });

		view.unmount();
		expect(realtimeLifecycle).toEqual({ active: 0, cleanups: 1, mounts: 1 });
	});
});