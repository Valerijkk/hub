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
		dm: { label: "Личные сообщения", channels: [] },
		group: { label: "Групповые", channels: [] },
		private: { label: "Приватные", channels: [] },
		public: { label: "Каналы", channels: [] },
	};
	for (const ch of channels) {
		const group = groups[ch.channel_type];
		if (group) {
			group.channels.push(ch);
		} else {
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

/** Кнопка "+" для создания канала рядом с заголовком секции. */
function SectionAddButton({ type }: { type: string }) {
	if (type === "dm") return <NewDmDialog />;
	return <CreateChannelDialog />;
}

/** Сворачиваемая секция каналов. */
function ChannelSection({
	type,
	label,
	children,
}: {
	type: string;
	label: string;
	children: React.ReactNode;
}) {
	const [collapsed, setCollapsed] = useState(false);

	return (
		<div>
			<div className="flex items-center justify-between px-3 py-1.5 group/section">
				<button
					type="button"
					className="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-wider text-sidebar-foreground/50 hover:text-sidebar-foreground/80 transition-colors"
					onClick={() => setCollapsed(!collapsed)}
				>
					<svg
						className={cn(
							"h-3 w-3 transition-transform duration-200",
							collapsed && "-rotate-90",
						)}
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
						strokeWidth={2}
					>
						<path
							strokeLinecap="round"
							strokeLinejoin="round"
							d="M19 9l-7 7-7-7"
						/>
					</svg>
					{label}
				</button>
				<div className="opacity-0 group-hover/section:opacity-100 transition-opacity">
					<SectionAddButton type={type} />
				</div>
			</div>
			{!collapsed && <div className="space-y-px">{children}</div>}
		</div>
	);
}

/**
 * ChannelList отображает список каналов в стиле Slack / Tinkoff Talk.
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
						Мессенджер
					</span>
				</div>
				<div className="flex flex-1 items-center justify-center">
					<div className="space-y-1 w-full px-2">
						{[1, 2, 3, 4, 5].map((i) => (
							<div key={i} className="flex items-center gap-2.5 px-3 py-2">
								<div className="h-5 w-5 rounded animate-shimmer shrink-0" />
								<div className="flex-1 space-y-1">
									<div className="h-3 w-24 rounded animate-shimmer" />
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
					Мессенджер
				</span>
				<div className="flex items-center gap-0.5">
					<BookmarksButton />
				</div>
			</div>

			{/* Channel list */}
			<div className="flex-1 overflow-y-auto">
				<div className="py-3 space-y-4">
					{grouped.map(([type, group]) => (
						<ChannelSection key={type} type={type} label={group.label}>
							{group.channels.map((channel) => {
								const isActive = matchRoute({
									to: "/messenger/$channelId",
									params: { channelId: channel.id },
									fuzzy: true,
								});
								const unread = unreadCounts[channel.id] ?? 0;
								const partnerId = getDmPartnerId(channel);

								return (
									<Link
										key={channel.id}
										to="/messenger/$channelId"
										params={{ channelId: channel.id }}
										className={cn(
											"group/item relative flex items-center gap-2.5 mx-2 rounded-md px-2.5 py-1.5 text-sm transition-all duration-150",
											"text-sidebar-foreground/70 hover:bg-white/[0.06] hover:text-sidebar-foreground",
											isActive &&
												"bg-primary/10 text-sidebar-foreground font-medium",
										)}
									>
										{/* Left accent bar for active channel */}
										{isActive && (
											<div className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-primary" />
										)}

										{/* Icon with online dot for DM */}
										<div className="relative shrink-0">
											<ChannelIcon type={channel.channel_type} />
											{channel.channel_type === "dm" && partnerId && (
												<span className="absolute -bottom-0.5 -right-0.5">
													<PresenceIndicator
														status={getStatus(partnerId)}
														size="sm"
														className="ring-2 ring-sidebar"
													/>
												</span>
											)}
										</div>

										{/* Name + time */}
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
												<span className="shrink-0 text-[10px] text-sidebar-foreground/35">
													{formatRelativeTime(channel.updated_at)}
												</span>
											</div>
										</div>

										{/* Unread dot / count */}
										{unread > 0 && (
											<span className="inline-flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">
												{unread > 99 ? "99+" : unread}
											</span>
										)}
									</Link>
								);
							})}
						</ChannelSection>
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
			</div>

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
