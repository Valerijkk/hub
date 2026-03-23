import { draggable } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Link } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import type { IssueBriefResponse } from "@/api/tracker";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

/** Конфигурация цветов и меток приоритетов задач. */
const PRIORITY_CONFIG: Record<
	string,
	{ dot: string; bg: string; text: string; label: string; borderColor: string }
> = {
	critical: {
		dot: "bg-red-500",
		bg: "bg-red-500/15",
		text: "text-red-700 dark:text-red-400",
		label: "Критический",
		borderColor: "#ef4444",
	},
	high: {
		dot: "bg-orange-500",
		bg: "bg-orange-500/15",
		text: "text-orange-700 dark:text-orange-400",
		label: "Высокий",
		borderColor: "#f97316",
	},
	medium: {
		dot: "bg-yellow-500",
		bg: "bg-yellow-500/15",
		text: "text-yellow-700 dark:text-yellow-400",
		label: "Средний",
		borderColor: "#eab308",
	},
	low: {
		dot: "bg-blue-500",
		bg: "bg-blue-500/15",
		text: "text-blue-700 dark:text-blue-400",
		label: "Низкий",
		borderColor: "#3b82f6",
	},
	trivial: {
		dot: "bg-gray-400",
		bg: "bg-gray-400/15",
		text: "text-gray-600 dark:text-gray-400",
		label: "Тривиальный",
		borderColor: "#9ca3af",
	},
};

/** Конфигурация типов задач — SVG-иконки с цветами для лучшей читаемости. */
const TYPE_CONFIG: Record<string, { icon: string; color: string }> = {
	bug: { icon: "\u{1F41B}", color: "text-red-500" },
	task: { icon: "\u{2705}", color: "text-blue-500" },
	story: { icon: "\u{1F4D6}", color: "text-green-500" },
	epic: { icon: "\u{26A1}", color: "text-purple-500" },
	sub_task: { icon: "\u{1F4CB}", color: "text-gray-500" },
	improvement: { icon: "\u{1F680}", color: "text-cyan-500" },
	tech_debt: { icon: "\u{1F527}", color: "text-amber-500" },
	spike: { icon: "\u{1F50D}", color: "text-indigo-500" },
	release: { icon: "\u{1F3F7}\uFE0F", color: "text-pink-500" },
};

/** Пропсы компонента IssueCard. */
interface IssueCardProps {
	/** Краткая информация о задаче. */
	issue: IssueBriefResponse;
	/** ID проекта. */
	projectId: string;
}

/**
 * Карточка задачи для Kanban-доски.
 * Поддерживает drag-and-drop перетаскивание между колонками.
 * Кликабельна — переход на страницу детальной информации о задаче.
 * Отображает тип, ключ, название (до 2 строк), приоритет, story points,
 * прогресс чеклиста и аватар исполнителя.
 *
 * Навигация реализована через useNavigate + отслеживание координат указателя,
 * чтобы избежать конфликта между pragmatic-drag-and-drop и обычным кликом.
 * Если указатель сместился более чем на 5 px — считается перетаскиванием.
 */
export function IssueCard({ issue, projectId }: IssueCardProps) {
	const cardRef = useRef<HTMLDivElement>(null);
	const [isDragging, setIsDragging] = useState(false);

	/** Координаты указателя при нажатии — для отличия клика от перетаскивания. */
	const pointerDownPos = useRef<{ x: number; y: number } | null>(null);
	/** Флаг: было ли реальное перетаскивание (не просто клик). */
	const wasDragged = useRef(false);

	useEffect(() => {
		const el = cardRef.current;
		if (!el) return;

		return draggable({
			element: el,
			getInitialData: () => ({
				issueKey: issue.issue_key,
				issueId: issue.id,
				currentStatusId: issue.status,
				rank: issue.rank ?? "",
			}),
			onDragStart: () => {
				setIsDragging(true);
				wasDragged.current = true;
			},
			onDrop: () => {
				setIsDragging(false);
				// Сбросим флаг с задержкой, чтобы Link.onClick мог его проверить.
				requestAnimationFrame(() => {
					wasDragged.current = false;
				});
			},
		});
	}, [issue.issue_key, issue.id, issue.status, issue.rank]);

	const priority = PRIORITY_CONFIG[issue.priority] ?? PRIORITY_CONFIG.medium!;
	const typeConfig = TYPE_CONFIG[issue.issue_type] ?? {
		icon: "\u{1F4CB}",
		color: "text-gray-500",
	};

	/**
	 * Запоминаем начальную позицию указателя при mousedown.
	 */
	const handlePointerDown = (e: React.PointerEvent) => {
		pointerDownPos.current = { x: e.clientX, y: e.clientY };
	};

	/**
	 * Предотвращаем переход по ссылке если произошёл drag.
	 * Используем onClick на Link-обёртке для перехвата.
	 */
	const handleLinkClick = (e: React.MouseEvent) => {
		// Если было реальное перетаскивание — отменяем навигацию.
		if (wasDragged.current || isDragging) {
			e.preventDefault();
			return;
		}
		// Если указатель сместился более чем на 5 px — тоже отменяем.
		const start = pointerDownPos.current;
		if (start) {
			const dx = Math.abs(e.clientX - start.x);
			const dy = Math.abs(e.clientY - start.y);
			if (dx > 5 || dy > 5) {
				e.preventDefault();
				return;
			}
		}
		// Иначе: обычный клик — Link выполнит навигацию.
	};

	/** Прогресс чеклиста (количество выполненных / всего). */
	const checklistTotal = issue.checklist?.length ?? 0;
	const checklistDone = issue.checklist?.filter((c) => c.checked).length ?? 0;
	const checklistPct = checklistTotal > 0 ? Math.round((checklistDone / checklistTotal) * 100) : 0;

	return (
		<Link
			to="/projects/$projectId/issues/$issueKey"
			params={{ projectId, issueKey: issue.issue_key }}
			onClick={handleLinkClick}
			draggable={false}
			className="block no-underline text-inherit"
		>
			<div
				ref={cardRef}
				role="button"
				tabIndex={0}
				className={cn(
					"cursor-grab active:cursor-grabbing select-none",
					isDragging && "opacity-30",
				)}
				onPointerDown={handlePointerDown}
			>
				<div
					className={cn(
						"rounded-lg border bg-card p-3 transition-all duration-200",
						"border-border/50 shadow-sm",
						"border-l-[3px]",
						"hover:shadow-md hover:border-primary/30 hover:-translate-y-[2px]",
						"focus-visible:ring-2 focus-visible:ring-primary/30",
					)}
					style={{ borderLeftColor: priority.borderColor }}
				>
					{/* Верхняя строка: тип + ключ */}
					<div className="flex items-center gap-1.5 mb-1.5">
						<span
							className={cn("text-sm leading-none", typeConfig.color)}
							title={issue.issue_type}
						>
							{typeConfig.icon}
						</span>
						<span className="font-mono text-[11px] font-semibold text-muted-foreground/80 tracking-wide">
							{issue.issue_key}
						</span>
					</div>

					{/* Название — до 2 строк */}
					<p className="text-[13px] font-medium leading-snug line-clamp-2 mb-2.5 text-foreground/90">
						{issue.title}
					</p>

					{/* Прогресс чеклиста */}
					{checklistTotal > 0 && (
						<div className="mb-2">
							<div className="flex items-center justify-between mb-0.5">
								<span className="text-[10px] text-muted-foreground">
									{checklistDone}/{checklistTotal}
								</span>
							</div>
							<div className="h-1 w-full rounded-full bg-muted">
								<div
									className={cn(
										"h-1 rounded-full transition-all",
										checklistPct === 100 ? "bg-green-500" : "bg-primary",
									)}
									style={{ width: `${checklistPct}%` }}
								/>
							</div>
						</div>
					)}

					{/* Нижняя строка: приоритет + story points + аватар исполнителя */}
					<div className="flex items-center justify-between">
						<div className="flex items-center gap-1.5">
							<Badge
								variant="secondary"
								className={cn(
									"text-[10px] font-semibold px-1.5 py-0 h-[18px] rounded-md border-0",
									priority.bg,
									priority.text,
								)}
							>
								{priority.label}
							</Badge>

							{issue.story_points != null && (
								<Badge
									variant="outline"
									className="text-[10px] font-semibold px-1.5 py-0 h-[18px] rounded-md"
								>
									{issue.story_points} SP
								</Badge>
							)}
						</div>

						{issue.assignee ? (
							<Avatar className="h-5 w-5 ring-1 ring-border/50">
								<AvatarImage src={issue.assignee.avatar_url} />
								<AvatarFallback className="text-[9px] font-medium bg-primary/10 text-primary">
									{issue.assignee.name.charAt(0).toUpperCase()}
								</AvatarFallback>
							</Avatar>
						) : (
							<div className="h-5 w-5 rounded-full border border-dashed border-border/40" />
						)}
					</div>
				</div>
			</div>
		</Link>
	);
}
