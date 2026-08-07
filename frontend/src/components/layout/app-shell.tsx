/**
 * Основная оболочка приложения: sidebar + topbar + content + панель звонка.
 *
 * На десктопе (md+): sidebar всегда видим, контент рядом.
 * На мобильных: sidebar скрыт, открывается как overlay. Нижняя навигация.
 */

import { Outlet, useNavigate, useRouterState } from "@tanstack/react-router";
import { lazy, Suspense, useEffect, useState } from "react";
import { toast } from "sonner";
import { SkipLink } from "@/components/layout/skip-link";
import { MobileNav } from "@/components/mobile-nav";
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

/** A13-015: AI / search / help — не в initial graph каждого authenticated page. */
const AIChatPanel = lazy(() =>
	import("@/components/ai/ai-chat-panel").then((m) => ({
		default: m.AIChatPanel,
	})),
);
const CommandPalette = lazy(() =>
	import("@/components/search/command-palette").then((m) => ({
		default: m.CommandPalette,
	})),
);
const ShortcutsDialog = lazy(() =>
	import("@/components/help/shortcuts-dialog").then((m) => ({
		default: m.ShortcutsDialog,
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
	const commandOpen = useCommandStore((s) => s.open);
	const toggleCommand = useCommandStore((s) => s.toggle);
	const aiPanelOpen = useAIStore((s) => s.panelOpen);
	const toggleAIPanel = useAIStore((s) => s.togglePanel);
	const navigate = useNavigate();
	const pathname = useRouterState({ select: (s) => s.location.pathname });

	// Закрыть mobile overlay при смене route / Escape (W9-A2-008).
	useEffect(() => {
		// pathname — намеренный сигнал эффекта: закрываем overlay после навигации.
		void pathname;
		setMobileMenuOpen(false);
	}, [pathname]);

	useEffect(() => {
		if (!mobileMenuOpen) return;
		function onKeyDown(e: KeyboardEvent) {
			if (e.key === "Escape") {
				setMobileMenuOpen(false);
			}
		}
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [mobileMenuOpen]);

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
						<ErrorBoundary resetKey={pathname}>
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

			{commandOpen && (
				<Suspense fallback={null}>
					<CommandPalette />
				</Suspense>
			)}
			{aiPanelOpen && (
				<Suspense fallback={null}>
					<AIChatPanel />
				</Suspense>
			)}
			{shortcutsOpen && (
				<Suspense fallback={null}>
					<ShortcutsDialog
						open={shortcutsOpen}
						onOpenChange={setShortcutsOpen}
					/>
				</Suspense>
			)}
		</TooltipProvider>
	);
}
