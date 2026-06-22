import { useNavigate } from "@tanstack/react-router";
import {
	AlertTriangle,
	AtSign,
	CheckCircle2,
	ClipboardList,
	Clock,
	Link2,
	MessageSquare,
	Phone,
	TestTube,
	X,
} from "lucide-react";
import { memo, useCallback, useRef } from "react";
import { toast } from "sonner";
import type { NotificationDTO } from "@/api/notifications";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import i18n from "@/i18n";
import { formatRelativeTime } from "@/lib/format-time";
import { useNotificationStore } from "@/stores/notification-store";

/** Карта иконок по типу уведомления. */
const TYPE_ICONS: Record<string, React.ElementType> = {
	assigned: ClipboardList,
	mentioned: AtSign,
	status_changed: CheckCircle2,
	comment_added: MessageSquare,
	message_received: MessageSquare,
	call_started: Phone,
	run_completed: TestTube,
	run_failed: AlertTriangle,
	deadline_approaching: Clock,
	blocker_added: Link2,
};

/** Цвета фона иконок по типу уведомления. */
const TYPE_ICON_COLORS: Record<string, string> = {
	assigned: "rgba(59,130,246,0.15)",
	mentioned: "rgba(168,85,247,0.15)",
	status_changed: "rgba(34,197,94,0.15)",
	comment_added: "rgba(59,130,246,0.15)",
	message_received: "rgba(59,130,246,0.15)",
	call_started: "rgba(34,197,94,0.15)",
	run_completed: "rgba(168,85,247,0.15)",
	run_failed: "rgba(239,68,68,0.15)",
	deadline_approaching: "rgba(249,115,22,0.15)",
	blocker_added: "rgba(239,68,68,0.15)",
};

const TYPE_ICON_TEXT_COLORS: Record<string, string> = {
	assigned: "text-blue-400",
	mentioned: "text-purple-400",
	status_changed: "text-emerald-400",
	comment_added: "text-blue-400",
	message_received: "text-blue-400",
	call_started: "text-emerald-400",
	run_completed: "text-purple-400",
	run_failed: "text-red-400",
	deadline_approaching: "text-orange-400",
	blocker_added: "text-red-400",
};

/** Русские названия типов уведомлений. */
const TYPE_LABELS: Record<string, string> = {
	assigned: "Назначение",
	mentioned: "Упоминание",
	status_changed: "Смена статуса",
	comment_added: "Комментарий",
	message_received: "Сообщение",
	call_started: "Звонок",
	run_completed: "Прогон завершён",
	run_failed: "Ошибки в прогоне",
	deadline_approaching: "Дедлайн",
	blocker_added: "Блокировщик",
};

/**
 * NotificationPanel — содержимое панели уведомлений.
 * Показывает список уведомлений с возможностью пометить прочитанным.
 */
export function NotificationPanel() {
	const notifications = useNotificationStore((s) => s.notifications);
	const loading = useNotificationStore((s) => s.loading);
	const hasMore = useNotificationStore((s) => s.hasMore);
	const markRead = useNotificationStore((s) => s.markRead);
	const markAllRead = useNotificationStore((s) => s.markAllRead);
	const clearAll = useNotificationStore((s) => s.clearAll);
	const deleteNotification = useNotificationStore((s) => s.deleteNotification);
	const fetchMore = useNotificationStore((s) => s.fetchMore);

	const scrollRef = useRef<HTMLDivElement>(null);

	/** Обработчик прокрутки для подгрузки следующей страницы. */
	const handleScroll = useCallback(
		(e: React.UIEvent<HTMLDivElement>) => {
			const target = e.currentTarget;
			if (
				target.scrollHeight - target.scrollTop - target.clientHeight < 100 &&
				hasMore &&
				!loading
			) {
				void fetchMore();
			}
		},
		[hasMore, loading, fetchMore],
	);

	const navigate = useNavigate();

	/**
	 * Определяет маршрут навигации по типу и связанной сущности уведомления.
	 * Поддерживает переход к задачам, каналам, тест-прогонам и звонкам.
	 */
	const resolveNotificationRoute = useCallback(
		(notif: NotificationDTO): string | null => {
			const { entity_type, entity_id, project_id } = notif;
			if (!entity_id) return null;

			switch (entity_type) {
				case "issue":
					// entity_id содержит issueKey (например "HUB-1")
					return project_id
						? `/projects/${project_id}/issues/${entity_id}`
						: null;
				case "channel":
					return `/messenger/${entity_id}`;
				case "test_run":
					return project_id
						? `/projects/${project_id}/tms/runs/${entity_id}`
						: null;
				case "call":
					// Для звонков переходим к каналу, в котором начался звонок
					return entity_id ? `/messenger/${entity_id}` : null;
				default:
					return null;
			}
		},
		[],
	);

	/** Обработчик клика по уведомлению — помечает прочитанным и переходит к сущности. */
	const handleItemClick = useCallback(
		(notif: NotificationDTO) => {
			if (!notif.is_read) {
				void markRead(notif.id);
			}

			const route = resolveNotificationRoute(notif);
			if (route) {
				try {
					void navigate({ to: route });
				} catch {
					toast.error(i18n.t("common:validation.objectNotFound"));
				}
			} else if (notif.entity_type) {
				toast.error(i18n.t("common:validation.objectNotFound"));
			}
		},
		[markRead, navigate, resolveNotificationRoute],
	);

	/** Обработчик удаления уведомления. */
	const handleDelete = useCallback(
		(e: React.MouseEvent, notifId: string) => {
			e.stopPropagation();
			void deleteNotification(notifId);
		},
		[deleteNotification],
	);

	return (
		<div className="flex flex-col animate-dropdown-in">
			{/* Заголовок */}
			<div className="flex items-center justify-between px-4 py-3">
				<h3 className="text-sm font-semibold">Уведомления</h3>
				<div className="flex items-center gap-1">
					<Button
						variant="ghost"
						size="sm"
						className="h-auto px-2 py-1 text-xs text-t2"
						onClick={() => void markAllRead()}
					>
						Прочитать все
					</Button>
					<Button
						variant="ghost"
						size="sm"
						className="h-auto px-2 py-1 text-xs text-destructive/70 hover:text-destructive"
						onClick={() => void clearAll()}
					>
						Удалить все
					</Button>
				</div>
			</div>
			<Separator />

			{/* Список уведомлений */}
			<ScrollArea
				className="h-96"
				ref={scrollRef}
				onScrollCapture={handleScroll}
			>
				{notifications.length === 0 && !loading ? (
					<div className="flex h-32 items-center justify-center text-sm text-t2">
						Нет уведомлений
					</div>
				) : (
					<div className="flex flex-col">
						{notifications.map((notif) => (
							<NotificationItem
								key={notif.id}
								notification={notif}
								onClick={() => handleItemClick(notif)}
								onDelete={(e) => handleDelete(e, notif.id)}
							/>
						))}
						{loading && (
							<div className="flex items-center justify-center py-4 text-xs text-t2">
								Загрузка...
							</div>
						)}
					</div>
				)}
			</ScrollArea>
		</div>
	);
}

/** Один элемент уведомления. */
const NotificationItem = memo(function NotificationItem({
	notification,
	onClick,
	onDelete,
}: {
	notification: NotificationDTO;
	onClick: () => void;
	onDelete: (e: React.MouseEvent) => void;
}) {
	const Icon = TYPE_ICONS[notification.type] ?? AlertTriangle;
	const typeLabel = TYPE_LABELS[notification.type] ?? notification.type;

	const iconBg =
		TYPE_ICON_COLORS[notification.type] ?? "rgba(148,163,184,0.15)";
	const iconTextColor = TYPE_ICON_TEXT_COLORS[notification.type] ?? "text-t2";

	return (
		<button
			type="button"
			className={`group flex w-full gap-3 px-4 py-3 text-left transition-colors hover:bg-raised/50 ${
				!notification.is_read
					? "border-l-2 border-blue-500 bg-blue-500/[0.04]"
					: ""
			}`}
			onClick={onClick}
		>
			{/* Иконка */}
			<div
				className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
				style={{ backgroundColor: iconBg }}
			>
				<Icon className={`h-4 w-4 ${iconTextColor}`} />
			</div>

			{/* Содержимое */}
			<div className="min-w-0 flex-1">
				<div className="flex items-start justify-between gap-2">
					<p className="text-sm font-medium leading-tight">
						<span className="mr-1 text-xs text-t2">{typeLabel}</span>
						{notification.title}
					</p>
					<div className="flex items-center gap-1 flex-shrink-0">
						{/* Кнопка удаления — видна при hover */}
						<button
							type="button"
							className="mt-0.5 hidden group-hover:flex h-5 w-5 items-center justify-center rounded-md hover:bg-destructive/10 text-t2 hover:text-destructive transition-colors"
							onClick={onDelete}
							title="Удалить"
							aria-label="Удалить уведомление"
						>
							<X className="h-3 w-3" />
						</button>
						{/* Индикатор непрочитанного */}
						{!notification.is_read && (
							<span className="mt-1 h-2 w-2 flex-shrink-0 rounded-full bg-blue-500" />
						)}
					</div>
				</div>
				{notification.body && (
					<p className="mt-0.5 line-clamp-2 text-xs text-t2">
						{notification.body}
					</p>
				)}
				<p className="mt-1 text-xs text-t2">
					{formatRelativeTime(notification.created_at)}
				</p>
			</div>
		</button>
	);
});
