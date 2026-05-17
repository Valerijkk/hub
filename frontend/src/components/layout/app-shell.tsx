/**
 * Основная оболочка приложения: sidebar + topbar + content + панель звонка.
 *
 * На десктопе (md+): sidebar всегда видим, контент рядом.
 * На мобильных: sidebar скрыт, открывается как overlay. Нижняя навигация.
 */

import { Outlet } from "@tanstack/react-router";
import { Suspense, lazy, useState } from "react";
import { MobileNav } from "@/components/mobile-nav";
import { TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useCallStore } from "@/stores/call-store";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

/**
 * CallPanel загружается лениво (FE-H5): тянет за собой ~478KB livekit-client
 * и @livekit/components-react. Юзеры без звонков никогда не платят за этот
 * чанк. Vite автоматически выносит транзитивные deps в отдельный chunk.
 */
const CallPanel = lazy(() =>
	import("@/components/calls/CallPanel").then((m) => ({ default: m.CallPanel })),
);

/** Основная оболочка приложения: sidebar, topbar, content, панель звонка и мобильная навигация. */
export function AppShell() {
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

	// CallPanel рендерим только когда есть активный/коннектящийся звонок
	// или ошибка — иначе React не подтягивает lazy chunk вовсе.
	const callPanelNeeded = useCallStore(
		(s) => s.isInCall || s.isConnecting || s.error !== null,
	);

	return (
		<TooltipProvider delayDuration={300}>
			<div className="flex h-screen overflow-hidden bg-base">
				{/* Mobile overlay для sidebar */}
				{mobileMenuOpen && (
					<div
						className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm md:hidden"
						onClick={() => setMobileMenuOpen(false)}
						onKeyDown={(e) => {
							if (e.key === "Escape") setMobileMenuOpen(false);
						}}
					/>
				)}

				{/* Sidebar: на мобильных — slide-over, на десктопе — статичный */}
				<div
					className={cn(
						"fixed inset-y-0 left-0 z-50 transition-transform duration-200 md:relative md:z-auto md:translate-x-0",
						mobileMenuOpen
							? "translate-x-0"
							: "-translate-x-full md:translate-x-0",
					)}
				>
					<Sidebar />
				</div>

				{/* Основной контент */}
				<div className="flex flex-1 flex-col overflow-hidden">
					<Topbar onMenuToggle={() => setMobileMenuOpen((v) => !v)} />
					<main className="flex-1 overflow-y-auto px-4 py-4 pb-20 md:px-6 md:py-6 md:pb-6">
						<Outlet />
					</main>
				</div>
			</div>

			{/* Нижняя навигация для мобильных */}
			<MobileNav />

			{/* Панель активного звонка — lazy-loaded (FE-H5). Suspense fallback пуст:
			    CallPanel сам показывает loader при isConnecting, а пока чанк
			    качается (десятки ms) — ничего рендерить не нужно. */}
			{callPanelNeeded && (
				<Suspense fallback={null}>
					<CallPanel />
				</Suspense>
			)}
		</TooltipProvider>
	);
}
