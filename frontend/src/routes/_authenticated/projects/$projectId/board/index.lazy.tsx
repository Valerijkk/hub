import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useCallback, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
	type BoardDataResponse,
	getBoardData,
	listBoards,
	moveIssue,
} from "@/api/tracker";
import { QueryStateView } from "@/components/query-state-view";
import { BoardColumn } from "@/components/tracker/board-column";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { computeInsertRank } from "@/lib/board-rank";
import { toastApiError } from "@/lib/format-api-error";
import { useBoardStore } from "@/stores/board-store";

/**
 * Страница Kanban-доски проекта.
 * Отображает колонки по статусам workflow с drag-and-drop перемещением.
 * Поддерживает фильтрацию, выбор доски и оптимистичные обновления.
 * Занимает всю доступную высоту с горизонтальной прокруткой колонок.
 */
function BoardPage() {
	const { t } = useTranslation("tracker");
	const { projectId } = Route.useParams();
	const queryClient = useQueryClient();
	const { activeBoardId, setActiveBoard, filters, setFilters, resetFilters } =
		useBoardStore();

	// Загрузка списка досок
	const { data: boards } = useQuery({
		queryKey: ["boards", projectId],
		queryFn: () => listBoards(projectId),
	});

	// A13-011: derive default board id from list — avoid waterfall
	// (listBoards → useEffect setActiveBoard → only then getBoardData).
	const effectiveBoardId = useMemo(() => {
		if (activeBoardId) return activeBoardId;
		if (!boards || boards.length === 0) return null;
		const defaultBoard = boards.find((b) => b.is_default) ?? boards[0];
		return defaultBoard?.id ?? null;
	}, [activeBoardId, boards]);

	// Persist derived default into store (UI select stays in sync).
	useEffect(() => {
		if (!activeBoardId && effectiveBoardId) {
			setActiveBoard(effectiveBoardId);
		}
	}, [activeBoardId, effectiveBoardId, setActiveBoard]);

	// Загрузка данных доски — starts as soon as list resolves (no effect wait).
	const boardQuery = useQuery({
		queryKey: ["board", projectId, effectiveBoardId],
		// biome-ignore lint/style/noNonNullAssertion: enabled flag guards effectiveBoardId
		queryFn: () => getBoardData(projectId, effectiveBoardId!),
		enabled: !!effectiveBoardId,
	});
	const { data: boardData } = boardQuery;

	// Мутация перемещения задачи с оптимистичным обновлением
	const moveCardMutation = useMutation({
		mutationFn: (data: { issueKey: string; statusId: string; rank: string }) =>
			moveIssue(projectId, data.issueKey, {
				status_id: data.statusId,
				rank: data.rank,
			}),
		onMutate: async (data) => {
			await queryClient.cancelQueries({
				queryKey: ["board", projectId, effectiveBoardId],
			});
			const previous = queryClient.getQueryData<BoardDataResponse>([
				"board",
				projectId,
				effectiveBoardId,
			]);
			queryClient.setQueryData<BoardDataResponse>(
				["board", projectId, effectiveBoardId],
				(old) => {
					if (!old) return old;
					return reorderBoard(old, data);
				},
			);
			return { previous };
		},
		onError: (err, _data, context) => {
			if (context?.previous) {
				queryClient.setQueryData(
					["board", projectId, effectiveBoardId],
					context.previous,
				);
			}
			toastApiError(err, t("kanban.moveError"));
		},
		onSettled: () => {
			void queryClient.invalidateQueries({
				queryKey: ["board", projectId, effectiveBoardId],
			});
		},
	});

	// Единственный путь moveIssue: BoardColumn dropTarget (W10-A4-001).
	// Rank от соседей целевой колонки (W10-A4-002) — не константа.
	const handleDrop = useCallback(
		(issueKey: string, statusId: string, index: number) => {
			const col = boardData?.columns.find((c) => c.status.id === statusId);
			const neighbors = (col?.issues ?? []).filter(
				(i) => i.issue_key !== issueKey,
			);
			const rank = computeInsertRank(neighbors, index);
			moveCardMutation.mutate({
				issueKey,
				statusId,
				rank,
			});
		},
		[moveCardMutation, boardData],
	);

	// Фильтрация задач в колонках
	const filteredColumns = useMemo(() => {
		if (!boardData?.columns) return [];
		return boardData.columns.map((col) => ({
			...col,
			issues: (col.issues ?? []).filter((issue) => {
				if (filters.search) {
					const q = filters.search.toLowerCase();
					if (
						!issue.title.toLowerCase().includes(q) &&
						!issue.issue_key.toLowerCase().includes(q)
					) {
						return false;
					}
				}
				if (
					filters.priority?.length &&
					!filters.priority.includes(issue.priority)
				) {
					return false;
				}
				if (filters.type && issue.issue_type !== filters.type) {
					return false;
				}
				if (filters.assignee && issue.assignee?.id !== filters.assignee) {
					return false;
				}
				return true;
			}),
		}));
	}, [boardData?.columns, filters]);

	const boardLoadingFallback = (
		<div className="board-page-root flex flex-col">
			{/* Скелетон панели фильтров */}
			<div className="flex items-center gap-3 px-2 pb-4">
				<div className="h-9 w-48 rounded-lg animate-shimmer" />
				<div className="h-9 w-24 rounded-lg animate-shimmer" />
				<div className="ml-auto h-9 w-32 rounded-lg animate-shimmer" />
			</div>
			{/* Скелетон колонок */}
			<div className="flex flex-1 gap-4 overflow-x-auto pb-4">
				{[1, 2, 3, 4, 5, 6].map((i) => (
					<div
						key={`board-col-sk-${i}`}
						className="flex w-[300px] min-w-[300px] flex-shrink-0 flex-col rounded-xl border border-border/40 bg-raised/30"
					>
						<div className="px-4 py-3">
							<div className="h-5 w-24 rounded animate-shimmer" />
						</div>
						<div className="space-y-2 px-2 pb-2">
							{[1, 2, 3].map((j) => (
								<div
									key={`board-card-sk-${i}-${j}`}
									className="h-28 rounded-lg animate-shimmer"
								/>
							))}
						</div>
					</div>
				))}
			</div>
		</div>
	);

	return (
		<div className="board-page-root flex flex-col">
			{/* Панель фильтров */}
			<div className="flex items-center gap-2.5 px-6 h-[50px] border-b border-[var(--b1)] shrink-0">
				{/* Выбор доски */}
				{boards && boards.length > 1 && (
					<Select
						value={activeBoardId ?? ""}
						onValueChange={(v) => setActiveBoard(v)}
					>
						<SelectTrigger
							aria-label="Выбор доски"
							className="rounded-xl border bg-card px-3 py-2 text-sm font-medium w-auto min-w-[140px]"
						>
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{boards.map((b) => (
								<SelectItem key={b.id} value={b.id}>
									{b.name}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				)}

				<Input
					placeholder="Поиск задач..."
					className="w-full sm:w-[220px] h-[34px] pl-9 pr-3 bg-[var(--input)] border border-transparent rounded-[var(--r2)] text-[13px] text-[var(--t1)] placeholder:text-[var(--t3)] focus:border-[rgba(91,156,245,.25)] focus:bg-[var(--card)] transition-colors outline-none"
					value={filters.search ?? ""}
					onChange={(e) => setFilters({ search: e.target.value || undefined })}
				/>

				<button
					type="button"
					className="kanban-fchip"
					onClick={() => setFilters({ myTasks: !filters.myTasks })}
				>
					Мои задачи
				</button>

				{(filters.search ||
					filters.priority?.length ||
					filters.type ||
					filters.assignee ||
					filters.myTasks) && (
					<button
						type="button"
						className="text-xs text-[var(--t4)] cursor-pointer ml-auto hover:text-[var(--red)] transition-colors bg-transparent border-none"
						onClick={resetFilters}
					>
						Сбросить
					</button>
				)}

				<div className="ml-auto">
					<Link
						to="/projects/$projectId/issues/new"
						params={{ projectId }}
						className="inline-flex items-center gap-[7px] px-[18px] py-2 rounded-[var(--r2)] text-[13px] font-semibold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_20px_var(--gB)] hover:shadow-[0_0_32px_rgba(91,156,245,.3)] hover:-translate-y-px active:scale-[.96] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] relative overflow-hidden"
					>
						<svg
							aria-hidden="true"
							className="h-3.5 w-3.5"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2.5}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M12 5v14m7-7H5"
							/>
						</svg>
						Задача
					</Link>
				</div>
			</div>

			{/* Колонки доски с горизонтальной прокруткой */}
			<QueryStateView
				query={boardQuery}
				loadingFallback={boardLoadingFallback}
				isEmpty={() => filteredColumns.length === 0}
				emptyFallback={
					<div className="flex flex-1 items-center justify-center">
						<div className="text-center max-w-sm animate-fade-in">
							<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10">
								<svg
									aria-hidden="true"
									className="h-8 w-8 text-primary/50"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={1.5}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2"
									/>
								</svg>
							</div>
							<h3 className="text-lg font-semibold mb-2">Нет задач на доске</h3>
							<p className="text-sm text-t2 mb-4">
								Создайте первую задачу, чтобы начать работу с доской
							</p>
							<Button asChild>
								<Link
									to="/projects/$projectId/issues/new"
									params={{ projectId }}
								>
									Создать задачу
								</Link>
							</Button>
						</div>
					</div>
				}
			>
				{() => (
					<div className="flex-1 overflow-x-auto overflow-y-hidden">
						<div className="flex gap-4 px-6 py-5 overflow-x-auto flex-1 items-start kanban-enter">
							{filteredColumns.map((col) => (
								<BoardColumn
									key={col.status.id}
									status={col.status}
									issues={col.issues}
									projectId={projectId}
									onDrop={handleDrop}
								/>
							))}
						</div>
					</div>
				)}
			</QueryStateView>
		</div>
	);
}

/**
 * Оптимистично перемещает задачу между колонками доски.
 * Используется для мгновенного отображения результата drag-and-drop.
 */
function reorderBoard(
	board: BoardDataResponse,
	data: { issueKey: string; statusId: string; rank: string },
): BoardDataResponse {
	const columns = board.columns.map((col) => ({
		...col,
		issues: [...(col.issues ?? [])],
	}));

	let movedIssue = null;
	for (const col of columns) {
		const idx = col.issues.findIndex((i) => i.issue_key === data.issueKey);
		if (idx !== -1) {
			movedIssue = col.issues.splice(idx, 1)[0];
			break;
		}
	}

	if (movedIssue) {
		const targetCol = columns.find((c) => c.status.id === data.statusId);
		if (targetCol) {
			targetCol.issues.push({
				...movedIssue,
				status: data.statusId,
				rank: data.rank,
			});
		}
	}

	return { ...board, columns };
}

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/board/",
)({
	component: BoardPage,
});
