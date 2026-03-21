import { createFileRoute, Outlet } from "@tanstack/react-router";
import { ChannelList } from "@/components/messenger/channel-list";

/**
 * Лейаут-маршрут мессенджера.
 * Sidebar с каналами слева, контент (Outlet) справа.
 */
function MessengerLayout() {
  return (
    <div className="flex h-full -m-6 overflow-hidden">
      {/* Sidebar каналов */}
      <div className="w-72 shrink-0 border-r bg-muted/30">
        <ChannelList />
      </div>

      {/* Основной контент */}
      <div className="flex flex-1 overflow-hidden">
        <Outlet />
      </div>
    </div>
  );
}

export const Route = createFileRoute("/_authenticated/messenger")({
  component: MessengerLayout,
});
