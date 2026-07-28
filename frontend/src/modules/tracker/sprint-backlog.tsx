import { Link } from "@tanstack/react-router";
import { ArrowRight, Inbox, ListTodo } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { type IssueResponse, listIssues } from "@/api/tracker";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { toastApiError } from "@/lib/format-api-error";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента SprintBacklog. */
interface SprintBacklogProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * SprintBacklog отображает задачи без спринта (бэклог)
 * и задачи в спринтах со статусом planning.
 * Поддерживает выбор задач и перемещение в спринт.
 *
 * A2W10-A3-002: fail load ≠ empty «Бэклог пуст».
 */
export function SprintBacklog({ projectId }: SprintBacklogProps) {
	const { sprints, fetchSprints, moveIssueToSprint } = useSprintStore();
	const [backlogIssues, setBacklogIssues] = useState<IssueResponse[]>([]);
	const [selectedIssues, setSelectedIssues] = useState<Set<string>>(new Set());
	const [loading, setLoading] = useState(false);
	const [loadError, setLoadError] = useState(false);

	const loadBacklog = useCallback(async () => {
		setLoading(true);
		try {
			const resp = await listIssues(projectId, { limit: 100 });
			setBacklogIssues(resp.items);
			setLoadError(false);
		} catch (err) {
			setLoadError(true);
			setBacklogIssues([]);
			toastApiError(err, "Не удалось загрузить бэклог");
		}
		setLoading(false);
	}, [projectId]);

	useEffect(() => {
		void fetchSprints(projectId);
		void loadBacklog();
	}, [projectId, fetchSprints, loadBacklog]);

	const toggleSelect = (issueId: string) => {
		const next = new Set(selectedIssues);
		if (next.has(issueId)) {
			next.delete(issueId);
		} else {
			next.add(issueId);
		}
		setSelectedIssues(next);
	};

	const handleMoveToSprint = async (sprintId: string) => {
		if (selectedIssues.size === 0) return;
		await moveIssueToSprint(projectId, sprintId, Array.from(selectedIssues));
		setSelectedIssues(new Set());
		void loadBacklog();
	};

	const planningSprints = sprints.filter((s) => s.status === "planning");

	return (
		<div className="space-y-6 animate-[eu_450ms_var(--spring)_both]">
			{/* Заголовок страницы */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<div className="flex h-10 w-10 items-center justify-center rounded-[var(--r3)] bg-gradient-to-br from-violet-500/20 to-blue-500/20 text-violet-500">
						<Inbox className="h-5 w-5" />
					</div>
					<div>
						<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">
							Бэклог
						</h1>
						<p className="mt-0.5 text-[13px] text-t2">
							{loadError
								? "Ошибка загрузки"
								: `${backlogIssues.length} задач без спринта`}
						</p>
					</div>
				</div>

				{/* Панель «Перенести в спринт» */}
				{selectedIssues.size > 0 && planningSprints.length > 0 && (
					<div className="flex items-center gap-3">
						<span className="rounded-full bg-blue-500/10 px-3 py-1 text-xs font-medium text-blue-400 border border-blue-500/20">
							Выбрано: {selectedIssues.size}
						</span>
						{planningSprints.map((s) => (
							<Button
								key={s.id}
								size="sm"
								className="btn-glow bg-gradient-to-r from-blue-600 to-blue-500 text-white shadow-md shadow-blue-500/20 hover:from-blue-500 hover:to-blue-400"
								onClick={() => void handleMoveToSprint(s.id)}
							>
								<ArrowRight className="mr-1.5 h-3.5 w-3.5" />В {s.name}
							</Button>
						))}
					</div>
				)}
			</div>

			{/* Задачи бэклога */}
			{loading ? (
				<div className="flex items-center justify-center p-12">
					<div className="flex flex-col items-center gap-3">
						<div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
						<p className="text-sm text-t2">Загрузка бэклога...</p>
					</div>
				</div>
			) : loadError ? (
				<QueryErrorAlert
					message="Не удалось загрузить бэклог"
					retry={() => void loadBacklog()}
				/>
			) : backlogIssues.length === 0 ? (
				<div className="flex flex-col items-center justify-center py-16 text-center">
					<div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500/20 to-blue-500/20">
						<ListTodo className="h-8 w-8 text-violet-400" />
					</div>
					<h3 className="text-lg font-semibold text-t1">Бэклог пуст</h3>
					<p className="mt-1 max-w-sm text-sm text-t2">
						Все задачи распределены по спринтам
					</p>
				</div>
			) : (
				<div className="rounded-[var(--r4)] border border-b1 bg-card/50 p-1">
					<div className="space-y-1">
						{backlogIssues.map((issue) => {
							const isSelected = selectedIssues.has(issue.id);
							return (
								<div
									key={issue.id}
									className={`group flex items-center gap-3 rounded-[var(--r3)] px-4 py-3 border-b border-b1 text-[13px] transition-colors duration-150 ${
										isSelected
											? "border-blue/20 bg-blue/5 ring-1 ring-blue/20"
											: "border-transparent hover:bg-[var(--hover)]"
									}`}
								>
									<input
										id={`backlog-issue-${issue.id}`}
										type="checkbox"
										checked={isSelected}
										onChange={() => toggleSelect(issue.id)}
										aria-label={`Выбрать ${issue.issue_key}`}
										className="h-4 w-4 shrink-0 cursor-pointer accent-blue-400"
									/>
									{/* Ключ задачи */}
									<span className="flex-shrink-0 rounded bg-white/[0.04] px-1.5 py-0.5 font-mono text-[11px] text-t2">
										{issue.issue_key}
									</span>

									{/* Название */}
									<Link
										to="/projects/$projectId/issues/$issueKey"
										params={{
											projectId,
											issueKey: issue.issue_key,
										}}
										className="min-w-0 flex-1 truncate text-sm text-t1 transition-colors hover:text-blue-400"
										onClick={(e) => e.stopPropagation()}
									>
										{issue.title}
									</Link>

									{/* Бейдж приоритета */}
									<span
										className={`flex-shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium ${
											issue.priority === "critical"
												? "bg-red-500/10 text-red-400 border border-red-500/20"
												: issue.priority === "high"
													? "bg-orange-500/10 text-orange-400 border border-orange-500/20"
													: issue.priority === "medium"
														? "bg-amber-500/10 text-amber-400 border border-amber-500/20"
														: "bg-zinc-500/10 text-zinc-400 border border-zinc-500/20"
										}`}
									>
										{issue.priority}
									</span>
								</div>
							);
						})}
					</div>
				</div>
			)}
		</div>
	);
}
