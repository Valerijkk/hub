/**
 * Основная оболочка приложения: sidebar + topbar + content + панель звонка.
 *
 * На десктопе (md+): sidebar всегда видим, контент рядом.
 * На мобильных: sidebar скрыт, открывается как overlay. Нижняя навигация.
 * Регистрирует глобальную горячую клавишу: Ctrl+N (новая задача).
 */

import { Outlet } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { CallPanel } from "@/components/calls/CallPanel";
import { MobileNav } from "@/components/mobile-nav";
import { TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

/** Основная оболочка приложения: sidebar, topbar, content, панель звонка и мобильная навигация. */
export function AppShell() {
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

	/* Глобальные горячие клавиши */
	useEffect(() => {
		function handleKeyDown(e: KeyboardEvent) {
			if ((e.ctrlKey || e.metaKey) && e.key === "n") {
				e.preventDefault();
				window.dispatchEvent(new CustomEvent("hub:create-issue"));
			}
		}

		document.addEventListener("keydown", handleKeyDown);
		return () => document.removeEventListener("keydown", handleKeyDown);
	}, []);

	return (
		<TooltipProvider delayDuration={300}>
			<div className="flex h-screen overflow-hidden bg-background">
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
					<div className="md:hidden">
						<Topbar onMenuToggle={() => setMobileMenuOpen((v) => !v)} />
					</div>
					<main className="flex-1 overflow-y-auto px-4 py-4 pb-20 md:px-6 md:pb-6">
						<Outlet />
					</main>
				</div>
			</div>

			{/* Нижняя навигация для мобильных */}
			<MobileNav />

			{/* Панель активного звонка */}
			<CallPanel />
		</TooltipProvider>
	);
}
