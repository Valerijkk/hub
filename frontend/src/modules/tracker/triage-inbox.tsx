import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { createSavedFilter, listSavedFilters } from "@/api/search";
import { listIssues } from "@/api/tracker";
import { QueryStateView } from "@/components/query-state-view";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

const TRIAGE_FILTER_NAME = "Triage inbox";

const PRIORITY_COLORS: Record<string, string> = {
	critical: "bg-red-500 text-white",
	high: "bg-orange-500 text-white",
	medium: "bg-yellow-500 text-black",
	low: "bg-blue-500 text-white",
	trivial: "bg-gray-400 text-white",
};

interface TriageInboxProps {
	projectId: string;
}

/** TRIAGE-01: очередь неназначенных задач с preset сохранённого фильтра. */
export function TriageInbox({ projectId }: TriageInboxProps) {
	const queryClient = useQueryClient();

	const presetQuery = useQuery({
		queryKey: ["saved-filters-triage", projectId],
		queryFn: () => listSavedFilters(projectId),
	});

	const createPreset = useMutation({
		mutationFn: () =>
			createSavedFilter({
				name: TRIAGE_FILTER_NAME,
				project_id: projectId,
				filter_config: { assignee: "unassigned" },
			}),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["saved-filters-triage", projectId],
			});
		},
	});

	const presetBootstrapDone = useRef(false);

	useEffect(() => {
		if (
			!presetQuery.data ||
			createPreset.isPending ||
			presetBootstrapDone.current
		)
			return;
		const hasPreset = presetQuery.data.filters.some(
			(f) => f.name === TRIAGE_FILTER_NAME,
		);
		if (!hasPreset) {
			presetBootstrapDone.current = true;
			createPreset.mutate();
		}
	}, [presetQuery.data, createPreset.isPending, createPreset.mutate]);

	const issuesQuery = useQuery({
		queryKey: ["triage-inbox", projectId],
		queryFn: () =>
			listIssues(projectId, {
				assignee: "unassigned",
				limit: 100,
				order_by: "updated_at",
			}),
	});

	return (
		<div className="mx-auto max-w-5xl space-y-6 px-4 py-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Triage inbox</h1>
				<p className="text-sm text-t2 mt-1">
					Задачи без исполнителя — назначьте или переведите в работу
				</p>
			</div>

			<QueryStateView
				query={issuesQuery}
				loadingFallback={<p className="text-sm text-t3">Загрузка…</p>}
				errorFallback={() => (
					<p className="text-sm text-red-500">Не удалось загрузить очередь</p>
				)}
				isEmpty={(data) => (data.items?.length ?? 0) === 0}
				emptyFallback={
					<p className="text-sm text-t3">
						Нет неназначенных задач — inbox пуст
					</p>
				}
			>
				{(data) => (
					<div className="overflow-hidden rounded-lg border border-b1">
						<table className="w-full border-collapse text-sm">
							<thead>
								<tr className="border-b border-b1 bg-[var(--raised)] text-left">
									<th className="px-4 py-3 font-medium text-t2">Ключ</th>
									<th className="px-4 py-3 font-medium text-t2">Название</th>
									<th className="px-4 py-3 font-medium text-t2">Приоритет</th>
									<th className="px-4 py-3 font-medium text-t2">Статус</th>
								</tr>
							</thead>
							<tbody>
								{(data.items ?? []).map((issue) => (
									<tr
										key={issue.id}
										className="border-b border-b1 hover:bg-[var(--hover)]"
									>
										<td className="px-4 py-3">
											<Link
												to="/projects/$projectId/issues/$issueKey"
												params={{
													projectId,
													issueKey: issue.issue_key,
												}}
												className="font-mono text-primary hover:underline"
											>
												{issue.issue_key}
											</Link>
										</td>
										<td className="px-4 py-3 text-t1">{issue.title}</td>
										<td className="px-4 py-3">
											<Badge
												className={cn(
													"text-xs",
													PRIORITY_COLORS[issue.priority] ??
														"bg-gray-400 text-white",
												)}
											>
												{issue.priority}
											</Badge>
										</td>
										<td className="px-4 py-3 text-t2">
											{issue.status?.name ?? "—"}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</QueryStateView>
		</div>
	);
}
