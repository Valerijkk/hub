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
	badge,
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
	/** Числовой бейдж (кол-во непрочитанных, задач и т.д.). */
	badge?: number;
}) {
	const matchRoute = useMatchRoute();
	const isActive = matchRoute({ to, fuzzy: true });

	const link = (
		<Link
			to={to}
			params={params}
			className={cn(
				"group relative flex h-8 items-center gap-[10px] rounded-lg px-2 text-[13.5px] font-medium transition-all duration-150",
				"border-l-2 border-transparent",
				"text-[#979A9F] hover:bg-white/[0.05] hover:text-white",
				isActive && "bg-white/[0.07] text-white border-l-blue-500",
				collapsed && "justify-center px-0 border-l-0",
				indent && !collapsed && "pl-[42px] text-[13px]",
				unread && "font-bold",
			)}
		>
			<span className="relative flex h-5 w-5 shrink-0 items-center justify-center">
				{icon}
				{collapsed && badge != null && badge > 0 && (
					<span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-red-500" />
				)}
			</span>
			{!collapsed && (
				<span className="truncate leading-tight">{label}</span>
			)}
			{badge != null && badge > 0 && !collapsed && (
				<span className="ml-auto bg-red-500 text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center">
					{badge}
				</span>
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
