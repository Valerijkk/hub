import { useMatchRoute, useNavigate, Link } from "@tanstack/react-router";
import { useState, useEffect, useMemo } from "react";
import { useChannels } from "@/hooks/use-channels";
import { useUnread } from "@/hooks/use-unread";
import { usePresence } from "@/hooks/use-presence";
import { useMessengerStore } from "@/stores/messenger-store";
import { useAuthStore } from "@/stores/auth-store";

import { createChannel, createOrGetDM } from "@/api/messenger";
import type { Channel, ChannelType } from "@/api/messenger";
import { api } from "@/api/client";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { PresenceIndicator } from "./presence-indicator";
import { BookmarksPanel } from "./bookmarks-panel";
import { cn } from "@/lib/utils";

/** Иконка типа канала -- стилизованная для тёмного сайдбара. */
function ChannelIcon({ type }: { type: ChannelType }) {
  switch (type) {
    case "dm":
      return (
        <svg
          className="h-4 w-4 shrink-0 opacity-70"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
          />
        </svg>
      );
    case "group":
      return (
        <svg
          className="h-4 w-4 shrink-0 opacity-70"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
      );
    case "project":
      return (
        <svg
          className="h-4 w-4 shrink-0 opacity-70"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
          />
        </svg>
      );
    case "public":
      return (
        <span className="shrink-0 text-sm opacity-70 font-medium">#</span>
      );
  }
}

/**
 * Форматирует относительное время для списка каналов.
 * "14:35" (сегодня), "Вчера", "Пн", "15 мар" и т.д.
 */
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const msgDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());

  if (msgDate.getTime() === today.getTime()) {
    return date.toLocaleTimeString("ru-RU", {
      hour: "2-digit",
      minute: "2-digit",
    });
  }
  if (msgDate.getTime() === yesterday.getTime()) return "Вчера";

  const diffDays = Math.floor(
    (today.getTime() - msgDate.getTime()) / (1000 * 60 * 60 * 24),
  );
  if (diffDays < 7) {
    return date.toLocaleDateString("ru-RU", { weekday: "short" });
  }
  return date.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}

/** Группировка каналов по типу. */
function groupChannels(channels: Channel[]) {
  const groups: Record<string, { label: string; channels: Channel[] }> = {
    dm: { label: "ЛИЧНЫЕ", channels: [] },
    group: { label: "ГРУППОВЫЕ", channels: [] },
    project: { label: "ПРОЕКТЫ", channels: [] },
    public: { label: "ПУБЛИЧНЫЕ", channels: [] },
  };
  for (const ch of channels) {
    groups[ch.channel_type]?.channels.push(ch);
  }
  return Object.entries(groups).filter(([, g]) => g.channels.length > 0);
}

/** Диалог создания нового канала. */
function CreateChannelDialog() {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<ChannelType>("group");
  const [isLoading, setIsLoading] = useState(false);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setIsLoading(true);
    try {
      await createChannel({ name: name.trim(), channel_type: type });
      setOpen(false);
      setName("");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button
          type="button"
          className="inline-flex h-7 w-7 items-center justify-center rounded-md text-sidebar-foreground/60 hover:bg-white/10 hover:text-sidebar-foreground transition-colors"
          title="Создать канал"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 4v16m8-8H4"
            />
          </svg>
        </button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Создать канал</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 pt-2">
          <Input
            placeholder="Название канала"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <div className="flex gap-2">
            {(["group", "public"] as const).map((t) => (
              <Button
                key={t}
                variant={type === t ? "default" : "outline"}
                size="sm"
                onClick={() => setType(t)}
              >
                {t === "group" ? "Групповой" : "Публичный"}
              </Button>
            ))}
          </div>
          <Button
            className="w-full"
            onClick={handleCreate}
            disabled={!name.trim() || isLoading}
          >
            {isLoading ? "Создание..." : "Создать"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

/** Кнопка открытия панели закладок в диалоге. */
function BookmarksButton() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button
        type="button"
        className="inline-flex h-7 w-7 items-center justify-center rounded-md text-sidebar-foreground/60 hover:bg-white/10 hover:text-sidebar-foreground transition-colors"
        title="Закладки"
        onClick={() => setOpen(true)}
      >
        <svg
          className="h-4 w-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={2}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
          />
        </svg>
      </button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-md max-h-[80vh]">
          <DialogHeader>
            <DialogTitle>Закладки</DialogTitle>
          </DialogHeader>
          <BookmarksList />
        </DialogContent>
      </Dialog>
    </>
  );
}

/** Содержимое панели закладок. */
function BookmarksList() {
  return (
    <div className="py-2">
      <BookmarksPanel />
    </div>
  );
}

/** Пользователь для выбора в диалоге написания ЛС. */
interface PickerUser {
  id: string;
  email: string;
  name?: string;
  display_name?: string;
}

/** Ответ списка пользователей от /admin/users. */
interface UsersListResponse {
  items?: PickerUser[];
  users?: PickerUser[];
}

/**
 * Диалог "Написать" -- выбрать пользователя и открыть/создать DM-канал.
 */
function NewDmDialog() {
  const [open, setOpen] = useState(false);
  const [users, setUsers] = useState<PickerUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [creating, setCreating] = useState<string | null>(null);
  const navigate = useNavigate();
  const currentUserId = useAuthStore((s) => s.user?.id);

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    api
      .get<UsersListResponse | PickerUser[]>("/admin/users")
      .then((data) => {
        let list: PickerUser[] = [];
        if (Array.isArray(data)) {
          list = data;
        } else if (data.items) {
          list = data.items;
        } else if (data.users) {
          list = data.users;
        }
        setUsers(list.filter((u) => u.id !== currentUserId));
      })
      .catch(() => setUsers([]))
      .finally(() => setLoading(false));
  }, [open, currentUserId]);

  const filtered = useMemo(() => {
    if (!search.trim()) return users;
    const q = search.toLowerCase();
    return users.filter(
      (u) =>
        (u.name && u.name.toLowerCase().includes(q)) ||
        (u.display_name && u.display_name.toLowerCase().includes(q)) ||
        u.email.toLowerCase().includes(q),
    );
  }, [users, search]);

  const displayName = (u: PickerUser) =>
    u.display_name || u.name || u.email;

  const handleSelect = async (userId: string) => {
    setCreating(userId);
    try {
      const channel = await createOrGetDM(userId);
      setOpen(false);
      setSearch("");
      void navigate({
        to: "/messenger/$channelId",
        params: { channelId: channel.id },
      });
    } finally {
      setCreating(null);
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <button
          type="button"
          className="inline-flex h-7 w-7 items-center justify-center rounded-md text-sidebar-foreground/60 hover:bg-white/10 hover:text-sidebar-foreground transition-colors"
          title="Написать"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
            />
          </svg>
        </button>
      </DialogTrigger>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Написать</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 pt-1">
          <Input
            placeholder="Поиск по имени или email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoFocus
          />
          <ScrollArea className="max-h-64">
            {loading ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                Загрузка...
              </div>
            ) : filtered.length === 0 ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                {search ? "Никого не найдено" : "Нет пользователей"}
              </div>
            ) : (
              <div className="space-y-0.5">
                {filtered.map((u) => (
                  <button
                    key={u.id}
                    type="button"
                    disabled={creating === u.id}
                    onClick={() => handleSelect(u.id)}
                    className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors disabled:opacity-50"
                  >
                    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                      {displayName(u).charAt(0).toUpperCase()}
                    </div>
                    <div className="flex flex-col items-start overflow-hidden">
                      <span className="truncate font-medium">
                        {displayName(u)}
                      </span>
                      {u.email && (
                        <span className="truncate text-xs text-muted-foreground">
                          {u.email}
                        </span>
                      )}
                    </div>
                    {creating === u.id && (
                      <span className="ml-auto text-xs text-muted-foreground">
                        ...
                      </span>
                    )}
                  </button>
                ))}
              </div>
            )}
          </ScrollArea>
        </div>
      </DialogContent>
    </Dialog>
  );
}

/**
 * ChannelList отображает список каналов в стиле Tinkoff Talk.
 * Каждый канал показывает: иконку, имя, последнее сообщение (превью),
 * относительное время, бейдж непрочитанных и online-точку для DM.
 */
export function ChannelList() {
  const { data: channels, isLoading } = useChannels();
  useUnread();
  const unreadCounts = useMessengerStore((s) => s.unreadCounts);
  const matchRoute = useMatchRoute();

  // DM presence tracking not available -- backend ChannelResponse does not include members list.
  const dmUserIds: string[] = [];
  const { getStatus } = usePresence(dmUserIds);

  /** Получить ID собеседника в DM (not available with current backend). */
  function getDmPartnerId(_channel: Channel): string | undefined {
    return undefined;
  }

  if (isLoading) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex h-14 items-center justify-between border-b border-white/10 px-4">
          <span className="text-sm font-semibold text-sidebar-foreground">
            Каналы
          </span>
        </div>
        <div className="flex flex-1 items-center justify-center">
          <div className="space-y-2 w-full px-3">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="flex items-center gap-3 px-3 py-2.5">
                <div className="h-10 w-10 rounded-full animate-shimmer shrink-0" />
                <div className="flex-1 space-y-1.5">
                  <div className="h-3 w-24 rounded animate-shimmer" />
                  <div className="h-2.5 w-36 rounded animate-shimmer" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  const grouped = groupChannels(channels ?? []);

  return (
    <div className="flex h-full flex-col">
      {/* Заголовок */}
      <div className="flex h-14 items-center justify-between border-b border-white/10 px-4">
        <span className="text-sm font-semibold tracking-wide text-sidebar-foreground">
          Каналы
        </span>
        <div className="flex items-center gap-0.5">
          <NewDmDialog />
          <BookmarksButton />
          <CreateChannelDialog />
        </div>
      </div>

      {/* Список каналов */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-5">
          {grouped.map(([type, group]) => (
            <div key={type}>
              <p className="mb-1.5 px-3 text-[10px] font-bold uppercase tracking-widest text-sidebar-foreground/40">
                {group.label}
              </p>
              <div className="space-y-0.5">
                {group.channels.map((channel) => {
                  const isActive = matchRoute({
                    to: "/messenger/$channelId",
                    params: { channelId: channel.id },
                    fuzzy: true,
                  });
                  const unread = unreadCounts[channel.id] ?? 0;

                  return (
                    <Link
                      key={channel.id}
                      to="/messenger/$channelId"
                      params={{ channelId: channel.id }}
                      className={cn(
                        "flex items-center gap-2.5 rounded-lg px-3 py-2.5 text-sm transition-all duration-200",
                        "text-sidebar-foreground/70 hover:bg-white/[0.08] hover:text-sidebar-foreground",
                        isActive &&
                          "bg-primary/15 text-sidebar-foreground border-l-[3px] border-primary font-medium shadow-[0_0_10px_rgba(59,130,246,0.1)]",
                      )}
                    >
                      {/* Иконка с online-точкой для DM */}
                      <div className="relative shrink-0">
                        <ChannelIcon type={channel.channel_type} />
                        {channel.channel_type === "dm" &&
                          (() => {
                            const partnerId = getDmPartnerId(channel);
                            return partnerId ? (
                              <span className="absolute -bottom-0.5 -right-0.5">
                                <PresenceIndicator
                                  status={getStatus(partnerId)}
                                  size="sm"
                                  className="ring-2 ring-sidebar"
                                />
                              </span>
                            ) : null;
                          })()}
                      </div>

                      {/* Название + превью последнего сообщения */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between gap-2">
                          <span
                            className={cn(
                              "truncate text-[13px]",
                              unread > 0 && "font-semibold text-sidebar-foreground",
                            )}
                          >
                            {channel.name}
                          </span>
                          {/* Относительное время */}
                          <span className="shrink-0 text-[10px] text-sidebar-foreground/40">
                            {formatRelativeTime(channel.updated_at)}
                          </span>
                        </div>
                      </div>

                      {/* Бейдж непрочитанных -- градиентная пилюля */}
                      {unread > 0 && (
                        <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full gradient-messenger px-1.5 text-[10px] font-bold text-white shadow-sm">
                          {unread > 99 ? "99+" : unread}
                        </span>
                      )}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}

          {grouped.length === 0 && (
            <div className="flex flex-col items-center justify-center py-12 px-4">
              <svg
                className="mb-3 h-10 w-10 text-sidebar-foreground/20"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={1}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                />
              </svg>
              <p className="text-xs text-sidebar-foreground/40 text-center">
                Нет каналов
              </p>
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Нижняя панель -- создать канал */}
      <div className="border-t border-white/10 p-3">
        <Dialog>
          <DialogTrigger asChild>
            <button
              type="button"
              className="flex w-full items-center justify-center gap-2 rounded-md px-3 py-2 text-xs font-medium text-sidebar-foreground/50 hover:bg-white/8 hover:text-sidebar-foreground transition-colors"
            >
              <svg
                className="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 4v16m8-8H4"
                />
              </svg>
              Создать канал
            </button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Создать канал</DialogTitle>
            </DialogHeader>
            <CreateChannelDialogContent />
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
}

/** Содержимое диалога создания канала (для нижней кнопки). */
function CreateChannelDialogContent() {
  const [name, setName] = useState("");
  const [type, setType] = useState<ChannelType>("group");
  const [isLoading, setIsLoading] = useState(false);

  const handleCreate = async () => {
    if (!name.trim()) return;
    setIsLoading(true);
    try {
      await createChannel({ name: name.trim(), channel_type: type });
      setName("");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-4 pt-2">
      <Input
        placeholder="Название канала"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <div className="flex gap-2">
        {(["group", "public"] as const).map((t) => (
          <Button
            key={t}
            variant={type === t ? "default" : "outline"}
            size="sm"
            onClick={() => setType(t)}
          >
            {t === "group" ? "Групповой" : "Публичный"}
          </Button>
        ))}
      </div>
      <Button
        className="w-full"
        onClick={handleCreate}
        disabled={!name.trim() || isLoading}
      >
        {isLoading ? "Создание..." : "Создать"}
      </Button>
    </div>
  );
}
