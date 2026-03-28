import { createFileRoute, Outlet } from "@tanstack/react-router";
import { ChannelList } from "@/components/messenger/channel-list";

/**
 * Лейаут-маршрут мессенджера.
 * Sidebar с каналами слева, контент (Outlet) справа.
 * Полная высота, без overflow-проблем.
 */
function MessengerLayout() {
	return (
		<div className="flex h-[calc(100vh)] -mx-4 -my-4 md:-mx-6 overflow-hidden">
			{/* Sidebar каналов — тёмный стиль, соответствует основному сайдбару */}
			<div className="hidden md:block w-[272px] shrink-0 border-r border-white/[0.06] bg-[hsl(220,13%,10%)] text-sidebar-foreground">
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
