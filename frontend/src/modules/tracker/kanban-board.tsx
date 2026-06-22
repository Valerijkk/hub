import { MoreHorizontal } from "lucide-react";
import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	type BoardColumnResponse,
	type BoardDataResponse,
	getBoardData,
	type IssueBriefResponse,
	moveIssue,
} from "@/api/tracker";
import { ContextHelpTip } from "@/components/help/context-help-tip";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";
import { sanitizeHexColor } from "@/lib/sanitize-color";
import { useBoardStore } from "@/stores/board-store";

/** Режим swimlanes на Kanban-доске. */
type SwimlaneMode = "none" | "assignee" | "priority" | "type";

/** Группирует задачи по полю для swimlanes. Pure — defined at module level. */
function getSwimlanes(
	issues: IssueBriefResponse[],
	mode: SwimlaneMode,
	unassignedLabel: string,
): Map<string, IssueBriefResponse[]> {
	if (mode === "none") {
		return new Map([["all", issues]]);
	}

	const groups = new Map<string, IssueBriefResponse[]>();
	for (const issue of issues) {
		let key: string;
		switch (mode) {
			case "assignee":
				key = issue.assignee?.name ?? unassignedLabel;
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
		const lane = groups.get(key);
		if (lane) {
			lane.push(issue);
		} else {
			groups.set(key, [issue]);
		}
	}
	return groups;
}

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
	const { t } = useTranslation("tracker");
	const [data, setData] = useState<BoardDataResponse | null>(null);
	const [swimlaneMode, setSwimlaneMode] = useState<SwimlaneMode>("none");
	const [loading, setLoading] = useState(true);
	const { setDragging } = useBoardStore();

	const loadBoard = useCallback(async () => {
		setLoading(true);
		try {
			const board = await getBoardData(projectId, boardId);
			setData(board);
		} catch (err: unknown) {
			toastApiError(err, t("kanban.loadError"));
			setData(null);
		}
		setLoading(false);
	}, [projectId, boardId, t]);

	useEffect(() => {
		void loadBoard();
	}, [loadBoard]);

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
			void loadBoard();
		} catch (err: unknown) {
			toastApiError(err, t("kanban.moveError"));
		}
		setDragging(false);
	};

	/** Получить все уникальные ключи swimlanes из всех колонок. */
	const unassignedLabel = t("kanban.unassigned");

	const swimlaneKeys = useMemo(() => {
		if (!data || swimlaneMode === "none") return ["all"];
		const keys = new Set<string>();
		for (const col of data.columns) {
			const lanes = getSwimlanes(
				col.issues ?? [],
				swimlaneMode,
				unassignedLabel,
			);
			for (const k of lanes.keys()) {
				keys.add(k);
			}
		}
		return Array.from(keys).sort();
	}, [data, swimlaneMode, unassignedLabel]);

	if (loading) {
		return (
			<div className="flex items-center justify-center p-8">
				<p className="text-t2">{t("kanban.loading")}</p>
			</div>
		);
	}

	if (!data) return null;

	const allStatuses = data.columns.map((col) => col.status);

	return (
		<div className="space-y-4">
			{/* Переключатель swimlanes */}
			<div className="flex items-center gap-3">
				<span className="text-sm font-medium">{t("kanban.swimlanes")}:</span>
				<Select
					value={swimlaneMode}
					onValueChange={(v) => setSwimlaneMode(v as SwimlaneMode)}
				>
					<SelectTrigger className="rounded-xl border bg-base px-3 py-1.5 text-sm w-auto min-w-[160px]">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="none">{t("kanban.swimlaneNone")}</SelectItem>
						<SelectItem value="assignee">
							{t("kanban.swimlaneAssignee")}
						</SelectItem>
						<SelectItem value="priority">
							{t("kanban.swimlanePriority")}
						</SelectItem>
						<SelectItem value="type">{t("kanban.swimlaneType")}</SelectItem>
					</SelectContent>
				</Select>
			</div>

			{/* Kanban-доска */}
			<div className="overflow-x-auto">
				{swimlaneKeys.map((laneKey) => (
					<div key={laneKey}>
						{swimlaneMode !== "none" && (
							<div className="mb-2 mt-4 border-b pb-1">
								<h3 className="text-sm font-semibold text-t2">{laneKey}</h3>
							</div>
						)}
						<div className="flex gap-4">
							{data.columns.map((col) => {
								const colIssues = col.issues ?? [];
								const lanes = getSwimlanes(
									colIssues,
									swimlaneMode,
									unassignedLabel,
								);
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
										allStatuses={allStatuses}
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
	allStatuses: BoardColumnResponse["status"][];
	onDrop: (issueKey: string, statusId: string, rank: string) => void;
}

/** Колонка Kanban-доски с заголовком, WIP-лимитом и списком карточек задач. */
const KanbanColumn = memo(function KanbanColumn({
	column,
	issues,
	wipLimit,
	totalCount,
	isOverLimit,
	allStatuses,
	onDrop,
}: KanbanColumnProps) {
	const { t } = useTranslation("tracker");
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
		<section
			aria-label={column.status.name}
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
						style={{ backgroundColor: sanitizeHexColor(column.status.color) }}
					/>
					<h3 className="text-sm font-semibold">{column.status.name}</h3>
				</div>
				<span
					className={`inline-flex items-center gap-1 text-xs font-medium ${
						isOverLimit ? "text-red-600" : "text-t2"
					}`}
				>
					{wipLimit > 0 ? `${totalCount}/${wipLimit}` : totalCount.toString()}
					{wipLimit > 0 && <ContextHelpTip tipKey="wip" className="h-3 w-3" />}
				</span>
			</div>

			<div className="space-y-2">
				{issues.map((issue) => (
					<KanbanCard
						key={issue.id}
						issue={issue}
						currentStatusId={column.status.id}
						allStatuses={allStatuses}
						onMove={(issueKey, statusId) => onDrop(issueKey, statusId, "")}
						moveLabel={t("kanban.moveIssue", { key: issue.issue_key })}
					/>
				))}
			</div>
		</section>
	);
});

// ---------------------------------------------------------------------------
// Компонент карточки задачи
// ---------------------------------------------------------------------------

/** Пропсы карточки задачи на Kanban-доске. */
interface KanbanCardProps {
	issue: IssueBriefResponse;
	currentStatusId: string;
	allStatuses: BoardColumnResponse["status"][];
	onMove: (issueKey: string, statusId: string) => void;
	moveLabel: string;
}

/** Карточка задачи: drag-and-drop + меню «Переместить в…» для touch/keyboard — UX-P0-04. */
const KanbanCard = memo(function KanbanCard({
	issue,
	currentStatusId,
	allStatuses,
	onMove,
	moveLabel,
}: KanbanCardProps) {
	const handleDragStart = (e: React.DragEvent) => {
		e.dataTransfer.setData("text/plain", issue.issue_key);
	};

	const moveTargets = allStatuses.filter((s) => s.id !== currentStatusId);

	return (
		<Card
			draggable
			onDragStart={handleDragStart}
			className="cursor-grab active:cursor-grabbing"
		>
			<CardHeader className="p-3">
				<div className="flex items-start justify-between gap-2">
					<div className="min-w-0 flex-1 space-y-1">
						<div className="flex items-center gap-2">
							<span className="text-xs font-mono text-t2">
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
					</div>
					{moveTargets.length > 0 ? (
						<DropdownMenu>
							<DropdownMenuTrigger asChild>
								<button
									type="button"
									className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-t3 hover:bg-[var(--hover)] hover:text-t1"
									aria-label={moveLabel}
								>
									<MoreHorizontal className="h-4 w-4" />
								</button>
							</DropdownMenuTrigger>
							<DropdownMenuContent align="end" className="min-w-[180px]">
								{moveTargets.map((status) => (
									<DropdownMenuItem
										key={status.id}
										onClick={() => onMove(issue.issue_key, status.id)}
									>
										<span
											className="mr-2 inline-block h-2 w-2 rounded-full"
											style={{
												backgroundColor: sanitizeHexColor(status.color),
											}}
											aria-hidden
										/>
										{status.name}
									</DropdownMenuItem>
								))}
							</DropdownMenuContent>
						</DropdownMenu>
					) : null}
				</div>
			</CardHeader>
			{issue.assignee && (
				<CardContent className="px-3 pb-3 pt-0">
					<span className="text-xs text-t2">{issue.assignee.name}</span>
				</CardContent>
			)}
		</Card>
	);
});
