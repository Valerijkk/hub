import { useEffect, useState } from "react";
import { Outlet } from "@tanstack/react-router";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";
import { cn } from "@/lib/utils";
import { CallPanel } from "@/components/calls/CallPanel";
import { CommandPalette } from "@/components/search/command-palette";
import { useSearchStore } from "@/stores/search-store";

/**
 * Основная оболочка приложения: sidebar + topbar + content + панель звонка.
 * Регистрирует глобальные горячие клавиши: Ctrl+K (поиск), Ctrl+N (новая задача).
 */
export function AppShell() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const setSearchOpen = useSearchStore((s) => s.setOpen);
  const searchIsOpen = useSearchStore((s) => s.isOpen);

  // Глобальные горячие клавиши
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Ctrl+K / Cmd+K — открыть/закрыть палитру поиска
      if ((e.ctrlKey || e.metaKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen(!searchIsOpen);
      }
      // Ctrl+N / Cmd+N — создание новой задачи (dispatch custom event)
      if ((e.ctrlKey || e.metaKey) && e.key === "n") {
        e.preventDefault();
        window.dispatchEvent(new CustomEvent("hub:create-issue"));
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [setSearchOpen, searchIsOpen]);

  return (
    <TooltipProvider delayDuration={300}>
      <div className="flex h-screen overflow-hidden">
        {/* Mobile overlay */}
        {mobileMenuOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/50 md:hidden"
            onClick={() => setMobileMenuOpen(false)}
            onKeyDown={(e) => {
              if (e.key === "Escape") setMobileMenuOpen(false);
            }}
          />
        )}

        {/* Sidebar */}
        <div
          className={cn(
            "fixed inset-y-0 left-0 z-50 md:relative md:z-auto",
            mobileMenuOpen ? "block" : "hidden md:block"
          )}
        >
          <Sidebar />
        </div>

        {/* Main content */}
        <div className="flex flex-1 flex-col overflow-hidden">
          <Topbar onMenuToggle={() => setMobileMenuOpen((v) => !v)} />
          <main className="flex-1 overflow-y-auto p-6">
            <Outlet />
          </main>
        </div>
      </div>
      {/* Палитра команд (Ctrl+K) */}
      <CommandPalette />
      {/* Панель активного звонка (фиксированная, если звонок активен) */}
      <CallPanel />
    </TooltipProvider>
  );
}
