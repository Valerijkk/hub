import { useQueryClient } from "@tanstack/react-query";
import { Link, useMatchRoute } from "@tanstack/react-router";
import { Search } from "lucide-react";
import { useCallback, useMemo, useState } from "react";
import type { Channel, PresenceStatus } from "@/api/messenger";
import {
	deleteChannel,
	muteChannel,
	starChannel,
	unmuteChannel,
} from "@/api/messenger";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
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

// ---------------------------------------------------------------------------
// Типы и константы
// ---------------------------------------------------------------------------

/** Варианты длительности мьюта канала. */
const MUTE_OPTIONS = [
	{ duration: "1h" as const, label: "1 час" },
	{ duration: "2h" as const, label: "2 часа" },
	{ duration: "8h" as const, label: "8 часов" },
	{ duration: "1d" as const, label: "1 день" },
	{ duration: "forever" as const, label: "Навсегда" },
] as const;

// ---------------------------------------------------------------------------
// Утилиты
// ---------------------------------------------------------------------------

/** Группирует каналы: закреплённые → личные → каналы (pub+priv+group). */
function groupChannels(channels: Channel[]) {
	const starred: Channel[] = [];
	const dm: Channel[] = [];
	const rest: Channel[] = [];

	for (const ch of channels) {
		if (ch.is_starred) {
			starred.push(ch);
		} else if (ch.channel_type === "dm") {
			dm.push(ch);
		} else {
			rest.push(ch);
		}
	}

	const groups: [string, { label: string; channels: Channel[] }][] = [];
	if (starred.length > 0)
		groups.push(["starred", { label: "Закреплённые", channels: starred }]);
	if (dm.length > 0) groups.push(["dm", { label: "Личные", channels: dm }]);
	if (rest.length > 0)
		groups.push(["channels", { label: "Каналы", channels: rest }]);
	return groups;
}

/** Проверяет, замьючен ли канал (muted_until в будущем). */
function isChannelMuted(channel: Channel): boolean {
	if (!channel.muted_until) return false;
	return new Date(channel.muted_until) > new Date();
}

// ---------------------------------------------------------------------------
// Подкомпоненты
// ---------------------------------------------------------------------------

/** Кнопка закладок, открывающая диалог со списком сохранённых сообщений. */
function BookmarksButton() {
	const [open, setOpen] = useState(false);
	return (
		<>
			<button
				type="button"
				className="w-[34px] h-[34px] flex items-center justify-center rounded-[var(--r2)] text-[var(--t3)] hover:bg-[var(--hover)] hover:text-[var(--t2)] hover:scale-[1.08] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]"
				title="Закладки"
				onClick={() => setOpen(true)}
			>
				<svg
					aria-hidden="true"
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
	if (type === "starred") return null;
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
			<div className="flex items-center justify-between px-4 py-1.5 group/section">
				<button
					type="button"
					className="flex items-center gap-1 text-[10px] font-bold uppercase tracking-[0.08em] text-[var(--t4)] hover:text-[var(--t3)] transition-colors"
					onClick={() => setCollapsed(!collapsed)}
				>
					<svg
						aria-hidden="true"
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

/** Иконка звезды (star) для закрепления канала. */
function StarIcon({ filled }: { filled: boolean }) {
	return (
		<svg
			aria-hidden="true"
			className={cn(
				"h-3.5 w-3.5",
				filled ? "text-yellow-400" : "text-sidebar-foreground/40",
			)}
			fill={filled ? "currentColor" : "none"}
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={2}
		>
			<path
				strokeLinecap="round"
				strokeLinejoin="round"
				d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
			/>
		</svg>
	);
}

/** Иконка отключённого звонка (мьют). */
function MutedIcon() {
	return (
		<svg
			aria-hidden="true"
			className="h-3.5 w-3.5 text-sidebar-foreground/40"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
			strokeWidth={2}
		>
			<path
				strokeLinecap="round"
				strokeLinejoin="round"
				d="M15.536 8.464a5 5 0 010 7.072M12 9.354V4.646a.5.5 0 00-.82-.385L7.232 7.5H4.5A.5.5 0 004 8v8a.5.5 0 00.5.5h2.732l3.948 3.239a.5.5 0 00.82-.385v-4.708M17.5 12h.01M3 3l18 18"
			/>
		</svg>
	);
}

/** Контекстное меню мьюта канала (появляется по правому клику). */
function MuteContextMenu({
	channelId,
	isMuted,
	position,
	onClose,
}: {
	channelId: string;
	isMuted: boolean;
	position: { x: number; y: number };
	onClose: () => void;
}) {
	const queryClient = useQueryClient();
	const [showDeleteDialog, setShowDeleteDialog] = useState(false);

	const handleMute = useCallback(
		async (duration: "1h" | "2h" | "8h" | "1d" | "forever") => {
			await muteChannel(channelId, duration);
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "channels"],
			});
			onClose();
		},
		[channelId, queryClient, onClose],
	);

	const handleUnmute = useCallback(async () => {
		await unmuteChannel(channelId);
		void queryClient.invalidateQueries({
			queryKey: ["messenger", "channels"],
		});
		onClose();
	}, [channelId, queryClient, onClose]);

	const handleDelete = useCallback(async () => {
		try {
			await deleteChannel(channelId);
			await queryClient.invalidateQueries({
				queryKey: ["messenger", "channels"],
			});
			await queryClient.invalidateQueries({ queryKey: ["messenger"] });
		} catch {
			// канал уже удалён
		}
		onClose();
	}, [channelId, queryClient, onClose]);

	return (
		<>
			{/* Невидимый оверлей для закрытия меню. */}
			<div className="fixed inset-0 z-40" onClick={onClose} />
			<div
				className="fixed z-50 min-w-[160px] rounded-md border bg-popover p-1 shadow-lg animate-in fade-in-0 zoom-in-95"
				style={{ left: position.x, top: position.y }}
			>
				{isMuted ? (
					<button
						type="button"
						className="flex w-full items-center rounded-sm px-2 py-1.5 text-xs hover:bg-accent transition-colors"
						onClick={() => void handleUnmute()}
					>
						Снять мьют
					</button>
				) : (
					<>
						<div className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-t2">
							Замьютить на
						</div>
						{MUTE_OPTIONS.map((opt) => (
							<button
								key={opt.duration}
								type="button"
								className="flex w-full items-center rounded-sm px-2 py-1.5 text-xs hover:bg-accent transition-colors"
								onClick={() => void handleMute(opt.duration)}
							>
								{opt.label}
							</button>
						))}
					</>
				)}

				{/* Разделитель + удаление */}
				<div className="my-1 h-px bg-border/50" />
				<button
					type="button"
					className="flex w-full items-center rounded-sm px-2 py-1.5 text-xs text-destructive hover:bg-destructive/10 transition-colors"
					onClick={() => setShowDeleteDialog(true)}
				>
					Удалить канал
				</button>
			</div>
			<ConfirmDialog
				open={showDeleteDialog}
				onOpenChange={setShowDeleteDialog}
				title="Удалить канал?"
				description="Это действие необратимо. История сообщений канала будет удалена."
				confirmLabel="Удалить"
				destructive
				onConfirm={handleDelete}
			/>
		</>
	);
}

/**
 * Элемент канала в списке.
 * Показывает иконку, название, время, непрочитанные, звезду (на hover), мьют-индикатор.
 */
function ChannelItem({
	channel,
	isActive,
	unread,
	getStatus,
	getDmPartnerId,
	onSelect,
}: {
	channel: Channel;
	isActive: boolean;
	unread: number;
	getStatus: (userId: string) => PresenceStatus;
	getDmPartnerId: (ch: Channel) => string | undefined;
	onSelect?: () => void;
}) {
	const queryClient = useQueryClient();
	const [muteMenu, setMuteMenu] = useState<{ x: number; y: number } | null>(
		null,
	);
	const partnerId = getDmPartnerId(channel);
	const muted = isChannelMuted(channel);

	/** Переключить звезду (star). */
	const handleToggleStar = useCallback(
		async (e: React.MouseEvent) => {
			e.preventDefault();
			e.stopPropagation();
			await starChannel(channel.id);
			void queryClient.invalidateQueries({
				queryKey: ["messenger", "channels"],
			});
		},
		[channel.id, queryClient],
	);

	/** Открыть меню мьюта по правому клику. */
	const handleContextMenu = useCallback((e: React.MouseEvent) => {
		e.preventDefault();
		setMuteMenu({ x: e.clientX, y: e.clientY });
	}, []);

	return (
		<>
			<Link
				key={channel.id}
				to="/messenger/$channelId"
				params={{ channelId: channel.id }}
				className={cn(
					"group/item flex items-center gap-[11px] py-2.5 px-3.5 cursor-pointer transition-colors border-l-2 border-transparent",
					isActive && "bg-[rgba(52,211,153,.04)] border-l-[var(--green)]",
					!isActive && "hover:bg-[var(--hover)]",
				)}
				onClick={onSelect}
				onContextMenu={handleContextMenu}
			>
				{/* Icon with online dot for DM */}
				<div className="w-[38px] h-[38px] rounded-full flex items-center justify-center text-[13.5px] font-bold shrink-0 relative hover:scale-105 transition-transform">
					<ChannelIcon type={channel.channel_type} />
					{channel.channel_type === "dm" && partnerId && (
						<span className="absolute bottom-0 right-0 w-[9px] h-[9px] rounded-full border-2 border-[var(--void)]">
							<PresenceIndicator status={getStatus(partnerId)} size="sm" />
						</span>
					)}
				</div>

				{/* Name + preview + time */}
				<div className="flex-1 min-w-0">
					<div className="flex items-center justify-between gap-2">
						<span
							className={cn(
								"text-sm font-semibold truncate tracking-[-0.01em] text-[var(--t2)]",
								(unread > 0 || isActive) && "text-[var(--t1)]",
							)}
						>
							{channel.name}
						</span>
						<span className="text-[10.5px] text-[var(--t4)] font-medium shrink-0">
							{formatRelativeTime(channel.updated_at)}
						</span>
					</div>
				</div>

				{/* Muted indicator */}
				{muted && (
					<span className="shrink-0" title="Канал замьючен">
						<MutedIcon />
					</span>
				)}

				{/* Star button (visible on hover or if starred) */}
				<button
					type="button"
					className={cn(
						"shrink-0 transition-opacity",
						channel.is_starred
							? "opacity-100"
							: "opacity-0 group-hover/item:opacity-100",
					)}
					title={channel.is_starred ? "Убрать из закреплённых" : "Закрепить"}
					onClick={(e) => void handleToggleStar(e)}
				>
					<StarIcon filled={channel.is_starred} />
				</button>

				{/* Unread badge */}
				{unread > 0 && (
					<span className="bg-gradient-to-br from-[var(--green)] to-[var(--greenS)] text-white text-[10px] font-bold px-[7px] py-0.5 rounded-full min-w-[20px] text-center shadow-[0_2px_8px_var(--gG)]">
						{unread > 99 ? "99+" : unread}
					</span>
				)}
			</Link>

			{/* Контекстное меню мьюта */}
			{muteMenu && (
				<MuteContextMenu
					channelId={channel.id}
					isMuted={muted}
					position={muteMenu}
					onClose={() => setMuteMenu(null)}
				/>
			)}
		</>
	);
}

// ---------------------------------------------------------------------------
// Главный компонент
// ---------------------------------------------------------------------------

/**
 * ChannelList отображает список каналов в стиле Slack / Tinkoff Talk.
 * Каждый канал показывает: иконку, название, превью последнего сообщения,
 * относительное время, badge непрочитанных и индикатор онлайна для ЛС.
 * Закреплённые (starred) каналы отображаются в отдельной секции сверху.
 */
/** Пропсы списка каналов. */
interface ChannelListProps {
	/**
	 * Колбэк выбора канала. Используется на мобилке, чтобы закрыть drawer.
	 * На десктопе можно не передавать.
	 */
	onChannelSelect?: () => void;
}

export function ChannelList({ onChannelSelect }: ChannelListProps = {}) {
	const { data: channels, isLoading } = useChannels();
	useUnread();
	const unreadCounts = useMessengerStore((s) => s.unreadCounts);
	const matchRoute = useMatchRoute();
	const [searchQuery, setSearchQuery] = useState("");

	// DM presence tracking not available -- backend ChannelResponse does not include members list.
	const dmUserIds: string[] = [];
	const { getStatus } = usePresence(dmUserIds);

	const allChannels = channels ?? [];

	// Фильтрация каналов по поисковому запросу.
	// NOTE: useMemo must be called unconditionally (before any early return)
	// to comply with React's rules of hooks.
	const filteredChannels = useMemo(() => {
		if (!searchQuery.trim()) return allChannels;
		const q = searchQuery.toLowerCase();
		return allChannels.filter((ch) => ch.name.toLowerCase().includes(q));
	}, [allChannels, searchQuery]);

	/** Получить ID собеседника в ЛС (пока недоступно на бэкенде). */
	function getDmPartnerId(_channel: Channel): string | undefined {
		return undefined;
	}

	if (isLoading) {
		return (
			<div className="w-[280px] min-w-[280px] h-full border-r border-[var(--b1)] flex flex-col bg-[var(--void)] msg-enter-x">
				<div className="px-3.5 pt-3.5 pb-2">
					<span className="text-sm font-semibold text-[var(--t1)]">
						Мессенджер
					</span>
				</div>
				<div className="flex flex-1 items-center justify-center">
					<div className="space-y-1 w-full px-2">
						{[1, 2, 3, 4, 5].map((i) => (
							<div
								key={`ch-sk-${i}`}
								className="flex items-center gap-[11px] px-3.5 py-2.5"
							>
								<div className="h-[38px] w-[38px] rounded-full animate-shimmer shrink-0" />
								<div className="flex-1 space-y-1">
									<div className="h-3 w-24 rounded animate-shimmer" />
									<div className="h-2.5 w-16 rounded animate-shimmer" />
								</div>
							</div>
						))}
					</div>
				</div>
			</div>
		);
	}

	// Разделяем закреплённые и обычные каналы.
	const starredChannels = filteredChannels.filter((ch) => ch.is_starred);
	const regularChannels = filteredChannels.filter((ch) => !ch.is_starred);
	const grouped = groupChannels(regularChannels);

	/** Рендер одного элемента канала. */
	const renderChannelItem = (channel: Channel) => {
		const isActive = matchRoute({
			to: "/messenger/$channelId",
			params: { channelId: channel.id },
			fuzzy: true,
		});
		const unread = unreadCounts[channel.id] ?? 0;

		return (
			<ChannelItem
				key={channel.id}
				channel={channel}
				isActive={!!isActive}
				unread={unread}
				getStatus={getStatus}
				getDmPartnerId={getDmPartnerId}
				onSelect={onChannelSelect}
			/>
		);
	};

	return (
		<div className="w-[280px] min-w-[280px] h-full border-r border-[var(--b1)] flex flex-col bg-[var(--void)] msg-enter-x">
			{/* Header */}
			<div className="px-3.5 pt-3.5 pb-2">
				<div className="flex items-center justify-between mb-2">
					<span className="text-sm font-semibold tracking-wide text-[var(--t1)]">
						Мессенджер
					</span>
					<div className="flex items-center gap-1">
						<BookmarksButton />
					</div>
				</div>
				{/* Pill-shaped search input per mockup 03-messenger */}
				<div className="relative">
					<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-[15px] w-[15px] text-[var(--t3)] pointer-events-none" />
					<input
						type="text"
						className="w-full h-9 pl-[38px] pr-3.5 bg-[var(--input)] border border-transparent rounded-full text-[13px] text-[var(--t1)] placeholder:text-[var(--t3)] outline-none transition-all duration-250 focus:border-[rgba(52,211,153,.3)] focus:bg-[var(--card)] focus:shadow-[0_0_0_3px_rgba(52,211,153,.06)]"
						placeholder="Поиск..."
						value={searchQuery}
						onChange={(e) => setSearchQuery(e.target.value)}
					/>
				</div>
			</div>

			{/* Channel list */}
			<div className="flex-1 overflow-y-auto">
				<div className="py-1 space-y-2">
					{/* Закреплённые каналы (starred) */}
					{starredChannels.length > 0 && (
						<ChannelSection type="starred" label="Закреплённые">
							{starredChannels.map(renderChannelItem)}
						</ChannelSection>
					)}

					{/* Обычные каналы по группам */}
					{grouped.map(([type, group]) => (
						<ChannelSection key={type} type={type} label={group.label}>
							{group.channels.map(renderChannelItem)}
						</ChannelSection>
					))}

					{allChannels.length === 0 && (
						<div className="flex flex-col items-center justify-center py-12 px-4">
							<svg
								aria-hidden="true"
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

			{/* Bottom panel -- create channel + new DM */}
			<div className="border-t border-[var(--b1)] px-3.5 h-12 flex items-center gap-2">
				<NewDmDialog />
				<Dialog>
					<DialogTrigger asChild>
						<button
							type="button"
							className="flex flex-1 items-center justify-center gap-1.5 rounded-lg px-2 py-1.5 text-xs font-medium text-[var(--t3)] hover:bg-[var(--hover)] hover:text-[var(--t2)] transition-colors"
						>
							<svg
								aria-hidden="true"
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
							Канал
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
