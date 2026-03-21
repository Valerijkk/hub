/**
 * Секция проектов сайдбара — список с expand/collapse подменю модулей.
 */
import { useState } from "react";
import { Link, useMatchRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { listProjects } from "@/api/projects";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { SidebarLink } from "./sidebar-link";

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

export function SidebarProjects({
  collapsed,
  t,
}: {
  collapsed: boolean;
  t: (key: string, options?: Record<string, string>) => string;
}) {
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
  );
}
