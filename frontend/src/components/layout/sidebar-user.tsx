/**
 * Профиль пользователя в нижней части сайдбара -- Slack-style.
 * Аватар 28px + имя + шестерёнка настроек. Зелёная точка online.
 */
import { Link } from "@tanstack/react-router";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

/** Профиль пользователя внизу сайдбара с аватаром, именем и настройками. */
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
		<div className="relative flex-shrink-0">
			{user?.avatar_url ? (
				<img
					src={user.avatar_url}
					alt={displayName}
					className="h-7 w-7 rounded-md object-cover"
				/>
			) : (
				<div className="flex h-7 w-7 items-center justify-center rounded-md bg-white/[0.08] text-xs font-semibold text-white">
					{initials}
				</div>
			)}
			{/* Online-статус */}
			<span className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-emerald-400 ring-2 ring-[#19171D]" />
		</div>
	);

	if (collapsed) {
		return (
			<div className="shrink-0 border-t border-white/[0.06] px-3 py-3">
				<Tooltip>
					<TooltipTrigger asChild>
						<button
							type="button"
							onClick={toggle}
							className="mx-auto flex items-center justify-center"
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
		<div className="shrink-0 border-t border-white/[0.06] px-3 py-2.5">
			<div className="flex items-center gap-2.5">
				{avatarBlock}
				<div className="flex-1 min-w-0">
					<p className="truncate text-[13px] font-medium text-white leading-tight">
						{displayName}
					</p>
					{user?.position && (
						<p className="truncate text-[11px] text-[#D1D2D3]/50 leading-tight">
							{user.position}
						</p>
					)}
				</div>
				{/* Настройки */}
				<Tooltip>
					<TooltipTrigger asChild>
						<Link
							to="/settings/profile"
							className={cn(
								"flex h-7 w-7 shrink-0 items-center justify-center rounded-md",
								"text-[#D1D2D3]/40 hover:text-white hover:bg-white/[0.05] transition-all duration-150",
							)}
						>
							<svg
								width="16"
								height="16"
								viewBox="0 0 16 16"
								fill="none"
								stroke="currentColor"
								strokeWidth="1.5"
								strokeLinecap="round"
								strokeLinejoin="round"
							>
								<circle cx="8" cy="8" r="2.5" />
								<path d="M13.5 8a5.5 5.5 0 01-.3 1.8l1.1.6a.5.5 0 01.2.7l-1 1.7a.5.5 0 01-.6.2l-1.1-.6a5.5 5.5 0 01-1.5.9v1.2a.5.5 0 01-.5.5h-2a.5.5 0 01-.5-.5v-1.2a5.5 5.5 0 01-1.5-.9l-1.1.6a.5.5 0 01-.6-.2l-1-1.7a.5.5 0 01.2-.7l1.1-.6A5.5 5.5 0 012.5 8c0-.6.1-1.2.3-1.8l-1.1-.6a.5.5 0 01-.2-.7l1-1.7a.5.5 0 01.6-.2l1.1.6a5.5 5.5 0 011.5-.9V1.5a.5.5 0 01.5-.5h2a.5.5 0 01.5.5v1.2a5.5 5.5 0 011.5.9l1.1-.6a.5.5 0 01.6.2l1 1.7a.5.5 0 01-.2.7l-1.1.6c.2.6.3 1.2.3 1.8z" />
							</svg>
						</Link>
					</TooltipTrigger>
					<TooltipContent side="top">Настройки</TooltipContent>
				</Tooltip>
			</div>
		</div>
	);
}

