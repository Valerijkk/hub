import { Bell } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import {
	useNotificationRealtime,
	useUnreadNotificationCount,
} from "@/hooks/use-notifications";
import { NotificationPanel } from "./notification-panel";

/**
 * NotificationBell — иконка колокольчика с badge непрочитанных.
 * Unread + list — TanStack Query (A10-007); WS — realtime hook.
 */
export function NotificationBell() {
	const [open, setOpen] = useState(false);
	const unreadQuery = useUnreadNotificationCount();
	const unreadCount = unreadQuery.data ?? 0;
	useNotificationRealtime();

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					variant="ghost"
					size="icon"
					className="relative h-7 w-7 text-[#D1D2D3]/40 hover:text-white hover:bg-white/[0.05]"
					aria-label="Уведомления"
					type="button"
				>
					<Bell className="h-4 w-4" />
					{unreadCount > 0 && (
						<span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 animate-scale-in items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-medium text-white">
							{unreadCount > 99 ? "99+" : unreadCount}
						</span>
					)}
				</Button>
			</PopoverTrigger>
			<PopoverContent
				align="end"
				className="w-[380px] max-h-[500px] overflow-hidden rounded-2xl border-white/10 bg-popover p-0 shadow-[0_16px_48px_rgba(0,0,0,0.4)]"
				sideOffset={8}
			>
				<NotificationPanel enabled={open} />
			</PopoverContent>
		</Popover>
	);
}
