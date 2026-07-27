import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useEffect, useMemo, useRef } from "react";
import {
	type IssueResponse,
	listIssues,
	listProjectIssueLinks,
} from "@/api/tracker";
import { QueryStateView } from "@/components/query-state-view";
import { cn } from "@/lib/utils";
import {
	flattenIssuePages,
	nextIssuesCursor,
} from "@/modules/tracker/issues-cursor";

interface GanttChartProps {
	projectId: string;
}

interface GanttBar {
	issue: IssueResponse;
	start: number;
	end: number;
}

const DAY_MS = 86_400_000;

/** Above this count, prefer row virtualization over full DOM map (A10-014; mirrors board A13-014). */
export const GANTT_VIRTUALIZE_THRESHOLD = 24;
const ESTIMATED_ROW_HEIGHT = 56;
const PAGE_LIMIT = 100;

/** GANTT-01: шкала created_at → due_date с подсказками зависимостей. */
export function GanttChart({ projectId }: GanttChartProps) {
	// A10-014: cursor pages instead of hard-capped single limit:100 fetch.
	const issuesQuery = useInfiniteQuery({
		queryKey: ["gantt-issues", projectId],
		queryFn: ({ pageParam }) =>
			listIssues(projectId, {
				limit: PAGE_LIMIT,
				order_by: "updated_at",
				cursor: pageParam,
			}),
		initialPageParam: undefined as string | undefined,
		getNextPageParam: nextIssuesCursor,
	});

	useEffect(() => {
		if (issuesQuery.hasNextPage && !issuesQuery.isFetchingNextPage) {
			void issuesQuery.fetchNextPage();
		}
	}, [
		issuesQuery.hasNextPage,
		issuesQuery.isFetchingNextPage,
		issuesQuery.fetchNextPage,
	]);

	// A13-004: один batch вместо N× listIssueLinks per row.
	const { data: projectLinks } = useQuery({
		queryKey: ["project-issue-links", projectId],
		queryFn: () => listProjectIssueLinks(projectId),
	});

	const blockersByKey = useMemo(() => {
		const map = new Map<string, number>();
		for (const link of projectLinks ?? []) {
			if (link.link_type !== "blocked_by") continue;
			const key = link.source_issue_key;
			if (!key) continue;
			map.set(key, (map.get(key) ?? 0) + 1);
		}
		return map;
	}, [projectLinks]);

	const allIssues = useMemo(
		() => flattenIssuePages(issuesQuery.data?.pages),
		[issuesQuery.data?.pages],
	);

	const issuesTotal = issuesQuery.data?.pages?.[0]?.total;
	const loadedCount = allIssues.length;

	const issues = useMemo(
		() => allIssues.filter((i) => i.due_date),
		[allIssues],
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

	// A10-011: QueryStateView — loading/error/empty вместо ad-hoc текста без error.
	return (
		<QueryStateView
			query={issuesQuery}
			loadingFallback={<p className="text-sm text-t3">Загрузка Gantt…</p>}
			isEmpty={() => bars.length === 0 && !issuesQuery.hasNextPage}
			emptyFallback={
				<p className="text-sm text-t3">
					Нет задач с due date — задайте срок в карточке задачи
				</p>
			}
		>
			{() => (
				<div className="space-y-4">
					<p className="text-sm text-t2">
						{totalDays} дн. · {bars.length} задач с дедлайном
						{typeof issuesTotal === "number" && issuesTotal > loadedCount
							? ` · показано ${loadedCount} из ${issuesTotal}`
							: null}
						{issuesQuery.hasNextPage || issuesQuery.isFetchingNextPage
							? " · загружаем…"
							: null}
					</p>
					<GanttBarsList
						projectId={projectId}
						bars={bars}
						rangeStart={rangeStart}
						rangeEnd={rangeEnd}
						blockersByKey={blockersByKey}
					/>
				</div>
			)}
		</QueryStateView>
	);
}

function GanttBarsList({
	projectId,
	bars,
	rangeStart,
	rangeEnd,
	blockersByKey,
}: {
	projectId: string;
	bars: GanttBar[];
	rangeStart: number;
	rangeEnd: number;
	blockersByKey: Map<string, number>;
}) {
	const listRef = useRef<HTMLDivElement>(null);
	const useVirtual = bars.length >= GANTT_VIRTUALIZE_THRESHOLD;
	const span = rangeEnd - rangeStart || 1;

	const virtualizer = useVirtualizer({
		count: useVirtual ? bars.length : 0,
		getScrollElement: () => listRef.current,
		estimateSize: () => ESTIMATED_ROW_HEIGHT,
		overscan: 6,
	});

	return (
		<div
			ref={listRef}
			className="max-h-[min(70vh,720px)] overflow-auto rounded-lg border border-b1"
			data-testid="gantt-bars-list"
		>
			<div className="min-w-[640px]">
				{useVirtual ? (
					<div
						className="relative w-full"
						style={{ height: `${virtualizer.getTotalSize()}px` }}
					>
						{virtualizer.getVirtualItems().map((virtualRow) => {
							const bar = bars[virtualRow.index];
							if (!bar) return null;
							const left = ((bar.start - rangeStart) / span) * 100;
							const width = Math.max(2, ((bar.end - bar.start) / span) * 100);
							return (
								<div
									key={bar.issue.id}
									data-index={virtualRow.index}
									ref={virtualizer.measureElement}
									className="absolute left-0 top-0 w-full"
									style={{
										transform: `translateY(${virtualRow.start}px)`,
									}}
								>
									<GanttRow
										projectId={projectId}
										bar={bar}
										left={left}
										width={width}
										blockers={blockersByKey.get(bar.issue.issue_key) ?? 0}
									/>
								</div>
							);
						})}
					</div>
				) : (
					bars.map((bar) => {
						const left = ((bar.start - rangeStart) / span) * 100;
						const width = Math.max(2, ((bar.end - bar.start) / span) * 100);
						return (
							<GanttRow
								key={bar.issue.id}
								projectId={projectId}
								bar={bar}
								left={left}
								width={width}
								blockers={blockersByKey.get(bar.issue.issue_key) ?? 0}
							/>
						);
					})
				)}
			</div>
		</div>
	);
}

function GanttRow({
	projectId,
	bar,
	left,
	width,
	blockers,
}: {
	projectId: string;
	bar: GanttBar;
	left: number;
	width: number;
	blockers: number;
}) {
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
