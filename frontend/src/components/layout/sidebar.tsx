/**
 * Боковая панель навигации -- Slack-style тёмная тема.
 *
 * Лого Hub слева, compose-кнопка справа, навигация со Slack-style
 * активными индикаторами (3px blue bar), проекты, профиль внизу.
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
				"flex h-screen flex-col border-r border-white/[0.06] transition-all duration-150 ease-in-out",
				"bg-[#19171D] text-[#D1D2D3]",
				collapsed ? "w-[68px]" : "w-[260px]",
			)}
		>
			{/* Лого и compose/collapse */}
			<div className="flex h-[49px] shrink-0 items-center justify-between px-3">
				{!collapsed ? (
					<>
						<Link
							to="/projects"
							className="flex items-center gap-2.5 group"
						>
							<div className="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 transition-transform duration-150 group-hover:scale-105">
								<span className="text-xs font-bold text-white select-none">
									H
								</span>
							</div>
							<span className="text-[15px] font-bold text-white tracking-tight">
								Hub
							</span>
						</Link>
						<div className="flex items-center gap-1">
							{/* Compose / новое сообщение */}
							<Tooltip>
								<TooltipTrigger asChild>
									<Link
										to="/projects/new"
										className="flex h-7 w-7 items-center justify-center rounded-full text-[#D1D2D3] hover:text-white hover:bg-white/[0.08] transition-all duration-150"
									>
										<svg
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
										className="flex h-7 w-7 items-center justify-center rounded-md text-[#D1D2D3]/50 hover:text-white hover:bg-white/[0.05] transition-all duration-150"
										aria-label={t("sidebar.collapse")}
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
								className="mx-auto flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 transition-transform duration-150 hover:scale-105"
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
			<div className="mx-3 h-px bg-white/[0.06]" />

			{/* Навигация */}
			<SidebarNav collapsed={collapsed} t={t} />

			{/* Разделитель */}
			<div className="mx-3 h-px bg-white/[0.06]" />

			{/* Секция проектов */}
			<SidebarProjects collapsed={collapsed} t={t} />

			{/* Профиль пользователя */}
			<SidebarUserProfile collapsed={collapsed} toggle={toggle} />
		</aside>
	);
}
