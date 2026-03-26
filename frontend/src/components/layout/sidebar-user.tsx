/**
 * Профиль пользователя в нижней части сайдбара.
 * Аватар (кликабельный → профиль) + имя + колокольчик уведомлений.
 */
import { Link } from "@tanstack/react-router";
import { NotificationBell } from "@/components/notifications/notification-bell";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthStore } from "@/stores/auth-store";

export function SidebarUserProfile({
	collapsed,
	toggle,
}: {
	collapsed: boolean;
	toggle: () => void;
}) {
	const user = useAuthStore((s) => s.user);

	const displayName = user?.display_name || user?.name || user?.email || "Пользователь";
	const initials = displayName.charAt(0).toUpperCase();

	const avatarBlock = (
		<Link to="/settings/profile" className="relative flex-shrink-0 group">
			{user?.avatar_url ? (
				<img
					src={user.avatar_url}
					alt={displayName}
					className="h-8 w-8 rounded-full object-cover group-hover:ring-1 group-hover:ring-white/20 transition-all"
				/>
			) : (
				<div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-xs font-semibold text-white group-hover:brightness-110 transition-all">
					{initials}
				</div>
			)}
			<span className="absolute bottom-0 right-0 h-[9px] w-[9px] rounded-full bg-emerald-400 ring-[1.5px] ring-[hsl(220,16%,8%)] animate-pulse-online" />
		</Link>
	);

	if (collapsed) {
		return (
			<div className="shrink-0 border-t border-white/[0.06] px-3 h-12 flex items-center justify-center">
				<Tooltip>
					<TooltipTrigger asChild>
						<button
							type="button"
							onClick={toggle}
							className="flex items-center justify-center"
							aria-label="Развернуть"
						>
							{avatarBlock}
						</button>
					</TooltipTrigger>
					<TooltipContent side="right">
						{displayName}
					</TooltipContent>
				</Tooltip>
			</div>
		);
	}

	return (
		<div className="shrink-0 border-t border-white/[0.06] px-3 h-14 flex items-center">
			<div className="flex items-center gap-2 w-full">
				{avatarBlock}
				<div className="flex-1 min-w-0">
					<p className="truncate text-[13px] font-medium text-white leading-tight">
						{displayName}
					</p>
					<div className="text-[11px] text-white/40">
						{user?.position || "Участник"} · <span className="text-emerald-400">Онлайн</span>
					</div>
				</div>
				<div className="flex items-center shrink-0">
					<NotificationBell />
				</div>
			</div>
		</div>
	);
}
