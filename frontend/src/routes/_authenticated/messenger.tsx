import { createFileRoute, Outlet } from "@tanstack/react-router";
import { ChannelList } from "@/components/messenger/channel-list";

/**
 * Лейаут-маршрут мессенджера.
 * Sidebar с каналами слева, контент (Outlet) справа.
 * Полная высота, без overflow-проблем.
 */
function MessengerLayout() {
	return (
		<div className="flex h-full -mx-4 -my-4 md:-mx-6 md:-my-6 overflow-hidden">
			{/* Sidebar каналов — растягивается на всю высоту */}
			<div className="hidden md:flex shrink-0">
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
