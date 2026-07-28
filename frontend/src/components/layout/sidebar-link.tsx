/**
 * Ссылка сайдбара v2.0 с accent glow active state.
 *
 * Активный элемент: accent-colored bar слева с glow shadow + gradient overlay.
 * Hover: var(--hover). Иконки 18px, текст 13.5px, gap 11px.
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
	"/calendar": "text-orange-400",
	"/crm": "text-teal-400",
	"/hr": "text-pink-400",
	"/marketplace": "text-lime-400",
	"/bases": "text-cyan-400",
	"/messenger": "text-emerald-400",
	"/calls": "text-violet-400",
	"/projects": "text-blue-400",
	"/intranet": "text-rose-400",
	"/intranet/directory": "text-sky-400",
	"/help": "text-amber-300",
	"/settings/profile": "text-slate-400",
	"/trash": "text-slate-500",
};

/** Элемент навигации сайдбара -- accent bar + glow active state. */
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
			aria-label={
				collapsed
					? badge != null && badge > 0
						? `${label}, ${badge} непрочитанных`
						: label
					: undefined
			}
			aria-current={isActive ? "page" : undefined}
			className={cn(
				"group relative flex items-center gap-[11px] rounded-[var(--r2)] px-3 py-[9px] text-[13.5px] font-medium transition-colors duration-200",
				"text-sidebar-muted hover:bg-[var(--hover)] hover:text-sidebar-foreground",
				isActive && "bg-[var(--active)] text-white font-semibold",
				collapsed && "justify-center px-0",
				indent && !collapsed && "pl-[42px] text-[13px]",
				unread && "font-bold",
			)}
		>
			{/* Акцентная полоска активного пункта */}
			{isActive && !collapsed && (
				<span
					className="absolute -left-[3px] top-[6px] bottom-[6px] w-[3px] rounded-sm bg-primary shadow-[0_0_12px_var(--primary),0_0_30px_rgba(91,156,245,.15)]"
					aria-hidden="true"
				/>
			)}
			{/* Градиентный overlay активного пункта */}
			{isActive && !collapsed && (
				<span
					className="pointer-events-none absolute inset-y-0 left-0 w-[60px] rounded-[var(--r2)] bg-gradient-to-r from-primary/[0.06] to-transparent"
					aria-hidden="true"
				/>
			)}
			<span
				className={cn(
					"relative flex h-5 w-5 shrink-0 items-center justify-center [&_svg]:w-[18px] [&_svg]:h-[18px] [&_svg]:opacity-45",
					isActive &&
						"[&_svg]:opacity-100 [&_svg]:drop-shadow-[0_0_4px_rgba(91,156,245,.3)]",
				)}
			>
				{icon}
				{collapsed && badge != null && badge > 0 && (
					<span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-red-500" />
				)}
			</span>
			{!collapsed && <span className="truncate leading-tight">{label}</span>}
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
