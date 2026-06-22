/**
 * Боковая панель навигации -- redesigned v2.0.
 *
 * Лого Hub слева, compose-кнопка справа, навигация со
 * активными индикаторами (accent bar + glow), проекты, профиль внизу.
 * Сворачивается до иконок. На мобильных -- overlay.
 */
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useSidebarStore } from "@/stores/sidebar-store";
import { SidebarNav } from "./sidebar-nav";
import { SidebarProjects } from "./sidebar-projects";
import { SidebarUserProfile } from "./sidebar-user";

/** Боковая панель навигации с логотипом, проектами, ссылками и профилем. */
export function Sidebar() {
	const { collapsed, toggle } = useSidebarStore();
	const { t } = useTranslation("common");

	return (
		<aside
			className={cn(
				"flex h-screen flex-col border-r border-b1 transition-[width,padding] duration-200 ease-[var(--ease)] relative",
				"bg-sidebar text-sidebar-foreground",
				collapsed ? "w-[64px]" : "w-[260px]",
			)}
		>
			{/* Градиентная граница sidebar */}
			<div
				className="pointer-events-none absolute top-0 right-[-1px] bottom-0 w-px bg-[linear-gradient(180deg,transparent_5%,rgba(251,191,36,.05)_30%,rgba(91,156,245,.04)_60%,transparent_95%)] z-10"
				aria-hidden="true"
			/>

			{/* Лого и compose/collapse */}
			<div className="flex h-14 shrink-0 items-center justify-between px-[18px]">
				{!collapsed ? (
					<>
						<Link to="/projects" className="flex items-center gap-2.5 group">
							<div className="flex h-[30px] w-[30px] items-center justify-center rounded-[9px] bg-gradient-to-br from-blue-500 to-purple-600 shadow-[0_2px_16px_var(--gB)] logo-glow transition-transform duration-150 group-hover:scale-105">
								<span className="text-sm font-bold text-white select-none">
									H
								</span>
							</div>
							<span className="text-[18px] font-[800] tracking-[-0.04em] bg-gradient-to-r from-white to-white/70 bg-clip-text text-transparent">
								Hub
							</span>
						</Link>
						<div className="flex items-center gap-1">
							{/* Compose / новое сообщение */}
							<Tooltip>
								<TooltipTrigger asChild>
									<Link
										to="/projects/new"
										className="flex h-7 w-7 items-center justify-center rounded-full text-sidebar-muted hover:text-sidebar-foreground hover:bg-[var(--hover)] transition-colors duration-150"
									>
										<svg
											aria-hidden="true"
											width="16"
											height="16"
											viewBox="0 0 16 16"
											fill="none"
											stroke="currentColor"
											strokeWidth="2"
											strokeLinecap="round"
										>
											<path d="M8 3v10M3 8h10" />
										</svg>
									</Link>
								</TooltipTrigger>
								<TooltipContent side="bottom">
									{t("sidebar.createProject", {
										defaultValue: "Создать",
									})}
								</TooltipContent>
							</Tooltip>
							{/* Свернуть сайдбар */}
							<Tooltip>
								<TooltipTrigger asChild>
									<button
										type="button"
										onClick={toggle}
										className="flex h-7 w-7 items-center justify-center rounded-md text-sidebar-subtle hover:text-sidebar-foreground hover:bg-[var(--hover)] transition-colors duration-150"
										aria-label={t("sidebar.collapse")}
									>
										<svg
											aria-hidden="true"
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
									</button>
								</TooltipTrigger>
								<TooltipContent side="bottom">
									{t("sidebar.collapse", {
										defaultValue: "Свернуть",
									})}
								</TooltipContent>
							</Tooltip>
						</div>
					</>
				) : (
					<Tooltip>
						<TooltipTrigger asChild>
							<button
								type="button"
								onClick={toggle}
								className="mx-auto flex h-[30px] w-[30px] items-center justify-center rounded-[9px] bg-gradient-to-br from-blue-500 to-purple-600 shadow-[0_2px_16px_var(--gB)] transition-transform duration-150 hover:scale-105"
								aria-label={t("sidebar.expand")}
							>
								<span className="text-xs font-bold text-white select-none">
									H
								</span>
							</button>
						</TooltipTrigger>
						<TooltipContent side="right">
							{t("sidebar.expand", {
								defaultValue: "Развернуть",
							})}
						</TooltipContent>
					</Tooltip>
				)}
			</div>

			{/* Разделитель */}
			<div className="mx-[18px] h-px bg-b1" />

			{/* Навигация */}
			<SidebarNav collapsed={collapsed} t={t} />

			{/* Разделитель */}
			<div className="mx-[18px] h-px bg-b1" />

			{/* Секция проектов */}
			<SidebarProjects collapsed={collapsed} t={t} />

			{/* Профиль пользователя */}
			<SidebarUserProfile collapsed={collapsed} toggle={toggle} />
		</aside>
	);
}
