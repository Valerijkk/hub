import { useNavigate } from "@tanstack/react-router";
import { Search } from "lucide-react";
import { useAuthStore } from "@/stores/auth-store";
import { useSearchStore } from "@/stores/search-store";
import { logout as apiLogout } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { NotificationBell } from "@/components/notifications/notification-bell";

interface TopbarProps {
  onMenuToggle?: () => void;
}

export function Topbar({ onMenuToggle }: TopbarProps) {
  const navigate = useNavigate();
  const { user, logout: storeLogout } = useAuthStore();
  const setSearchOpen = useSearchStore((s) => s.setOpen);

  const initials = user?.name
    ? user.name
        .split(" ")
        .map((n) => n[0])
        .join("")
        .toUpperCase()
        .slice(0, 2)
    : user?.email?.charAt(0)?.toUpperCase() ?? "U";

  async function handleLogout() {
    try {
      await apiLogout();
    } catch {
      // Logout even if API fails
    }
    storeLogout();
    void navigate({ to: "/login" });
  }

  return (
    <header className="flex h-14 items-center justify-between border-b bg-background px-4">
      {/* Left side */}
      <div className="flex items-center gap-2">
        {/* Mobile hamburger */}
        {onMenuToggle && (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 md:hidden"
            onClick={onMenuToggle}
            aria-label="Меню"
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M2 4h12M2 8h12M2 12h12" /></svg>
          </Button>
        )}
      </div>

      {/* Right side */}
      <div className="flex items-center gap-2">
        {/* Кнопка глобального поиска */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="hidden gap-2 text-muted-foreground sm:flex"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="h-3.5 w-3.5" />
              <span className="text-xs">Поиск</span>
              <kbd className="pointer-events-none ml-2 inline-flex h-5 items-center gap-1 rounded border bg-muted px-1.5 text-[10px] font-medium text-muted-foreground">
                Ctrl+K
              </kbd>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Глобальный поиск (Ctrl+K)</TooltipContent>
        </Tooltip>

        {/* Уведомления */}
        <NotificationBell />

        {/* User menu */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full">
              <Avatar className="h-8 w-8">
                <AvatarImage src={user?.avatar_url ?? undefined} alt={user?.name ?? "User"} />
                <AvatarFallback className="text-xs">{initials}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48">
            <div className="px-2 py-1.5">
              <p className="text-sm font-medium">{user?.name ?? "Пользователь"}</p>
              <p className="text-xs text-muted-foreground">{user?.email}</p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => void navigate({ to: "/settings/profile" })}>
              Профиль
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              Выйти
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
