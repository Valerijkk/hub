/**
 * Правая колонка детальной страницы задачи -- метаданные.
 */
import { Link } from "@tanstack/react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";

/** Конфигурация цветов приоритетов. */
export const PRIORITY_COLORS: Record<string, string> = {
	critical: "bg-red-500 text-white",
	high: "bg-orange-500 text-white",
	medium: "bg-yellow-500 text-black",
	low: "bg-blue-500 text-white",
	trivial: "bg-gray-400 text-white",
};

interface IssueSidebarMetaProps {
	projectId: string;
	issue: {
		status: { id: string; name: string };
		priority: string;
		assignee?: { name: string; avatar_url?: string } | null;
		reporter?: { name: string; avatar_url?: string } | null;
		labels?: { id: string; name: string; color: string }[];
		estimate_minutes?: number | null;
		time_spent_minutes?: number | null;
		due_date?: string | null;
		parent?: { issue_key: string; title: string } | null;
		children_count: number;
		created_at: string;
		updated_at: string;
		version: number;
	};
	statuses?: { id: string; name: string }[];
	onUpdateField: (fields: Record<string, unknown>) => void;
}

export function IssueSidebarMeta({
	projectId,
	issue,
	statuses,
	onUpdateField,
}: IssueSidebarMetaProps) {
	return (
		<div className="space-y-4">
			{/* Статус */}
			<div className="space-y-1">
				<Label className="text-xs text-muted-foreground">Статус</Label>
				<select
					className="w-full rounded-md border bg-background px-3 py-2 text-sm"
					value={issue.status.id}
					onChange={(e) => onUpdateField({ status_id: e.target.value })}
				>
					{statuses?.map((s) => (
						<option key={s.id} value={s.id}>
							{s.name}
						</option>
					))}
					{!statuses && (
						<option value={issue.status.id}>{issue.status.name}</option>
					)}
				</select>
			</div>

			{/* Приоритет */}
			<div className="space-y-1">
				<Label className="text-xs text-muted-foreground">Приоритет</Label>
				<select
					className="w-full rounded-md border bg-background px-3 py-2 text-sm"
					value={issue.priority}
					onChange={(e) => onUpdateField({ priority: e.target.value })}
				>
					<option value="critical">Критический</option>
					<option value="high">Высокий</option>
					<option value="medium">Средний</option>
					<option value="low">Низкий</option>
					<option value="trivial">Тривиальный</option>
				</select>
				<Badge
					variant="secondary"
					className={cn("text-xs", PRIORITY_COLORS[issue.priority])}
				>
					{issue.priority}
				</Badge>
			</div>

			<Separator />

			{/* Исполнитель */}
			<div className="space-y-1">
				<Label className="text-xs text-muted-foreground">Исполнитель</Label>
				{issue.assignee ? (
					<div className="flex items-center gap-2">
						<Avatar className="h-6 w-6">
							<AvatarImage src={issue.assignee.avatar_url} />
							<AvatarFallback className="text-[10px]">
								{issue.assignee.name.charAt(0).toUpperCase()}
							</AvatarFallback>
						</Avatar>
						<span className="text-sm">{issue.assignee.name}</span>
					</div>
				) : (
					<p className="text-sm text-muted-foreground">Не назначен</p>
				)}
			</div>

			{/* Автор */}
			<div className="space-y-1">
				<Label className="text-xs text-muted-foreground">Автор</Label>
				{issue.reporter ? (
					<div className="flex items-center gap-2">
						<Avatar className="h-6 w-6">
							<AvatarImage src={issue.reporter.avatar_url} />
							<AvatarFallback className="text-[10px]">
								{issue.reporter.name.charAt(0).toUpperCase()}
							</AvatarFallback>
						</Avatar>
						<span className="text-sm">{issue.reporter.name}</span>
					</div>
				) : (
					<p className="text-sm text-muted-foreground">--</p>
				)}
			</div>

			<Separator />

			{/* Метки */}
			<div className="space-y-1">
				<Label className="text-xs text-muted-foreground">Метки</Label>
				<div className="flex flex-wrap gap-1">
					{issue.labels && issue.labels.length > 0 ? (
						issue.labels.map((label) => (
							<Badge
								key={label.id}
								variant="outline"
								className="text-xs"
								style={{ borderColor: label.color, color: label.color }}
							>
								{label.name}
							</Badge>
						))
					) : (
						<p className="text-sm text-muted-foreground">Нет меток</p>
					)}
				</div>
			</div>

			<Separator />

			{/* Оценка и срок */}
			<div className="grid grid-cols-2 gap-3">
				<div className="space-y-1">
					<Label className="text-xs text-muted-foreground">Оценка</Label>
					<p className="text-sm">
						{issue.estimate_minutes
							? `${Math.floor(issue.estimate_minutes / 60)}ч ${issue.estimate_minutes % 60}м`
							: "--"}
					</p>
				</div>
				<div className="space-y-1">
					<Label className="text-xs text-muted-foreground">Потрачено</Label>
					<p className="text-sm">
						{issue.time_spent_minutes
							? `${Math.floor(issue.time_spent_minutes / 60)}ч ${issue.time_spent_minutes % 60}м`
							: "0м"}
					</p>
				</div>
			</div>

			{issue.due_date && (
				<div className="space-y-1">
					<Label className="text-xs text-muted-foreground">Срок</Label>
					<p className="text-sm">
						{new Date(issue.due_date).toLocaleDateString("ru-RU")}
					</p>
				</div>
			)}

			{/* Родительская задача */}
			{issue.parent && (
				<div className="space-y-1">
					<Label className="text-xs text-muted-foreground">
						Родительская задача
					</Label>
					<Link
						to="/projects/$projectId/issues/$issueKey"
						params={{ projectId, issueKey: issue.parent.issue_key }}
						className="text-sm text-primary hover:underline"
					>
						{issue.parent.issue_key}: {issue.parent.title}
					</Link>
				</div>
			)}

			{/* Дочерние задачи */}
			{issue.children_count > 0 && (
				<div className="space-y-1">
					<Label className="text-xs text-muted-foreground">
						Подзадачи ({issue.children_count})
					</Label>
				</div>
			)}

			<Separator />

			{/* Мета */}
			<div className="space-y-1 text-xs text-muted-foreground">
				<p>Создана: {new Date(issue.created_at).toLocaleDateString("ru-RU")}</p>
				<p>
					Обновлена: {new Date(issue.updated_at).toLocaleDateString("ru-RU")}
				</p>
				<p>Версия: {issue.version}</p>
			</div>
		</div>
	);
}
