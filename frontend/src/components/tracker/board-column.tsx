import { dropTargetForElements } from "@atlaskit/pragmatic-drag-and-drop/element/adapter";
import { Link } from "@tanstack/react-router";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useRef, useState } from "react";
import type { IssueBriefResponse, WorkflowStatusResponse } from "@/api/tracker";
import { cn } from "@/lib/utils";
import { IssueCard } from "./issue-card";

/** Пропсы компонента BoardColumn. */
interface BoardColumnProps {
	/** Статус workflow для данной колонки. */
	status: WorkflowStatusResponse;
	/** Задачи в этой колонке. */
	issues: IssueBriefResponse[];
	/** ID проекта. */
	projectId: string;
	/** Колбэк перемещения задачи в эту колонку. */
	onDrop: (issueKey: string, statusId: string, index: number) => void;
}

/**
 * Колонка Kanban-доски.
 * Является зоной сброса (drop target) для перетаскиваемых карточек.
 * Отображает заголовок со статусом и список карточек задач с анимацией.
 * Borderless design per mockup — cards carry their own borders.
 */
export function BoardColumn({
	status,
	issues,
	projectId,
	onDrop,
}: BoardColumnProps) {
	const columnRef = useRef<HTMLDivElement>(null);
	const [isOver, setIsOver] = useState(false);

	useEffect(() => {
		const el = columnRef.current;
		if (!el) return;

		return dropTargetForElements({
			element: el,
			getData: () => ({ statusId: status.id }),
			onDragEnter: () => setIsOver(true),
			onDragLeave: () => setIsOver(false),
			onDrop: ({ source }) => {
				setIsOver(false);
				const issueKey = source.data.issueKey as string;
				const currentStatusId = source.data.currentStatusId as string;
				if (currentStatusId === status.id) return;
				onDrop(issueKey, status.id, issues.length);
			},
		});
	}, [status.id, issues.length, onDrop]);

	return (
		<div
			ref={columnRef}
			className={cn(
				"min-w-[290px] max-w-[290px] flex flex-col",
				"max-h-[calc(100vh-172px)]",
				isOver && "ring-2 ring-primary/20 rounded-xl bg-primary/5",
			)}
		>
			{/* Заголовок колонки — group hover for dot glow */}
			<div className="group flex items-center gap-[9px] px-1.5 py-2.5 mb-2.5">
				<span
					className="w-[9px] h-[9px] rounded-full transition-[box-shadow] group-hover:shadow-[0_0_12px_currentColor]"
					style={{
						backgroundColor: status.color || "var(--gray)",
						color: status.color || "var(--gray)",
					}}
				/>
				<span className="text-[13.5px] font-bold tracking-[-0.01em]">
					{status.name}
				</span>
				<span className="text-[10.5px] font-bold text-[var(--t4)] bg-[rgba(255,255,255,.035)] px-[9px] py-0.5 rounded-full">
					{issues.length}
				</span>
				<Link
					to="/projects/$projectId/issues/new"
					params={{ projectId }}
					className="ml-auto w-7 h-7 flex items-center justify-center rounded-[var(--r2)] text-[var(--t4)] hover:bg-[var(--hover)] hover:text-[var(--t1)] hover:scale-[1.12] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity]"
				>
					<svg
						aria-hidden="true"
						className="w-4 h-4"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						strokeWidth={2}
					>
						<line x1="12" y1="5" x2="12" y2="19" />
						<line x1="5" y1="12" x2="19" y2="12" />
					</svg>
				</Link>
			</div>

			{/* Список карточек — скроллируемый контейнер */}
			<div className="flex flex-col gap-2.5 overflow-y-auto pr-1 flex-1">
				<AnimatePresence mode="popLayout">
					{issues.map((issue) => (
						<motion.div
							key={issue.id}
							layout
							initial={{ opacity: 0, y: 8 }}
							animate={{ opacity: 1, y: 0 }}
							exit={{ opacity: 0, scale: 0.95 }}
							transition={{ duration: 0.15 }}
						>
							<IssueCard issue={issue} projectId={projectId} />
						</motion.div>
					))}
				</AnimatePresence>

				{/* Пустое состояние — подсказка для перетаскивания */}
				{issues.length === 0 && (
					<div className="flex min-h-[100px] items-center justify-center rounded-lg border-2 border-dashed border-border/40 mx-0.5">
						<p className="text-xs text-t2/50 select-none">
							Перетащите задачу сюда
						</p>
					</div>
				)}
			</div>
		</div>
	);
}
