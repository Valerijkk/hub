/**
 * Верхняя панель приложения — чистый дизайн с градиентным акцентом.
 *
 * Тонкая градиентная линия снизу, стеклянная кнопка поиска,
 * аватар с градиентной рамкой.
 * На мобильных устройствах показывается hamburger-кнопка для открытия sidebar.
 * Все строки локализованы через i18next (common namespace).
 */
import { useNavigate } from "@tanstack/react-router";
import { Search } from "lucide-react";
import { useTranslation } from "react-i18next";
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
import { LanguageSwitcher } from "@/components/language-switcher";

interface TopbarProps {
  onMenuToggle?: () => void;
}

export function Topbar({ onMenuToggle }: TopbarProps) {
  const navigate = useNavigate();
  const { user, logout: storeLogout } = useAuthStore();
  const setSearchOpen = useSearchStore((s) => s.setOpen);
  const { t } = useTranslation("common");

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
      // Выходим даже при ошибке API
    }
    storeLogout();
    void navigate({ to: "/login" });
  }

  return (
    <header className="relative flex h-14 items-center justify-between border-b border-border/50 bg-background/80 backdrop-blur-sm px-4 gradient-border-bottom">
      {/* Левая часть */}
      <div className="flex items-center gap-2">
        {/* Hamburger для мобильных */}
        {onMenuToggle && (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 md:hidden"
            onClick={onMenuToggle}
            aria-label={t("topbar.menu")}
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M2 4h12M2 8h12M2 12h12" /></svg>
          </Button>
        )}
      </div>

      {/* Правая часть */}
      <div className="flex items-center gap-2">
        {/* Кнопка глобального поиска — стиль glass */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="hidden gap-2 text-muted-foreground sm:flex rounded-xl border border-border/50 bg-muted/50 hover:bg-muted hover:border-border"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="h-3.5 w-3.5" />
              <span className="text-xs">{t("btn.search")}</span>
              <kbd className="pointer-events-none ml-2 inline-flex h-5 items-center gap-1 rounded-md border border-border/50 bg-background px-1.5 text-[10px] font-medium text-muted-foreground">
                {t("search.hint")}
              </kbd>
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t("search.globalSearch")} ({t("search.hint")})</TooltipContent>
        </Tooltip>

        {/* Переключатель языка */}
        <LanguageSwitcher compact className="hidden sm:flex" />

        {/* Уведомления */}
        <NotificationBell />

        {/* Меню пользователя — аватар с градиентной рамкой */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-9 w-9 rounded-full p-0">
              <div className="rounded-full bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 p-[2px]">
                <Avatar className="h-8 w-8 border-2 border-background">
                  <AvatarImage src={user?.avatar_url ?? undefined} alt={user?.name ?? t("auth.user")} />
                  <AvatarFallback className="text-xs font-semibold bg-muted">{initials}</AvatarFallback>
                </Avatar>
              </div>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-48 rounded-xl shadow-card">
            <div className="px-2 py-1.5">
              <p className="text-sm font-medium">{user?.name ?? t("auth.user")}</p>
              <p className="text-xs text-muted-foreground">{user?.email}</p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => void navigate({ to: "/settings/profile" })}>
              {t("auth.profile")}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              {t("auth.signOut")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
