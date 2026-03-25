import { memo, useEffect, useMemo, useState } from "react";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	type BoardColumnResponse,
	type BoardDataResponse,
	getBoardData,
	type IssueBriefResponse,
	moveIssue,
} from "@/api/tracker";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useBoardStore } from "@/stores/board-store";

/** Режим swimlanes на Kanban-доске. */
type SwimlaneMode = "none" | "assignee" | "priority" | "type";

/** Пропсы компонента KanbanBoard. */
interface KanbanBoardProps {
	/** ID проекта. */
	projectId: string;
	/** ID доски. */
	boardId: string;
}

/**
 * KanbanBoard отображает Kanban-доску с колонками статусов.
 * Поддерживает WIP-лимиты (KANB-02) и swimlanes (KANB-03).
 */
export function KanbanBoard({ projectId, boardId }: KanbanBoardProps) {
	const [data, setData] = useState<BoardDataResponse | null>(null);
	const [swimlaneMode, setSwimlaneMode] = useState<SwimlaneMode>("none");
	const [loading, setLoading] = useState(true);
	const { setDragging } = useBoardStore();

	useEffect(() => {
		loadBoard();
	}, [projectId, boardId]);

	const loadBoard = async () => {
		setLoading(true);
		try {
			const board = await getBoardData(projectId, boardId);
			setData(board);
		} catch {
			// ошибка загрузки доски
		}
		setLoading(false);
	};

	const handleDrop = async (
		issueKey: string,
		targetStatusId: string,
		rank: string,
	) => {
		try {
			await moveIssue(projectId, issueKey, {
				status_id: targetStatusId,
				rank,
			});
			loadBoard();
		} catch (err: unknown) {
			// Проверяем WIP-лимит ошибку (409)
			const apiErr = err as { code?: number; message?: string } | undefined;
			if (apiErr?.code === 409) {
				alert(apiErr.message ?? "WIP-лимит достигнут");
			}
		}
		setDragging(false);
	};

	/** Группирует задачи по полю для swimlanes. */
	const getSwimlanes = (
		issues: IssueBriefResponse[],
		mode: SwimlaneMode,
	): Map<string, IssueBriefResponse[]> => {
		if (mode === "none") {
			return new Map([["all", issues]]);
		}

		const groups = new Map<string, IssueBriefResponse[]>();
		for (const issue of issues) {
			let key: string;
			switch (mode) {
				case "assignee":
					key = issue.assignee?.name ?? "Без исполнителя";
					break;
				case "priority":
					key = issue.priority;
					break;
				case "type":
					key = issue.issue_type;
					break;
				default:
					key = "all";
			}
			if (!groups.has(key)) {
				groups.set(key, []);
			}
			groups.get(key)!.push(issue);
		}
		return groups;
	};

	/** Получить все уникальные ключи swimlanes из всех колонок. */
	const swimlaneKeys = useMemo(() => {
		if (!data || swimlaneMode === "none") return ["all"];
		const keys = new Set<string>();
		for (const col of data.columns) {
			const lanes = getSwimlanes(col.issues ?? [], swimlaneMode);
			for (const k of lanes.keys()) {
				keys.add(k);
			}
		}
		return Array.from(keys).sort();
	}, [data, swimlaneMode]);

	if (loading) {
		return (
			<div className="flex items-center justify-center p-8">
				<p className="text-muted-foreground">Загрузка доски...</p>
			</div>
		);
	}

	if (!data) return null;

	return (
		<div className="space-y-4">
			{/* Переключатель swimlanes */}
			<div className="flex items-center gap-3">
				<label className="text-sm font-medium">Swimlanes:</label>
				<Select
					value={swimlaneMode}
					onValueChange={(v) => setSwimlaneMode(v as SwimlaneMode)}
				>
					<SelectTrigger className="rounded-xl border bg-background px-3 py-1.5 text-sm w-auto min-w-[160px]">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="none">Без группировки</SelectItem>
						<SelectItem value="assignee">По исполнителю</SelectItem>
						<SelectItem value="priority">По приоритету</SelectItem>
						<SelectItem value="type">По типу</SelectItem>
					</SelectContent>
				</Select>
			</div>

			{/* Kanban-доска */}
			<div className="overflow-x-auto">
				{swimlaneKeys.map((laneKey) => (
					<div key={laneKey}>
						{swimlaneMode !== "none" && (
							<div className="mb-2 mt-4 border-b pb-1">
								<h3 className="text-sm font-semibold text-muted-foreground">
									{laneKey}
								</h3>
							</div>
						)}
						<div className="flex gap-4">
							{data.columns.map((col) => {
								const colIssues = col.issues ?? [];
								const lanes = getSwimlanes(colIssues, swimlaneMode);
								const laneIssues = lanes.get(laneKey) ?? [];
								const wipLimit =
									(col.status as { wip_limit?: number }).wip_limit ?? 0;
								const isOverLimit =
									wipLimit > 0 && colIssues.length >= wipLimit;

								return (
									<KanbanColumn
										key={col.status.id}
										column={col}
										issues={laneIssues}
										wipLimit={wipLimit}
										totalCount={colIssues.length}
										isOverLimit={isOverLimit}
										onDrop={handleDrop}
									/>
								);
							})}
						</div>
					</div>
				))}
			</div>
		</div>
	);
}

// ---------------------------------------------------------------------------
// Компонент колонки
// ---------------------------------------------------------------------------

/** Пропсы колонки Kanban-доски. */
interface KanbanColumnProps {
	column: BoardColumnResponse;
	issues: IssueBriefResponse[];
	wipLimit: number;
	totalCount: number;
	isOverLimit: boolean;
	onDrop: (issueKey: string, statusId: string, rank: string) => void;
}

/** Колонка Kanban-доски с заголовком, WIP-лимитом и списком карточек задач. */
const KanbanColumn = memo(function KanbanColumn({
	column,
	issues,
	wipLimit,
	totalCount,
	isOverLimit,
	onDrop,
}: KanbanColumnProps) {
	const handleDragOver = (e: React.DragEvent) => {
		if (isOverLimit) return;
		e.preventDefault();
	};

	const handleDrop = (e: React.DragEvent) => {
		if (isOverLimit) return;
		e.preventDefault();
		const issueKey = e.dataTransfer.getData("text/plain");
		if (issueKey) {
			onDrop(issueKey, column.status.id, "");
		}
	};

	return (
		<div
			className={`min-w-[280px] flex-1 rounded-lg border p-3 ${
				isOverLimit ? "border-red-300 bg-red-50" : "border-gray-200 bg-gray-50"
			}`}
			onDragOver={handleDragOver}
			onDrop={handleDrop}
		>
			<div className="mb-3 flex items-center justify-between">
				<div className="flex items-center gap-2">
					<div
						className="h-3 w-3 rounded-full"
						style={{ backgroundColor: column.status.color }}
					/>
					<h3 className="text-sm font-semibold">{column.status.name}</h3>
				</div>
				<span
					className={`text-xs font-medium ${
						isOverLimit ? "text-red-600" : "text-muted-foreground"
					}`}
				>
					{wipLimit > 0 ? `${totalCount}/${wipLimit}` : totalCount.toString()}
				</span>
			</div>

			<div className="space-y-2">
				{issues.map((issue) => (
					<KanbanCard key={issue.id} issue={issue} />
				))}
			</div>
		</div>
	);
});

// ---------------------------------------------------------------------------
// Компонент карточки задачи
// ---------------------------------------------------------------------------

/** Пропсы карточки задачи на Kanban-доске. */
interface KanbanCardProps {
	issue: IssueBriefResponse;
}

/** Перетаскиваемая карточка задачи на Kanban-доске с ключом, приоритетом и исполнителем. */
const KanbanCard = memo(function KanbanCard({ issue }: KanbanCardProps) {
	const handleDragStart = (e: React.DragEvent) => {
		e.dataTransfer.setData("text/plain", issue.issue_key);
	};

	return (
		<Card
			draggable
			onDragStart={handleDragStart}
			className="cursor-grab active:cursor-grabbing"
		>
			<CardHeader className="p-3">
				<div className="flex items-center gap-2">
					<span className="text-xs font-mono text-muted-foreground">
						{issue.issue_key}
					</span>
					<span
						className={`rounded px-1.5 py-0.5 text-xs ${
							issue.priority === "critical"
								? "bg-red-100 text-red-800"
								: issue.priority === "high"
									? "bg-orange-100 text-orange-800"
									: "bg-gray-100 text-gray-600"
						}`}
					>
						{issue.priority}
					</span>
				</div>
				<CardTitle className="text-sm font-normal">{issue.title}</CardTitle>
			</CardHeader>
			{issue.assignee && (
				<CardContent className="px-3 pb-3 pt-0">
					<span className="text-xs text-muted-foreground">
						{issue.assignee.name}
					</span>
				</CardContent>
			)}
		</Card>
	);
});
