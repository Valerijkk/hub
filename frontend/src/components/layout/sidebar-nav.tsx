/**
 * Навигация сайдбара -- Slack-style ссылки на основные разделы.
 * Иконки 20px (viewBox 20x20), высота элемента 32px.
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
		<div className="space-y-0.5 px-3 py-2">
			<SidebarLink
				to="/dashboard"
				label={t("nav.dashboard", { defaultValue: "Дашборд" })}
				collapsed={collapsed}
				icon={
					<svg
						className={NAV_ICON_COLORS["/dashboard"]}
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<rect x="2" y="2" width="7" height="7" rx="1.5" />
						<rect x="11" y="2" width="7" height="4" rx="1.5" />
						<rect x="11" y="8" width="7" height="10" rx="1.5" />
						<rect x="2" y="11" width="7" height="7" rx="1.5" />
					</svg>
				}
			/>
			<SidebarLink
				to="/messenger"
				label={t("nav.messenger")}
				collapsed={collapsed}
				badge={5}
				icon={
					<svg
						className={NAV_ICON_COLORS["/messenger"]}
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path d="M3 14.5l1.8-4.5L3 5a1.2 1.2 0 011.1-1.7h11.8A1.2 1.2 0 0117 5l-2.3 5 1.8 4.5a1.2 1.2 0 01-1.1 1.7H4.1A1.2 1.2 0 013 14.5z" />
						<path d="M7.5 8.5h5M7.5 11.5h3" />
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
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path d="M3 4.5A1.5 1.5 0 014.5 3h2.6a1 1 0 01.97.76l.55 2.2a1 1 0 01-.5 1.1L6.5 8.2a9.5 9.5 0 005.3 5.3l1.14-.62a1 1 0 011.1-.5l2.2.55a1 1 0 01.76.97v2.6a1.5 1.5 0 01-1.5 1.5A14 14 0 013 4.5z" />
					</svg>
				}
			/>
			<SidebarLink
				to="/projects"
				label={t("nav.allProjects")}
				collapsed={collapsed}
				badge={12}
				icon={
					<svg
						className={NAV_ICON_COLORS["/projects"]}
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<rect x="2.5" y="2.5" width="6" height="6" rx="1.5" />
						<rect x="11.5" y="2.5" width="6" height="6" rx="1.5" />
						<rect x="2.5" y="11.5" width="6" height="6" rx="1.5" />
						<rect x="11.5" y="11.5" width="6" height="6" rx="1.5" />
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
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<circle cx="10" cy="10" r="3" />
						<path d="M16.5 10a6.5 6.5 0 01-.35 2.1l1.3.75a.6.6 0 01.22.82l-1.2 2.06a.6.6 0 01-.72.26l-1.3-.75a6.5 6.5 0 01-1.8 1.04v1.5a.6.6 0 01-.6.6h-2.1a.6.6 0 01-.6-.6v-1.5a6.5 6.5 0 01-1.8-1.04l-1.3.75a.6.6 0 01-.72-.26l-1.2-2.06a.6.6 0 01.22-.82l1.3-.75A6.5 6.5 0 013.5 10c0-.72.12-1.42.35-2.1l-1.3-.75a.6.6 0 01-.22-.82l1.2-2.06a.6.6 0 01.72-.26l1.3.75A6.5 6.5 0 017.35 3.72v-1.5a.6.6 0 01.6-.6h2.1a.6.6 0 01.6.6v1.5a6.5 6.5 0 011.8 1.04l1.3-.75a.6.6 0 01.72.26l1.2 2.06a.6.6 0 01-.22.82l-1.3.75c.23.68.35 1.38.35 2.1z" />
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
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path d="M3 5h14M7 5V3.5A1.5 1.5 0 018.5 2h3A1.5 1.5 0 0113 3.5V5M15.5 5v11a1.5 1.5 0 01-1.5 1.5H6A1.5 1.5 0 014.5 16V5" />
					</svg>
				}
			/>
		</div>
	);
}
