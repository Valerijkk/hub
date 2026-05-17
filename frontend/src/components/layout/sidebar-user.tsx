/**
 * Профиль пользователя в нижней части сайдбара v2.0.
 * 34px аватар с gradient glow, 10px online indicator, role badge pill.
 */
import { Link } from "@tanstack/react-router";
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
					className="h-[34px] w-[34px] rounded-full object-cover shadow-[0_2px_12px_var(--gB)] group-hover:scale-[1.06] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]"
				/>
			) : (
				<div className="flex h-[34px] w-[34px] items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-[11.5px] font-bold text-white shadow-[0_2px_12px_var(--gB)] group-hover:scale-[1.06] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]">
					{initials}
				</div>
			)}
			<span className="sb-on absolute bottom-0 right-0 h-[10px] w-[10px] rounded-full bg-emerald-400 border-[2.5px] border-sidebar shadow-[0_0_8px_var(--gG)]" />
		</Link>
	);

	if (collapsed) {
		return (
			<div className="shrink-0 border-t border-b1 px-3 h-12 flex items-center justify-center">
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
		<div className="shrink-0 border-t border-b1 px-[18px] h-14 flex items-center transition-colors duration-200 hover:bg-[var(--hover)] group">
			<div className="flex items-center gap-3 w-full">
				{avatarBlock}
				<div className="flex-1 min-w-0">
					<p className="truncate text-[13px] font-semibold text-t1 leading-tight">
						{displayName}
					</p>
					<div className="text-[10px] font-semibold text-t3">
						<span className="bg-[rgba(255,255,255,.04)] px-2 py-[2px] rounded-[var(--pill)]">
							{user?.position || "Участник"}
						</span>
						{" "}<span className="text-emerald-400">Онлайн</span>
					</div>
				</div>
			</div>
		</div>
	);
}
