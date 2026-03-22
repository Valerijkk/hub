/**
 * Slack-style ссылка сайдбара с tooltip при свёрнутом режиме.
 *
 * Активный элемент: 3px blue bar слева + подсветка #35373B.
 * Hover: rgba(255,255,255,0.05). Иконки 20px, текст 15px, высота 32px.
 */
import { Link, useMatchRoute } from "@tanstack/react-router";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";

/** Цвета для иконок навигации */
export const NAV_ICON_COLORS: Record<string, string> = {
	"/dashboard": "text-amber-400",
	"/messenger": "text-emerald-400",
	"/calls": "text-violet-400",
	"/projects": "text-blue-400",
	"/settings/profile": "text-slate-400",
	"/trash": "text-slate-500",
};

/** Элемент навигации сайдбара -- Slack-style с left accent bar. */
export function SidebarLink({
	to,
	label,
	icon,
	collapsed,
	params,
	indent,
	unread,
}: {
	to: string;
	label: string;
	icon: React.ReactNode;
	collapsed: boolean;
	params?: Record<string, string>;
	/** Отступ для подменю. */
	indent?: boolean;
	/** Непрочитанные -- жирный текст. */
	unread?: boolean;
}) {
	const matchRoute = useMatchRoute();
	const isActive = matchRoute({ to, fuzzy: true });

	const link = (
		<Link
			to={to}
			params={params}
			className={cn(
				"group relative flex h-8 items-center gap-[10px] rounded-md px-3 text-[15px] transition-all duration-150",
				"text-[#D1D2D3] hover:bg-white/[0.05] hover:text-white",
				isActive && "bg-[#35373B] text-white",
				collapsed && "justify-center px-0",
				indent && !collapsed && "pl-[42px] text-[13px]",
				unread && "font-bold",
			)}
		>
			{/* Активный индикатор — 3px blue bar слева */}
			{isActive && (
				<div className="absolute left-0 top-1/2 -translate-y-1/2 h-5 w-[3px] rounded-r-full bg-blue-500" />
			)}
			<span className="flex h-5 w-5 shrink-0 items-center justify-center">
				{icon}
			</span>
			{!collapsed && (
				<span className="truncate leading-tight">{label}</span>
			)}
		</Link>
	);

	if (collapsed) {
		return (
			<Tooltip>
				<TooltipTrigger asChild>
					<div className="relative flex justify-center">{link}</div>
				</TooltipTrigger>
				<TooltipContent side="right">{label}</TooltipContent>
			</Tooltip>
		);
	}

	return <div className="relative">{link}</div>;
}
