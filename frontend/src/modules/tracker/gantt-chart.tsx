import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import { type IssueResponse, listIssueLinks, listIssues } from "@/api/tracker";
import { cn } from "@/lib/utils";

interface GanttChartProps {
	projectId: string;
}

interface GanttBar {
	issue: IssueResponse;
	start: number;
	end: number;
}

const DAY_MS = 86_400_000;

/** GANTT-01: шкала created_at → due_date с подсказками зависимостей. */
export function GanttChart({ projectId }: GanttChartProps) {
	const { data, isLoading } = useQuery({
		queryKey: ["gantt-issues", projectId],
		queryFn: () =>
			listIssues(projectId, { limit: 100, order_by: "updated_at" }),
	});

	const issues = useMemo(
		() => (data?.items ?? []).filter((i) => i.due_date),
		[data?.items],
	);

	const { bars, rangeStart, rangeEnd, totalDays } = useMemo(() => {
		if (issues.length === 0) {
			return {
				bars: [] as GanttBar[],
				rangeStart: Date.now(),
				rangeEnd: Date.now(),
				totalDays: 1,
			};
		}
		const parsed: GanttBar[] = issues
			.filter((issue): issue is typeof issue & { due_date: string } =>
				Boolean(issue.due_date),
			)
			.map((issue) => {
				const start = new Date(issue.created_at).getTime();
				const end = new Date(issue.due_date).getTime();
				return {
					issue,
					start: Math.min(start, end),
					end: Math.max(start, end),
				};
			});
		const rangeStart = Math.min(...parsed.map((b) => b.start));
		const rangeEnd = Math.max(...parsed.map((b) => b.end));
		const totalDays = Math.max(1, Math.ceil((rangeEnd - rangeStart) / DAY_MS));
		return { bars: parsed, rangeStart, rangeEnd, totalDays };
	}, [issues]);

	if (isLoading) {
		return <p className="text-sm text-t3">Загрузка Gantt…</p>;
	}

	if (bars.length === 0) {
		return (
			<p className="text-sm text-t3">
				Нет задач с due date — задайте срок в карточке задачи
			</p>
		);
	}

	return (
		<div className="space-y-4">
			<p className="text-sm text-t2">
				{totalDays} дн. · {bars.length} задач с дедлайном
			</p>
			<div className="overflow-x-auto rounded-lg border border-b1">
				<div className="min-w-[640px]">
					{bars.map((bar) => {
						const left =
							((bar.start - rangeStart) / (rangeEnd - rangeStart || 1)) * 100;
						const width = Math.max(
							2,
							((bar.end - bar.start) / (rangeEnd - rangeStart || 1)) * 100,
						);
						return (
							<GanttRow
								key={bar.issue.id}
								projectId={projectId}
								bar={bar}
								left={left}
								width={width}
							/>
						);
					})}
				</div>
			</div>
		</div>
	);
}

function GanttRow({
	projectId,
	bar,
	left,
	width,
}: {
	projectId: string;
	bar: GanttBar;
	left: number;
	width: number;
}) {
	const { data: links } = useQuery({
		queryKey: ["issue-links", bar.issue.issue_key],
		queryFn: () => listIssueLinks(projectId, bar.issue.issue_key),
	});

	const blockers =
		links?.filter((l) => l.link_type === "blocked_by").length ?? 0;

	return (
		<div className="flex items-center gap-3 border-b border-b1 px-4 py-2 last:border-0">
			<div className="w-36 shrink-0 truncate">
				<Link
					to="/projects/$projectId/issues/$issueKey"
					params={{ projectId, issueKey: bar.issue.issue_key }}
					className="font-mono text-xs text-primary hover:underline"
				>
					{bar.issue.issue_key}
				</Link>
				<p className="truncate text-xs text-t3">{bar.issue.title}</p>
			</div>
			<div className="relative h-6 flex-1 rounded bg-[var(--raised)]">
				<div
					className={cn(
						"absolute top-1 h-4 rounded-sm bg-primary/70",
						blockers > 0 && "ring-1 ring-amber-400",
					)}
					style={{ left: `${left}%`, width: `${width}%` }}
					title={
						blockers > 0 ? `${blockers} blocker link(s)` : bar.issue.status.name
					}
				/>
			</div>
		</div>
	);
}
