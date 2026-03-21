/**
 * Верхняя панель приложения — чистый, минимальный дизайн.
 *
 * Тонкий бордер снизу, кнопка поиска с kbd-хинтом,
 * аватар пользователя с выпадающим меню.
 * На мобильных — hamburger для открытия sidebar.
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
      /* Выходим даже при ошибке API */
    }
    storeLogout();
    void navigate({ to: "/login" });
  }

  return (
    <header className="flex h-14 items-center justify-between border-b bg-card/80 backdrop-blur-sm px-4 lg:px-6">
      {/* Левая часть */}
      <div className="flex items-center gap-2">
        {/* Hamburger для мобильных */}
        {onMenuToggle && (
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 md:hidden text-muted-foreground"
            onClick={onMenuToggle}
            aria-label={t("topbar.menu")}
          >
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M2 4h12M2 8h12M2 12h12" /></svg>
          </Button>
        )}
      </div>

      {/* Правая часть */}
      <div className="flex items-center gap-1.5">
        {/* Кнопка глобального поиска */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="ghost"
              size="sm"
              className="hidden gap-2 text-muted-foreground sm:flex rounded-lg border border-border bg-muted/40 hover:bg-muted px-3 h-8"
              onClick={() => setSearchOpen(true)}
            >
              <Search className="h-3.5 w-3.5" />
              <span className="text-xs">{t("btn.search")}</span>
              <kbd className="pointer-events-none ml-2 inline-flex h-5 items-center rounded border border-border bg-background px-1.5 text-[10px] font-mono font-medium text-muted-foreground">
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

        {/* Меню пользователя */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8 rounded-full p-0 ml-1">
              <Avatar className="h-7 w-7">
                <AvatarImage src={user?.avatar_url ?? undefined} alt={user?.name ?? t("auth.user")} />
                <AvatarFallback className="text-[11px] font-semibold bg-primary/10 text-primary">
                  {initials}
                </AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-52 rounded-lg">
            <div className="px-3 py-2">
              <p className="text-sm font-medium">{user?.name ?? t("auth.user")}</p>
              <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={() => void navigate({ to: "/settings/profile" })}>
              {t("auth.profile")}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              onClick={handleLogout}
              className="text-destructive focus:text-destructive"
            >
              {t("auth.signOut")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
