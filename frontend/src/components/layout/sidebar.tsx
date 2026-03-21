/**
 * Боковая панель навигации -- всегда тёмная, минимальный дизайн.
 *
 * Лого Hub с градиентным значком, проекты с цветными точками,
 * навигация с тонкими hover-эффектами и активными индикаторами.
 * Сворачивается до иконок. На мобильных -- overlay.
 * При раскрытии проекта показывает подменю: Доска, Задачи, TMS, Бэклог, Спринты.
 */
import { useState } from "react";
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
  "/dashboard": "text-amber-400",
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

/** Подменю модулей проекта (Доска, Задачи, TMS, Бэклог, Спринты). */
interface ProjectSubNavItem {
  to: string;
  label: string;
  icon: React.ReactNode;
}

/** Элементы подменю проекта. */
function getProjectSubNav(projectId: string): ProjectSubNavItem[] {
  return [
    {
      to: `/projects/${projectId}/board`,
      label: "Доска",
      icon: (
        <svg className="h-3.5 w-3.5 text-amber-400" fill="none" viewBox="0 0 16 16" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <rect x="1" y="2" width="4" height="12" rx="0.5" /><rect x="6" y="2" width="4" height="8" rx="0.5" /><rect x="11" y="2" width="4" height="5" rx="0.5" />
        </svg>
      ),
    },
    {
      to: `/projects/${projectId}/issues`,
      label: "Задачи",
      icon: (
        <svg className="h-3.5 w-3.5 text-blue-400" fill="none" viewBox="0 0 16 16" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="8" cy="8" r="6" /><path d="M5.5 8l2 2 3.5-4" />
        </svg>
      ),
    },
    {
      to: `/projects/${projectId}/tms`,
      label: "TMS",
      icon: (
        <svg className="h-3.5 w-3.5 text-green-400" fill="none" viewBox="0 0 16 16" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M2 3h12M2 6.5h8M2 10h10M2 13.5h6" />
        </svg>
      ),
    },
    {
      to: `/projects/${projectId}/backlog`,
      label: "Бэклог",
      icon: (
        <svg className="h-3.5 w-3.5 text-slate-400" fill="none" viewBox="0 0 16 16" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M3 4h10M3 8h10M3 12h10" />
        </svg>
      ),
    },
    {
      to: `/projects/${projectId}/sprints`,
      label: "Спринты",
      icon: (
        <svg className="h-3.5 w-3.5 text-purple-400" fill="none" viewBox="0 0 16 16" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M3 13c2-4 4-8 10-10" /><path d="M10 3h3v3" />
        </svg>
      ),
    },
  ];
}

export function Sidebar() {
  const { collapsed, toggle } = useSidebarStore();
  const { t } = useTranslation("common");
  const matchRoute = useMatchRoute();

  /** ID раскрытого проекта в боковом меню. */
  const [expandedProjectId, setExpandedProjectId] = useState<string | null>(
    null,
  );

  const { data: projects } = useQuery({
    queryKey: ["projects"],
    queryFn: listProjects,
  });

  /** Переключение раскрытия подменю проекта. */
  const toggleProject = (projectId: string) => {
    setExpandedProjectId((prev) => (prev === projectId ? null : projectId));
  };

  return (
    <aside
      className={cn(
        "flex h-screen flex-col border-r border-white/[0.06] transition-all duration-300 ease-in-out",
        "bg-[#0C1222] text-slate-300",
        collapsed ? "w-[68px]" : "w-[260px]",
      )}
    >
      {/* Логотип и кнопка сворачивания */}
      <div className="flex h-14 items-center justify-between px-4">
        {!collapsed && (
          <Link to="/projects" className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600">
              <span className="text-xs font-bold text-white">H</span>
            </div>
            <span className="text-lg font-bold text-white tracking-tight">
              Hub
            </span>
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
            <svg
              width="14"
              height="14"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M10 12l-4-4 4-4" />
            </svg>
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
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 14 14"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                >
                  <path d="M7 1v12M1 7h12" />
                </svg>
              </Link>
            </TooltipTrigger>
            <TooltipContent side="right">
              {t("sidebar.createProject")}
            </TooltipContent>
          </Tooltip>
        </div>

        <div className="space-y-0.5">
          {projects?.map((project) => {
            const isProjectActive = matchRoute({
              to: "/projects/$projectId",
              fuzzy: true,
              params: { projectId: project.id },
            });
            const isExpanded =
              expandedProjectId === project.id || isProjectActive;

            return (
              <div key={project.id}>
                {/* Основная ссылка проекта с кнопкой раскрытия */}
                <div className="flex items-center">
                  {/* Кнопка раскрытия подменю (только в развёрнутом режиме) */}
                  {!collapsed && (
                    <button
                      type="button"
                      onClick={() => toggleProject(project.id)}
                      className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-slate-500 hover:text-white transition-colors"
                      aria-label={
                        isExpanded ? "Свернуть модули" : "Развернуть модули"
                      }
                    >
                      <svg
                        width="10"
                        height="10"
                        viewBox="0 0 10 10"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="1.5"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className={cn(
                          "transition-transform duration-150",
                          isExpanded && "rotate-90",
                        )}
                      >
                        <path d="M3.5 1.5l4 3.5-4 3.5" />
                      </svg>
                    </button>
                  )}
                  <div className="flex-1">
                    <SidebarLink
                      to={`/projects/${project.id}`}
                      label={project.name}
                      collapsed={collapsed}
                      icon={
                        <span className="relative flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-blue-500/20 text-[10px] font-bold text-blue-400">
                          {project.key.charAt(0)}
                          {/* Цветной индикатор активности проекта */}
                          <span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-emerald-400 ring-2 ring-[#0C1222]" />
                        </span>
                      }
                    />
                  </div>
                </div>

                {/* Подменю модулей проекта */}
                {isExpanded && !collapsed && (
                  <div className="space-y-0.5 mt-0.5 mb-1">
                    {getProjectSubNav(project.id).map((item) => (
                      <SidebarLink
                        key={item.to}
                        to={item.to}
                        label={item.label}
                        collapsed={collapsed}
                        params={{ projectId: project.id }}
                        indent
                        icon={item.icon}
                      />
                    ))}
                  </div>
                )}
              </div>
            );
          })}
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
          to="/dashboard"
          label={t("nav.dashboard", { defaultValue: "Дашборд" })}
          collapsed={collapsed}
          icon={
            <svg
              className={NAV_ICON_COLORS["/dashboard"]}
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="1" y="1" width="6" height="6" rx="1" />
              <rect x="9" y="1" width="6" height="3" rx="1" />
              <rect x="9" y="6" width="6" height="9" rx="1" />
              <rect x="1" y="9" width="6" height="6" rx="1" />
            </svg>
          }
        />
        <SidebarLink
          to="/messenger"
          label={t("nav.messenger")}
          collapsed={collapsed}
          icon={
            <svg
              className={NAV_ICON_COLORS["/messenger"]}
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M2.5 12.5l1.5-4L2 4a1 1 0 01.9-1.4h10.2A1 1 0 0114 4l-2 4.5 1.5 4a1 1 0 01-.9 1.4H2.4a1 1 0 01-.9-1.4z" />
              <path d="M6 7h4M6 9.5h2.5" />
            </svg>
          }
        />
        <SidebarLink
          to="/projects"
          label={t("nav.allProjects")}
          collapsed={collapsed}
          icon={
            <svg
              className={NAV_ICON_COLORS["/projects"]}
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <rect x="2" y="2" width="5" height="5" rx="1" />
              <rect x="9" y="2" width="5" height="5" rx="1" />
              <rect x="2" y="9" width="5" height="5" rx="1" />
              <rect x="9" y="9" width="5" height="5" rx="1" />
            </svg>
          }
        />
        <SidebarLink
          to="/settings/profile"
          label={t("nav.settings")}
          collapsed={collapsed}
          icon={
            <svg
              className={NAV_ICON_COLORS["/settings/profile"]}
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
          }
        />
        <SidebarLink
          to="/trash"
          label={t("nav.trash")}
          collapsed={collapsed}
          icon={
            <svg
              className={NAV_ICON_COLORS["/trash"]}
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M2 4h12M5.3 4V2.7a1.3 1.3 0 011.4-1.4h2.6a1.3 1.3 0 011.4 1.4V4M12.7 4v9.3a1.3 1.3 0 01-1.4 1.4H4.7a1.3 1.3 0 01-1.4-1.4V4" />
            </svg>
          }
        />
      </div>

      {/* Кнопка развернуть -- в свёрнутом состоянии */}
      {collapsed && (
        <div className="px-3 pb-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            className="w-full h-7 text-slate-500 hover:text-white hover:bg-white/[0.06] rounded-md"
            aria-label={t("sidebar.expand")}
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M6 12l4-4-4-4" />
            </svg>
          </Button>
        </div>
      )}
    </aside>
  );
}
