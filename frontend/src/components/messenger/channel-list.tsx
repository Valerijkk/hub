import { useMatchRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useChannels } from "@/hooks/use-channels";
import { useUnread } from "@/hooks/use-unread";
import { usePresence } from "@/hooks/use-presence";
import { useMessengerStore } from "@/stores/messenger-store";
import { useAuthStore } from "@/stores/auth-store";
import { createChannel } from "@/api/messenger";
import type { Channel, ChannelType } from "@/api/messenger";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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

/** Иконка типа канала. */
function ChannelIcon({ type }: { type: ChannelType }) {
  switch (type) {
    case "dm":
      return (
        <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
      );
    case "group":
      return (
        <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
        </svg>
      );
    case "project":
      return (
        <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
      );
    case "public":
      return (
        <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
        </svg>
      );
  }
}

/** Группировка каналов по типу. */
function groupChannels(channels: Channel[]) {
  const groups: Record<string, { label: string; channels: Channel[] }> = {
    dm: { label: "Личные", channels: [] },
    group: { label: "Групповые", channels: [] },
    project: { label: "Проекты", channels: [] },
    public: { label: "Публичные", channels: [] },
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
        <Button variant="ghost" size="icon" className="h-6 w-6">
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
        </Button>
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
        className="inline-flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground"
        title="Закладки"
        onClick={() => setOpen(true)}
      >
        <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
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
 */
export function ChannelList() {
  const { data: channels, isLoading } = useChannels();
  useUnread();
  const currentUserId = useAuthStore((s) => s.user?.id);
  const unreadCounts = useMessengerStore((s) => s.unreadCounts);
  const matchRoute = useMatchRoute();

  // Собираем ID собеседников DM для отслеживания присутствия.
  const dmUserIds = (channels ?? [])
    .filter((ch) => ch.channel_type === "dm" && ch.members)
    .flatMap((ch) =>
      (ch.members ?? [])
        .filter((m) => m.user_id !== currentUserId)
        .map((m) => m.user_id),
    );
  const { getStatus } = usePresence(dmUserIds);

  /** Получить ID собеседника в DM. */
  function getDmPartnerId(channel: Channel): string | undefined {
    return channel.members?.find((m) => m.user_id !== currentUserId)?.user_id;
  }

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <span className="text-sm text-muted-foreground">Загрузка...</span>
      </div>
    );
  }

  const grouped = groupChannels(channels ?? []);

  return (
    <div className="flex h-full flex-col">
      {/* Заголовок */}
      <div className="flex h-12 items-center justify-between border-b px-4">
        <span className="text-sm font-semibold">Каналы</span>
        <div className="flex items-center gap-1">
          <BookmarksButton />
          <CreateChannelDialog />
        </div>
      </div>

      {/* Список каналов */}
      <ScrollArea className="flex-1">
        <div className="p-2 space-y-4">
          {grouped.map(([type, group]) => (
            <div key={type}>
              <p className="mb-1 px-2 text-xs font-semibold uppercase text-muted-foreground">
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
                        "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors hover:bg-accent",
                        isActive && "bg-accent font-medium",
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
                        <Badge
                          variant="destructive"
                          className="h-5 min-w-5 justify-center px-1 text-xs"
                        >
                          {unread > 99 ? "99+" : unread}
                        </Badge>
                      )}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}

          {grouped.length === 0 && (
            <p className="px-2 py-4 text-center text-xs text-muted-foreground">
              Нет каналов
            </p>
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
