/**
 * Общий компонент ссылки сайдбара с tooltip при свёрнутом режиме.
 */
import { Link, useMatchRoute } from "@tanstack/react-router";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Цвета для иконок навигации */
export const NAV_ICON_COLORS: Record<string, string> = {
  "/dashboard": "text-amber-400",
  "/messenger": "text-emerald-400",
  "/calls": "text-violet-400",
  "/projects": "text-blue-400",
  "/settings/profile": "text-slate-400",
  "/trash": "text-slate-500",
};

/** Элемент навигации сайдбара с tooltip при свёрнутом режиме и индикатором активности. */
export function SidebarLink({
  to,
  label,
  icon,
  collapsed,
  params,
  indent,
}: {
  to: string;
  label: string;
  icon: React.ReactNode;
  collapsed: boolean;
  params?: Record<string, string>;
  /** Отступ для подменю. */
  indent?: boolean;
}) {
  const matchRoute = useMatchRoute();
  const isActive = matchRoute({ to, fuzzy: true });

  const link = (
    <Link
      to={to}
      params={params}
      className={cn(
        "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200",
        "text-slate-400 hover:text-white hover:bg-white/[0.07]",
        isActive && "text-white bg-white/[0.1] shadow-[0_0_12px_rgba(59,130,246,0.08)]",
        collapsed && "justify-center px-2",
        indent && !collapsed && "pl-9 text-[13px]",
      )}
    >
      {/* Активный индикатор — яркая полоска слева со свечением */}
      {isActive && (
        <div className="absolute left-0 h-5 w-[3px] rounded-r-full bg-primary shadow-[0_0_8px_rgba(59,130,246,0.5)]" />
      )}
      {icon}
      {!collapsed && <span className="truncate">{label}</span>}
    </Link>
  );

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="relative">{link}</div>
        </TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    );
  }

  return <div className="relative">{link}</div>;
}
