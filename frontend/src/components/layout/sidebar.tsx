/**
 * Боковая панель навигации — всегда тёмная, минимальный дизайн.
 *
 * Лого Hub с градиентным значком, проекты с цветными точками,
 * навигация с тонкими hover-эффектами и активными индикаторами.
 * Сворачивается до иконок. На мобильных — overlay.
 */
import { Link, useMatchRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { listProjects } from "@/api/projects";
import { useSidebarStore } from "@/stores/sidebar-store";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Цвета для иконок навигации */
const NAV_ICON_COLORS: Record<string, string> = {
  "/messenger": "text-emerald-400",
  "/projects": "text-blue-400",
  "/settings/profile": "text-slate-400",
  "/trash": "text-slate-500",
};

function SidebarLink({
  to,
  label,
  icon,
  collapsed,
}: {
  to: string;
  label: string;
  icon: React.ReactNode;
  collapsed: boolean;
}) {
  const matchRoute = useMatchRoute();
  const isActive = matchRoute({ to, fuzzy: true });

  const link = (
    <Link
      to={to}
      className={cn(
        "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-150",
        "text-slate-400 hover:text-white hover:bg-white/[0.06]",
        isActive && "text-white bg-white/[0.08]",
        collapsed && "justify-center px-2"
      )}
    >
      {/* Активный индикатор — тонкая полоска слева */}
      {isActive && (
        <div className="absolute left-0 h-5 w-[2px] rounded-r-full bg-primary" />
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

export function Sidebar() {
  const { collapsed, toggle } = useSidebarStore();
  const { t } = useTranslation("common");

  const { data: projects } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  return (
    <aside
      className={cn(
        "flex h-screen flex-col border-r border-white/[0.06] transition-all duration-200",
        "bg-[#0C1222] text-slate-300",
        collapsed ? "w-[68px]" : "w-[260px]"
      )}
    >
      {/* Логотип и кнопка сворачивания */}
      <div className="flex h-14 items-center justify-between px-4">
        {!collapsed && (
          <Link to="/projects" className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <span className="text-xs font-bold text-white">H</span>
            </div>
            <span className="text-lg font-bold text-white tracking-tight">Hub</span>
          </Link>
        )}
        {collapsed && (
          <Link to="/projects" className="mx-auto">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <span className="text-xs font-bold text-white">H</span>
            </div>
          </Link>
        )}
        {!collapsed && (
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            className="h-7 w-7 text-slate-500 hover:text-white hover:bg-white/[0.06] rounded-md"
            aria-label={collapsed ? t("sidebar.expand") : t("sidebar.collapse")}
          >
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M10 12l-4-4 4-4" /></svg>
          </Button>
        )}
      </div>

      {/* Разделитель */}
      <div className="mx-4 h-px bg-white/[0.06]" />

      {/* Секция проектов */}
      <div className="flex-1 overflow-y-auto px-3 py-3">
        <div className="mb-2 flex items-center justify-between px-2">
          {!collapsed && (
            <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">
              {t("nav.projects")}
            </span>
          )}
          <Tooltip>
            <TooltipTrigger asChild>
              <Link
                to="/projects/new"
                className="inline-flex h-5 w-5 items-center justify-center rounded-md text-slate-500 hover:text-white hover:bg-white/[0.06] transition-colors"
              >
                <svg width="12" height="12" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M7 1v12M1 7h12" /></svg>
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right">{t("sidebar.createProject")}</TooltipContent>
          </Tooltip>
        </div>

        <div className="space-y-0.5">
          {projects?.map((project) => (
            <SidebarLink
              key={project.id}
              to={`/projects/${project.id}`}
              label={project.name}
              collapsed={collapsed}
              icon={
                <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-blue-500/20 text-[10px] font-bold text-blue-400">
                  {project.key.charAt(0)}
                </span>
              }
            />
          ))}
        </div>

        {projects?.length === 0 && !collapsed && (
          <p className="px-3 py-2 text-xs text-slate-600">
            {t("sidebar.noProjects")}
          </p>
        )}
      </div>

      {/* Разделитель */}
      <div className="mx-4 h-px bg-white/[0.06]" />

      {/* Нижняя навигация */}
      <div className="space-y-0.5 px-3 py-3">
        <SidebarLink
          to="/messenger"
          label={t("nav.messenger")}
          collapsed={collapsed}
          icon={
            <svg className={NAV_ICON_COLORS["/messenger"]} width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M2.5 12.5l1.5-4L2 4a1 1 0 01.9-1.4h10.2A1 1 0 0114 4l-2 4.5 1.5 4a1 1 0 01-.9 1.4H2.4a1 1 0 01-.9-1.4z" /><path d="M6 7h4M6 9.5h2.5" /></svg>
          }
        />
        <SidebarLink
          to="/projects"
          label={t("nav.allProjects")}
          collapsed={collapsed}
          icon={
            <svg className={NAV_ICON_COLORS["/projects"]} width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="2" width="5" height="5" rx="1" /><rect x="9" y="2" width="5" height="5" rx="1" /><rect x="2" y="9" width="5" height="5" rx="1" /><rect x="9" y="9" width="5" height="5" rx="1" /></svg>
          }
        />
        <SidebarLink
          to="/settings/profile"
          label={t("nav.settings")}
          collapsed={collapsed}
          icon={
            <svg className={NAV_ICON_COLORS["/settings/profile"]} width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="8" cy="8" r="2.5" /><path d="M13.5 8a5.5 5.5 0 01-.3 1.8l1.1.6a.5.5 0 01.2.7l-1 1.7a.5.5 0 01-.6.2l-1.1-.6a5.5 5.5 0 01-1.5.9v1.2a.5.5 0 01-.5.5h-2a.5.5 0 01-.5-.5v-1.2a5.5 5.5 0 01-1.5-.9l-1.1.6a.5.5 0 01-.6-.2l-1-1.7a.5.5 0 01.2-.7l1.1-.6A5.5 5.5 0 012.5 8c0-.6.1-1.2.3-1.8l-1.1-.6a.5.5 0 01-.2-.7l1-1.7a.5.5 0 01.6-.2l1.1.6a5.5 5.5 0 011.5-.9V1.5a.5.5 0 01.5-.5h2a.5.5 0 01.5.5v1.2a5.5 5.5 0 011.5.9l1.1-.6a.5.5 0 01.6.2l1 1.7a.5.5 0 01-.2.7l-1.1.6c.2.6.3 1.2.3 1.8z" /></svg>
          }
        />
        <SidebarLink
          to="/trash"
          label={t("nav.trash")}
          collapsed={collapsed}
          icon={
            <svg className={NAV_ICON_COLORS["/trash"]} width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"><path d="M2 4h12M5.3 4V2.7a1.3 1.3 0 011.4-1.4h2.6a1.3 1.3 0 011.4 1.4V4M12.7 4v9.3a1.3 1.3 0 01-1.4 1.4H4.7a1.3 1.3 0 01-1.4-1.4V4" /></svg>
          }
        />
      </div>

      {/* Кнопка развернуть — в свёрнутом состоянии */}
      {collapsed && (
        <div className="px-3 pb-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            className="w-full h-7 text-slate-500 hover:text-white hover:bg-white/[0.06] rounded-md"
            aria-label={t("sidebar.expand")}
          >
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M6 12l4-4-4-4" /></svg>
          </Button>
        </div>
      )}
    </aside>
  );
}
