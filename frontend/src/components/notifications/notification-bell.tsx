import { Bell } from "lucide-react";
import { useEffect, useRef } from "react";
import type { NotificationDTO } from "@/api/notifications";
import { Button } from "@/components/ui/button";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { useNotificationStore } from "@/stores/notification-store";
import { useWsStore } from "@/stores/ws-store";
import { NotificationPanel } from "./notification-panel";

/**
 * NotificationBell — иконка колокольчика с badge непрочитанных уведомлений.
 * При клике открывает панель уведомлений в Popover.
 * Подписывается на WS-события notification.new для real-time обновлений.
 */
export function NotificationBell() {
	const unreadCount = useNotificationStore((s) => s.unreadCount);
	const fetchUnreadCount = useNotificationStore((s) => s.fetchUnreadCount);
	const fetchNotifications = useNotificationStore((s) => s.fetchNotifications);
	const addNotification = useNotificationStore((s) => s.addNotification);
	const addListener = useWsStore((s) => s.addListener);
	const removeListener = useWsStore((s) => s.removeListener);

	// useRef для WS обработчика чтобы избежать stale closures.
	const addNotificationRef = useRef(addNotification);
	addNotificationRef.current = addNotification;

	useEffect(() => {
		void fetchUnreadCount();
	}, [fetchUnreadCount]);

	useEffect(() => {
		const handler = (event: string, payload: unknown) => {
			if (event === "notification.new" && payload) {
				addNotificationRef.current(payload as NotificationDTO);
			}
		};

		addListener(handler);
		return () => {
			removeListener(handler);
		};
	}, [addListener, removeListener]);

	return (
		<Popover>
			<PopoverTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="relative h-8 w-8"
					onClick={() => void fetchNotifications()}
					aria-label="Уведомления"
				>
					<Bell className="h-4 w-4" />
					{unreadCount > 0 && (
						<span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-medium text-destructive-foreground">
							{unreadCount > 99 ? "99+" : unreadCount}
						</span>
					)}
				</Button>
			</PopoverTrigger>
			<PopoverContent align="end" className="w-96 p-0" sideOffset={8}>
				<NotificationPanel />
			</PopoverContent>
		</Popover>
	);
}
