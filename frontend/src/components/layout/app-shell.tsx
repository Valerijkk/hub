/**
 * Основная оболочка приложения: sidebar + topbar + content + панель звонка.
 *
 * На десктопе (md+): sidebar всегда видим, контент рядом.
 * На мобильных: sidebar скрыт, открывается как overlay. Нижняя навигация.
 */

import { Outlet, useNavigate } from "@tanstack/react-router";
import { lazy, Suspense, useEffect, useState } from "react";
import { toast } from "sonner";
import { AIChatPanel } from "@/components/ai/ai-chat-panel";
import { ShortcutsDialog } from "@/components/help/shortcuts-dialog";
import { SkipLink } from "@/components/layout/skip-link";
import { MobileNav } from "@/components/mobile-nav";
import { CommandPalette } from "@/components/search/command-palette";
import { ErrorBoundary } from "@/components/ui/error-boundary";
import { TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { useAIStore } from "@/stores/ai-store";
import { useCallStore } from "@/stores/call-store";
import { useCommandStore } from "@/stores/command-store";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

const CallPanel = lazy(() =>
	import("@/components/calls/CallPanel").then((m) => ({
		default: m.CallPanel,
	})),
);

function isEditableTarget(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) return false;
	const tag = target.tagName;
	if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
	if (target.isContentEditable) return true;
	return target.closest("[contenteditable='true']") !== null;
}

function projectIdFromPath(pathname: string): string | null {
	const m = pathname.match(/\/projects\/([0-9a-f-]{36})/i);
	return m?.[1] ?? null;
}

/** Основная оболочка приложения: sidebar, topbar, content, панель звонка и мобильная навигация. */
export function AppShell() {
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
	const [shortcutsOpen, setShortcutsOpen] = useState(false);
	const toggleCommand = useCommandStore((s) => s.toggle);
	const toggleAIPanel = useAIStore((s) => s.togglePanel);
	const navigate = useNavigate();

	useEffect(() => {
		function onKeyDown(e: KeyboardEvent) {
			if ((e.metaKey || e.ctrlKey) && (e.key === "k" || e.key === "K")) {
				e.preventDefault();
				toggleCommand();
				return;
			}

			if (
				(e.metaKey || e.ctrlKey) &&
				(e.key === "n" || e.key === "N") &&
				!isEditableTarget(e.target)
			) {
				e.preventDefault();
				const projectId = projectIdFromPath(window.location.pathname);
				if (projectId) {
					void navigate({
						to: "/projects/$projectId/issues/new",
						params: { projectId },
					});
				} else {
					toast.message("Откройте проект для создания задачи (Ctrl+N)");
				}
				return;
			}

			if (
				(e.metaKey || e.ctrlKey) &&
				e.shiftKey &&
				(e.key === "a" || e.key === "A") &&
				!isEditableTarget(e.target)
			) {
				e.preventDefault();
				toggleAIPanel();
				return;
			}

			if (
				e.key === "?" &&
				!e.metaKey &&
				!e.ctrlKey &&
				!e.altKey &&
				!isEditableTarget(e.target)
			) {
				e.preventDefault();
				setShortcutsOpen(true);
			}
		}
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [toggleCommand, toggleAIPanel, navigate]);

	const callPanelNeeded = useCallStore(
		(s) => s.isInCall || s.isConnecting || s.error !== null,
	);

	return (
		<TooltipProvider delayDuration={300}>
			<SkipLink />
			<div className="flex h-screen overflow-hidden bg-base">
				{mobileMenuOpen && (
					<button
						type="button"
						className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm md:hidden"
						aria-label="Закрыть меню"
						onClick={() => setMobileMenuOpen(false)}
					/>
				)}

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

				<div className="flex flex-1 flex-col overflow-hidden">
					<Topbar onMenuToggle={() => setMobileMenuOpen((v) => !v)} />
					<main
						id="main-content"
						className="flex-1 overflow-y-auto px-4 py-4 pb-20 md:px-6 md:py-6 md:pb-6"
					>
						<ErrorBoundary>
							<Outlet />
						</ErrorBoundary>
					</main>
				</div>
			</div>

			<MobileNav />

			{callPanelNeeded && (
				<Suspense fallback={null}>
					<CallPanel />
				</Suspense>
			)}

			<CommandPalette />
			<AIChatPanel />
			<ShortcutsDialog open={shortcutsOpen} onOpenChange={setShortcutsOpen} />
		</TooltipProvider>
	);
}
