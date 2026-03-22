/**
 * Нижняя навигация сайдбара -- ссылки на основные разделы.
 */
import { NAV_ICON_COLORS, SidebarLink } from "./sidebar-link";

/** Блок навигации сайдбара: ссылки на дашборд, мессенджер, звонки, проекты, настройки, корзину. */
export function SidebarNav({
	collapsed,
	t,
}: {
	collapsed: boolean;
	t: (key: string, options?: Record<string, string>) => string;
}) {
	return (
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
				to="/calls"
				label={t("nav.calls", { defaultValue: "Созвоны" })}
				collapsed={collapsed}
				icon={
					<svg
						className={NAV_ICON_COLORS["/calls"]}
						width="16"
						height="16"
						viewBox="0 0 16 16"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path d="M2 3.5A1.5 1.5 0 013.5 2h2.13a1 1 0 01.97.76l.45 1.83a1 1 0 01-.5 1.1L5.1 6.5a8 8 0 004.4 4.4l.81-1.45a1 1 0 011.1-.5l1.83.45a1 1 0 01.76.97v2.13a1.5 1.5 0 01-1.5 1.5A11.5 11.5 0 012 3.5z" />
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
	);
}
