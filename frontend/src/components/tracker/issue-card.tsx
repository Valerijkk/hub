import { draggable } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Link } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import type { IssueBriefResponse } from "@/api/tracker";
import { cn } from "@/lib/utils";

/** Конфигурация цветов и меток приоритетов задач. */
const PRIORITY_CONFIG: Record<
	string,
	{ color: string; label: string }
> = {
	critical: { color: "#ef4444", label: "Критический" },
	high: { color: "#f97316", label: "Высокий" },
	medium: { color: "#eab308", label: "Средний" },
	low: { color: "#3b82f6", label: "Низкий" },
	trivial: { color: "#9ca3af", label: "Тривиальный" },
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

	/** Assignee color from avatar_url or fallback gradient */
	const assigneeBg = issue.assignee
		? "linear-gradient(140deg, var(--blueS), var(--purpleS))"
		: undefined;

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
					"select-none",
					isDragging && "opacity-30",
				)}
				onPointerDown={handlePointerDown}
			>
				<div
					className={cn(
						"kanban-card",
						issue.priority === "critical" && "kanban-card-critical",
						issue.priority === "high" && "kanban-card-high",
						issue.status === "done" && "kanban-card-done",
					)}
				>
					{/* Верхняя строка: тип + ключ */}
					<div className="flex items-center gap-[7px] mb-[7px]">
						<span
							className={cn("text-sm leading-none", typeConfig.color)}
							title={issue.issue_type}
						>
							{typeConfig.icon}
						</span>
						<span className="font-mono text-[11px] text-[var(--t4)] font-medium">
							{issue.issue_key}
						</span>
					</div>

					{/* Название — до 2 строк */}
					<p className="text-[13.5px] font-semibold leading-[1.45] tracking-[-0.005em] line-clamp-2 mb-3 transition-colors hover:text-white">
						{issue.title}
					</p>

					{/* Meta row: priority dot + SP + checklist */}
					<div className="flex items-center gap-2">
						<span
							className="w-2 h-2 rounded-[3px] transition-all hover:shadow-[0_0_10px_currentColor] hover:scale-125"
							style={{ backgroundColor: priority.color, color: priority.color }}
						/>

						{issue.story_points != null && (
							<span className="text-[10.5px] font-bold text-[var(--t4)] bg-[rgba(255,255,255,.035)] px-[7px] py-0.5 rounded-[var(--r1)]">
								{issue.story_points} SP
							</span>
						)}

						{checklistTotal > 0 && (
							<div className="flex items-center gap-[5px] text-[11px] text-[var(--t4)]">
								<div className="w-10 h-[3px] bg-[rgba(255,255,255,.04)] rounded-sm overflow-hidden">
									<div
										className="h-full bg-[var(--green)] rounded-sm shadow-[0_0_6px_var(--gG)]"
										style={{ width: `${checklistPct}%` }}
									/>
								</div>
								{checklistDone}/{checklistTotal}
							</div>
						)}
					</div>

					{/* Footer row: assignee avatar + labels */}
					<div className="flex items-center justify-between mt-3.5">
						{issue.assignee ? (
							<div
								className="w-[26px] h-[26px] rounded-full flex items-center justify-center text-[10px] font-bold text-white shadow-[0_2px_6px_rgba(0,0,0,.3)] hover:scale-[1.12] transition-transform"
								style={{ background: assigneeBg }}
							>
								{issue.assignee.name.charAt(0).toUpperCase()}
							</div>
						) : (
							<div className="w-[26px] h-[26px] rounded-full border border-dashed border-border/40 flex items-center justify-center text-[10px] text-[var(--t4)]">
								-
							</div>
						)}

						{/* Labels — placeholder for future IssueBriefResponse extension */}
					</div>
				</div>
			</div>
		</Link>
	);
}
