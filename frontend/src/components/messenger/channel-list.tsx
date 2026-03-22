import { Link, useMatchRoute } from "@tanstack/react-router";
import { useState } from "react";
import type { Channel } from "@/api/messenger";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useChannels } from "@/hooks/use-channels";
import { usePresence } from "@/hooks/use-presence";
import { useUnread } from "@/hooks/use-unread";
import { formatRelativeTime } from "@/lib/format-time";
import { cn } from "@/lib/utils";
import { useMessengerStore } from "@/stores/messenger-store";
import { BookmarksPanel } from "./bookmarks-panel";
import { ChannelIcon } from "./channel-icon";
import {
	CreateChannelDialog,
	CreateChannelDialogContent,
} from "./create-channel-dialog";
import { NewDmDialog } from "./new-dm-dialog";
import { PresenceIndicator } from "./presence-indicator";

/** Группирует каналы по типу (личные, групповые, проекты, публичные). */
function groupChannels(channels: Channel[]) {
	const groups: Record<string, { label: string; channels: Channel[] }> = {
		dm: { label: "ЛИЧНЫЕ", channels: [] },
		group: { label: "ГРУППОВЫЕ", channels: [] },
		private: { label: "ПРИВАТНЫЕ", channels: [] },
		public: { label: "ПУБЛИЧНЫЕ", channels: [] },
	};
	for (const ch of channels) {
		const group = groups[ch.channel_type];
		if (group) {
			group.channels.push(ch);
		} else {
			// Unknown channel type -- put in group to avoid losing it
			groups.group?.channels.push(ch);
		}
	}
	return Object.entries(groups).filter(([, g]) => g.channels.length > 0);
}

/** Кнопка закладок, открывающая диалог со списком сохранённых сообщений. */
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

/**
 * ChannelList отображает список каналов в стиле Tinkoff Talk.
 * Каждый канал показывает: иконку, название, превью последнего сообщения,
 * относительное время, badge непрочитанных и индикатор онлайна для ЛС.
 */
export function ChannelList() {
	const { data: channels, isLoading } = useChannels();
	useUnread();
	const unreadCounts = useMessengerStore((s) => s.unreadCounts);
	const matchRoute = useMatchRoute();

	// DM presence tracking not available -- backend ChannelResponse does not include members list.
	const dmUserIds: string[] = [];
	const { getStatus } = usePresence(dmUserIds);

	/** Получить ID собеседника в ЛС (пока недоступно на бэкенде). */
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
			{/* Header */}
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

			{/* Channel list */}
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
											{/* Icon with online dot for DM */}
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

											{/* Name + last message preview */}
											<div className="flex-1 min-w-0">
												<div className="flex items-center justify-between gap-2">
													<span
														className={cn(
															"truncate text-[13px]",
															unread > 0 &&
																"font-semibold text-sidebar-foreground",
														)}
													>
														{channel.name}
													</span>
													{/* Relative time */}
													<span className="shrink-0 text-[10px] text-sidebar-foreground/40">
														{formatRelativeTime(channel.updated_at)}
													</span>
												</div>
											</div>

											{/* Unread badge -- gradient pill */}
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

			{/* Bottom panel -- create channel */}
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
