import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import {
	type IssueResponse,
	type ListIssuesParams,
	listIssues,
} from "@/api/tracker";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

/** Конфигурация цветов приоритетов. */
const PRIORITY_COLORS: Record<string, string> = {
	critical: "bg-red-500 text-white",
	high: "bg-orange-500 text-white",
	medium: "bg-yellow-500 text-black",
	low: "bg-blue-500 text-white",
	trivial: "bg-gray-400 text-white",
};

/** Конфигурация иконок типов задач. */
const TYPE_ICONS: Record<string, string> = {
	bug: "\u{1F41B}",
	task: "\u{2705}",
	story: "\u{1F4D6}",
	epic: "\u{26A1}",
	sub_task: "\u{1F4CB}",
	improvement: "\u{1F680}",
	tech_debt: "\u{1F527}",
	spike: "\u{1F50D}",
	release: "\u{1F3F7}\uFE0F",
};

/**
 * Страница списка задач проекта.
 * Отображает задачи в табличном виде с фильтрацией и курсорной пагинацией.
 */
function IssueListPage() {
	const { projectId } = Route.useParams();

	const [search, setSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState<string>("");
	const [priorityFilter, setPriorityFilter] = useState<string>("");
	const [typeFilter, setTypeFilter] = useState<string>("");
	const [cursor, setCursor] = useState<string | undefined>();
	const [allIssues, setAllIssues] = useState<IssueResponse[]>([]);

	const params: ListIssuesParams = {
		cursor,
		limit: 50,
		search: search || undefined,
		status: statusFilter ? [statusFilter] : undefined,
		priority: priorityFilter ? [priorityFilter] : undefined,
		issue_type: typeFilter || undefined,
	};

	const { data, isLoading } = useQuery({
		queryKey: ["issues", projectId, params],
		queryFn: () => listIssues(projectId, params),
	});

	// Для cursor-based: накапливаем задачи
	const displayIssues = cursor ? allIssues : (data?.items ?? []);

	const loadMore = useCallback(() => {
		if (data?.has_more && data.next_cursor) {
			setAllIssues((prev) => [...prev, ...(data.items ?? [])]);
			setCursor(data.next_cursor);
		}
	}, [data]);

	const resetFilters = () => {
		setSearch("");
		setStatusFilter("");
		setPriorityFilter("");
		setTypeFilter("");
		setCursor(undefined);
		setAllIssues([]);
	};

	return (
		<div>
			{/* Заголовок */}
			<div className="mb-6 flex items-center justify-between">
				<h1 className="text-2xl font-bold">Задачи</h1>
				<Button asChild>
					<Link to="/projects/$projectId/issues/new" params={{ projectId }}>
						Новая задача
					</Link>
				</Button>
			</div>

			{/* Фильтры */}
			<div className="mb-4 flex flex-wrap items-center gap-3">
				<Input
					placeholder="Поиск..."
					className="w-48"
					value={search}
					onChange={(e) => {
						setSearch(e.target.value);
						setCursor(undefined);
						setAllIssues([]);
					}}
				/>
				<select
					className="rounded-md border bg-background px-3 py-1.5 text-sm"
					value={priorityFilter}
					onChange={(e) => {
						setPriorityFilter(e.target.value);
						setCursor(undefined);
						setAllIssues([]);
					}}
				>
					<option value="">Все приоритеты</option>
					<option value="critical">Критический</option>
					<option value="high">Высокий</option>
					<option value="medium">Средний</option>
					<option value="low">Низкий</option>
					<option value="trivial">Тривиальный</option>
				</select>
				<select
					className="rounded-md border bg-background px-3 py-1.5 text-sm"
					value={typeFilter}
					onChange={(e) => {
						setTypeFilter(e.target.value);
						setCursor(undefined);
						setAllIssues([]);
					}}
				>
					<option value="">Все типы</option>
					<option value="bug">Баг</option>
					<option value="task">Задача</option>
					<option value="story">История</option>
					<option value="epic">Эпик</option>
					<option value="sub_task">Подзадача</option>
					<option value="improvement">Улучшение</option>
					<option value="tech_debt">Техдолг</option>
					<option value="spike">Исследование</option>
					<option value="release">Релиз</option>
				</select>

				{(search || statusFilter || priorityFilter || typeFilter) && (
					<Button variant="ghost" size="sm" onClick={resetFilters}>
						Сбросить
					</Button>
				)}
			</div>

			{/* Таблица задач */}
			{isLoading && !cursor ? (
				<div className="animate-fade-in rounded-md border">
					{/* Table header skeleton */}
					<div className="flex items-center gap-4 border-b bg-muted/30 px-4 py-3">
						<div className="h-4 w-20 rounded animate-shimmer" />
						<div className="h-4 w-12 rounded animate-shimmer" />
						<div className="h-4 w-48 rounded animate-shimmer flex-1" />
						<div className="h-4 w-20 rounded animate-shimmer" />
						<div className="h-4 w-20 rounded animate-shimmer" />
						<div className="h-4 w-16 rounded animate-shimmer" />
					</div>
					{/* Table rows skeleton */}
					{[1, 2, 3, 4, 5, 6].map((i) => (
						<div
							key={i}
							className="flex items-center gap-4 border-b last:border-b-0 px-4 py-3"
						>
							<div className="h-4 w-20 rounded animate-shimmer" />
							<div className="h-5 w-5 rounded animate-shimmer" />
							<div
								className="h-4 rounded animate-shimmer flex-1"
								style={{ width: `${40 + ((i * 7) % 30)}%` }}
							/>
							<div className="h-5 w-16 rounded-full animate-shimmer" />
							<div className="h-5 w-16 rounded-full animate-shimmer" />
							<div className="h-6 w-6 rounded-full animate-shimmer" />
						</div>
					))}
				</div>
			) : displayIssues.length === 0 ? (
				<div className="flex flex-col items-center justify-center py-12">
					<p className="mb-4 text-lg text-muted-foreground">
						Задачи не найдены
					</p>
					<Button asChild>
						<Link to="/projects/$projectId/issues/new" params={{ projectId }}>
							Создать задачу
						</Link>
					</Button>
				</div>
			) : (
				<>
					<div className="rounded-md border">
						<table className="w-full">
							<thead>
								<tr className="border-b bg-muted/50">
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Ключ
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Название
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Тип
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Приоритет
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Статус
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Исполнитель
									</th>
									<th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">
										Создана
									</th>
								</tr>
							</thead>
							<tbody>
								{displayIssues.map((issue) => (
									<tr
										key={issue.id}
										className="border-b transition-colors hover:bg-muted/50"
									>
										<td className="px-4 py-3">
											<Link
												to="/projects/$projectId/issues/$issueKey"
												params={{
													projectId,
													issueKey: issue.issue_key,
												}}
												className="text-sm font-bold text-primary hover:underline"
											>
												{issue.issue_key}
											</Link>
										</td>
										<td className="max-w-md truncate px-4 py-3 text-sm">
											<Link
												to="/projects/$projectId/issues/$issueKey"
												params={{
													projectId,
													issueKey: issue.issue_key,
												}}
												className="hover:underline"
											>
												{issue.title}
											</Link>
										</td>
										<td className="px-4 py-3 text-sm">
											<span title={issue.issue_type}>
												{TYPE_ICONS[issue.issue_type] ?? "\u{1F4CB}"}{" "}
												{issue.issue_type}
											</span>
										</td>
										<td className="px-4 py-3">
											<Badge
												variant="secondary"
												className={cn(
													"text-xs",
													PRIORITY_COLORS[issue.priority],
												)}
											>
												{issue.priority}
											</Badge>
										</td>
										<td className="px-4 py-3">
											<Badge variant="outline" className="text-xs">
												{issue.status.name}
											</Badge>
										</td>
										<td className="px-4 py-3">
											{issue.assignee ? (
												<div className="flex items-center gap-2">
													<Avatar className="h-5 w-5">
														<AvatarImage src={issue.assignee.avatar_url} />
														<AvatarFallback className="text-[10px]">
															{issue.assignee.name.charAt(0).toUpperCase()}
														</AvatarFallback>
													</Avatar>
													<span className="text-sm">{issue.assignee.name}</span>
												</div>
											) : (
												<span className="text-xs text-muted-foreground">
													Не назначен
												</span>
											)}
										</td>
										<td className="px-4 py-3 text-xs text-muted-foreground">
											{new Date(issue.created_at).toLocaleDateString("ru-RU")}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{data?.has_more && (
						<div className="mt-4 flex justify-center">
							<Button variant="outline" onClick={loadMore} disabled={isLoading}>
								Загрузить ещё
							</Button>
						</div>
					)}
				</>
			)}
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/issues/",
)({
	component: IssueListPage,
});
