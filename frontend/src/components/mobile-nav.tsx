/**
 * Нижняя навигация для мобильных устройств (< 768px).
 *
 * Отображает 5 иконок: Дашборд, Трекер, Мессенджер, Звонки, Ещё (меню).
 * Фиксируется внизу экрана. Подсвечивает активную секцию.
 * Все подписи локализованы через i18next.
 */
import { Link, useMatchRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

interface MobileNavItemProps {
	to: string;
	label: string;
	icon: React.ReactNode;
	isActive: boolean;
}

function MobileNavItem({ to, label, icon, isActive }: MobileNavItemProps) {
	return (
		<Link
			to={to}
			className={cn(
				"flex flex-1 flex-col items-center gap-0.5 py-2 text-[10px] font-medium transition-colors",
				isActive ? "text-primary" : "text-t3",
			)}
		>
			{icon}
			<span>{label}</span>
		</Link>
	);
}

/**
 * Нижняя навигация, видимая только на мобильных (md:hidden)
 */
export function MobileNav() {
	const { t } = useTranslation("common");
	const matchRoute = useMatchRoute();
	const [moreOpen, setMoreOpen] = useState(false);

	const navItems = [
		{
			to: "/dashboard",
			label: t("nav.dashboard"),
			match: "/dashboard",
			icon: (
				<svg
					aria-hidden="true"
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
					<rect x="11" y="2" width="7" height="7" rx="1.5" />
					<rect x="2" y="11" width="7" height="7" rx="1.5" />
					<rect x="11" y="11" width="7" height="7" rx="1.5" />
				</svg>
			),
		},
		{
			to: "/projects",
			label: t("nav.tracker"),
			match: "/projects",
			icon: (
				<svg
					aria-hidden="true"
					width="20"
					height="20"
					viewBox="0 0 20 20"
					fill="none"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M4 5h12M4 10h12M4 15h8" />
				</svg>
			),
		},
		{
			to: "/messenger",
			label: t("nav.messenger"),
			match: "/messenger",
			icon: (
				<svg
					aria-hidden="true"
					width="20"
					height="20"
					viewBox="0 0 20 20"
					fill="none"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M3 14.5l1.5-4L3 5a1 1 0 01.9-1.4h12.2A1 1 0 0117 5l-2 5.5 1.5 4a1 1 0 01-.9 1.4H3.4a1 1 0 01-.4-.9z" />
					<path d="M7 8h6M7 11h3" />
				</svg>
			),
		},
		{
			to: "/calls",
			label: t("nav.calls"),
			match: "/calls",
			icon: (
				<svg
					aria-hidden="true"
					width="20"
					height="20"
					viewBox="0 0 20 20"
					fill="none"
					stroke="currentColor"
					strokeWidth="1.5"
					strokeLinecap="round"
					strokeLinejoin="round"
				>
					<path d="M17 13.5v2a2 2 0 01-2.18 2 13.16 13.16 0 01-5.74-2.04 13 13 0 01-4-4A13.16 13.16 0 013 5.68 2 2 0 014.98 3.5h2a2 2 0 012 1.72c.13.96.36 1.9.7 2.81a2 2 0 01-.45 2.11l-.85.85a16 16 0 004 4l.85-.85a2 2 0 012.11-.45c.91.34 1.85.57 2.81.7a2 2 0 011.72 2.04z" />
				</svg>
			),
		},
	];

	return (
		<>
			{/* Меню "Ещё" -- overlay */}
			{moreOpen && (
				<>
					<button
						type="button"
						className="fixed inset-0 z-40 cursor-default border-0 bg-black/50 p-0 md:hidden"
						aria-label="Закрыть меню"
						onClick={() => setMoreOpen(false)}
					/>
					<div
						id="mobile-nav-more-panel"
						className="fixed inset-x-0 bottom-14 z-50 mx-4 mb-1 rounded-lg border border-b1 bg-card p-3 shadow-lg md:hidden"
					>
						<div className="space-y-1">
							<Link
								to="/settings/profile"
								className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium hover:bg-[var(--hover)]"
								onClick={() => setMoreOpen(false)}
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
									strokeLinejoin="round"
								>
									<circle cx="8" cy="8" r="2.5" />
									<path d="M13.5 8a5.5 5.5 0 01-.3 1.8l1.1.6a.5.5 0 01.2.7l-1 1.7a.5.5 0 01-.6.2l-1.1-.6a5.5 5.5 0 01-1.5.9v1.2a.5.5 0 01-.5.5h-2a.5.5 0 01-.5-.5v-1.2a5.5 5.5 0 01-1.5-.9l-1.1.6a.5.5 0 01-.6-.2l-1-1.7a.5.5 0 01.2-.7l1.1-.6A5.5 5.5 0 012.5 8c0-.6.1-1.2.3-1.8l-1.1-.6a.5.5 0 01-.2-.7l1-1.7a.5.5 0 01.6-.2l1.1.6a5.5 5.5 0 011.5-.9V1.5a.5.5 0 01.5-.5h2a.5.5 0 01.5.5v1.2a5.5 5.5 0 011.5.9l1.1-.6a.5.5 0 01.6.2l1 1.7a.5.5 0 01-.2.7l-1.1.6c.2.6.3 1.2.3 1.8z" />
								</svg>
								{t("nav.settings")}
							</Link>
							<Link
								to="/trash"
								className="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium hover:bg-[var(--hover)]"
								onClick={() => setMoreOpen(false)}
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
									strokeLinejoin="round"
								>
									<path d="M2 4h12M5.3 4V2.7a1.3 1.3 0 011.4-1.4h2.6a1.3 1.3 0 011.4 1.4V4M12.7 4v9.3a1.3 1.3 0 01-1.4 1.4H4.7a1.3 1.3 0 01-1.4-1.4V4" />
								</svg>
								{t("nav.trash")}
							</Link>
						</div>
					</div>
				</>
			)}

			{/* Нижняя панель */}
			<nav className="fixed inset-x-0 bottom-0 z-30 flex border-t border-b1 bg-base md:hidden">
				{navItems.map((item) => (
					<MobileNavItem
						key={item.to}
						to={item.to}
						label={item.label}
						icon={item.icon}
						isActive={!!matchRoute({ to: item.match, fuzzy: true })}
					/>
				))}
				{/* Кнопка "Ещё" */}
				<button
					type="button"
					onClick={() => setMoreOpen((v) => !v)}
					aria-expanded={moreOpen}
					aria-controls="mobile-nav-more-panel"
					className={cn(
						"flex flex-1 flex-col items-center gap-0.5 py-2 text-[10px] font-medium transition-colors",
						moreOpen ? "text-primary" : "text-t3",
					)}
				>
					<svg
						aria-hidden="true"
						width="20"
						height="20"
						viewBox="0 0 20 20"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
					>
						<circle cx="10" cy="5" r="1.5" />
						<circle cx="10" cy="10" r="1.5" />
						<circle cx="10" cy="15" r="1.5" />
					</svg>
					<span>{t("nav.more")}</span>
				</button>
			</nav>
		</>
	);
}
