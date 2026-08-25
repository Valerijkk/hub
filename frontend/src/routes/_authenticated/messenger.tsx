import { createFileRoute, Outlet } from "@tanstack/react-router";
import { useCallback, useMemo, useState } from "react";
import { ChannelDrawerContext } from "@/components/messenger/channel-drawer-context";
import { ChannelList } from "@/components/messenger/channel-list";
import { RouteErrorBoundary } from "@/components/route-error-boundary";
import { Sheet, SheetContent, SheetTitle } from "@/components/ui/sheet";
import { useMessengerChannelsRealtime } from "@/hooks/use-messenger-channels-realtime";

/**
 * Лейаут-маршрут мессенджера.
 *
 * Desktop (>=md): sidebar с каналами слева, контент (Outlet) справа.
 * Mobile (<md): sidebar скрыт, открывается в `Sheet` слева через гамбургер
 * в шапке чата. При выборе канала sheet закрывается (см. `ChannelList`,
 * который читает `useChannelDrawer` через onSelect).
 */
function MessengerLayout() {
	useMessengerChannelsRealtime();
	const [drawerOpen, setDrawerOpen] = useState(false);

	const drawerCtx = useMemo(
		() => ({ open: drawerOpen, setOpen: setDrawerOpen }),
		[drawerOpen],
	);

	/** Закрыть drawer при выборе канала на мобилке. */
	const handleChannelSelect = useCallback(() => {
		setDrawerOpen(false);
	}, []);

	return (
		<ChannelDrawerContext.Provider value={drawerCtx}>
			<div className="flex h-full -mx-4 -my-4 md:-mx-6 md:-my-6 overflow-hidden">
				{/* Desktop sidebar на всю высоту */}
				<div className="hidden md:flex shrink-0">
					<ChannelList />
				</div>

				{/* Mobile drawer списка каналов */}
				<Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
					<SheetContent side="left" aria-label="Список каналов" className="p-0">
						<SheetTitle className="sr-only">Список каналов</SheetTitle>
						<ChannelList onChannelSelect={handleChannelSelect} />
					</SheetContent>
				</Sheet>

				{/* Основной контент */}
				<div className="flex flex-1 flex-col overflow-hidden">
					<RouteErrorBoundary module="Мессенджер">
						<Outlet />
					</RouteErrorBoundary>
				</div>
			</div>
		</ChannelDrawerContext.Provider>
	);
}

export const Route = createFileRoute("/_authenticated/messenger")({
	component: MessengerLayout,
});
