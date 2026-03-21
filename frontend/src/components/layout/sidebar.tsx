/**
 * Боковая панель навигации -- всегда тёмная, минимальный дизайн.
 *
 * Лого Hub с градиентным значком, проекты с цветными точками,
 * навигация с тонкими hover-эффектами и активными индикаторами.
 * Сворачивается до иконок. На мобильных -- overlay.
 * При раскрытии проекта показывает подменю: Доска, Задачи, TMS, Бэклог, Спринты.
 */
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { useSidebarStore } from "@/stores/sidebar-store";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { SidebarProjects } from "./sidebar-projects";
import { SidebarNav } from "./sidebar-nav";
import { SidebarExpandButton } from "./sidebar-user";

export function Sidebar() {
  const { collapsed, toggle } = useSidebarStore();
  const { t } = useTranslation("common");

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
      <SidebarProjects collapsed={collapsed} t={t} />

      {/* Разделитель */}
      <div className="mx-4 h-px bg-white/[0.06]" />

      {/* Нижняя навигация */}
      <SidebarNav collapsed={collapsed} t={t} />

      {/* Кнопка развернуть -- в свёрнутом состоянии */}
      <SidebarExpandButton
        collapsed={collapsed}
        toggle={toggle}
        ariaLabel={t("sidebar.expand")}
      />
    </aside>
  );
}
