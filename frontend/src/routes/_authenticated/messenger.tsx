import { createFileRoute, Outlet } from "@tanstack/react-router";
import { ChannelList } from "@/components/messenger/channel-list";

/**
 * Лейаут-маршрут мессенджера.
 * Sidebar с каналами слева, контент (Outlet) справа.
 * Полная высота, без overflow-проблем.
 */
function MessengerLayout() {
	return (
		<div className="flex h-[calc(100vh-3.5rem)] -mx-4 -my-6 md:-mx-6 overflow-hidden">
			{/* Sidebar каналов — тёмный стиль, соответствует основному сайдбару */}
			<div className="w-72 shrink-0 border-r border-border/50 bg-sidebar text-sidebar-foreground">
				<ChannelList />
			</div>

			{/* Основной контент */}
			<div className="flex flex-1 flex-col overflow-hidden">
				<Outlet />
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_authenticated/messenger")({
	component: MessengerLayout,
});
