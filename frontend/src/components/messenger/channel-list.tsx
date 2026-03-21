import { useMatchRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useChannels } from "@/hooks/use-channels";
import { useUnread } from "@/hooks/use-unread";
import { usePresence } from "@/hooks/use-presence";
import { useMessengerStore } from "@/stores/messenger-store";

import { createChannel } from "@/api/messenger";
import type { Channel, ChannelType } from "@/api/messenger";
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

/** Иконка типа канала — стилизованная для тёмного сайдбара. */
function ChannelIcon({ type }: { type: ChannelType }) {
  switch (type) {
    case "dm":
      return (
        <svg className="h-4 w-4 shrink-0 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
      );
    case "group":
      return (
        <svg className="h-4 w-4 shrink-0 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
      );
    case "project":
      return (
        <svg className="h-4 w-4 shrink-0 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      );
    case "public":
      return (
        <span className="shrink-0 text-sm opacity-70 font-medium">#</span>
      );
  }
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
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
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
        <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
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

/**
 * ChannelList отображает список каналов пользователя с группировкой по типу,
 * badge непрочитанных и кнопкой создания нового канала.
 * Тёмный стиль сайдбара, соответствующий основному сайдбару.
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
          <span className="text-sm font-semibold text-sidebar-foreground">Каналы</span>
        </div>
        <div className="flex flex-1 items-center justify-center">
          {/* Скелетон загрузки */}
          <div className="space-y-3 w-full px-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-8 rounded-md animate-shimmer" />
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
        <span className="text-sm font-semibold tracking-wide text-sidebar-foreground">Каналы</span>
        <div className="flex items-center gap-0.5">
          <BookmarksButton />
          <CreateChannelDialog />
        </div>
      </div>

      {/* Список каналов */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-5">
          {grouped.map(([type, group]) => (
            <div key={type}>
              {/* Заголовок секции — uppercase, мелкий, приглушённый */}
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
                        "text-sidebar-foreground/70 hover:bg-white/[0.08] hover:text-sidebar-foreground hover:translate-x-0.5",
                        isActive && "bg-primary/15 text-sidebar-foreground border-l-[3px] border-primary font-medium shadow-[0_0_10px_rgba(59,130,246,0.1)]",
                      )}
                    >
                      <ChannelIcon type={channel.channel_type} />
                      {channel.channel_type === "dm" && (() => {
                        const partnerId = getDmPartnerId(channel);
                        return partnerId ? (
                          <PresenceIndicator status={getStatus(partnerId)} />
                        ) : null;
                      })()}
                      <span className="flex-1 truncate">{channel.name}</span>
                      {unread > 0 && (
                        <span className="inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-[10px] font-bold text-primary-foreground">
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
              <svg className="mb-3 h-10 w-10 text-sidebar-foreground/20" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <p className="text-xs text-sidebar-foreground/40 text-center">Нет каналов</p>
            </div>
          )}
        </div>
      </ScrollArea>

      {/* Нижняя панель — создать канал */}
      <div className="border-t border-white/10 p-3">
        <Dialog>
          <DialogTrigger asChild>
            <button
              type="button"
              className="flex w-full items-center justify-center gap-2 rounded-md px-3 py-2 text-xs font-medium text-sidebar-foreground/50 hover:bg-white/8 hover:text-sidebar-foreground transition-colors"
            >
              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
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
