/**
 * Правая колонка детальной страницы задачи -- метаданные.
 */
import { Link } from "@tanstack/react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
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
		story_points?: number | null;
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
		<div className="w-[320px] min-w-[320px] border-l border-b1 overflow-y-auto p-[22px] bg-void">
			{/* Статус */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Статус
				</div>
				<Select
					value={issue.status.id}
					onValueChange={(v) => onUpdateField({ status_id: v })}
				>
					<SelectTrigger className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px] w-full border-0 bg-transparent">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{statuses?.map((s) => (
							<SelectItem key={s.id} value={s.id}>
								{s.name}
							</SelectItem>
						))}
						{!statuses && (
							<SelectItem value={issue.status.id}>
								{issue.status.name}
							</SelectItem>
						)}
					</SelectContent>
				</Select>
			</div>

			{/* Приоритет */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Приоритет
				</div>
				<Select
					value={issue.priority}
					onValueChange={(v) => onUpdateField({ priority: v })}
				>
					<SelectTrigger className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px] w-full border-0 bg-transparent">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="critical">Критический</SelectItem>
						<SelectItem value="high">Высокий</SelectItem>
						<SelectItem value="medium">Средний</SelectItem>
						<SelectItem value="low">Низкий</SelectItem>
						<SelectItem value="trivial">Тривиальный</SelectItem>
					</SelectContent>
				</Select>
				<Badge
					variant="secondary"
					className={cn(
						"inline-flex items-center gap-[5px] px-[10px] py-[3px] rounded-full text-[11px] font-[650]",
						issue.priority === "critical" || issue.priority === "high"
							? "bg-red/[0.08] text-red border border-red/[0.15]"
							: "bg-blue/[0.08] text-blue border border-blue/[0.15]",
					)}
				>
					{issue.priority}
				</Badge>
			</div>

			{/* Исполнитель */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Исполнитель
				</div>
				{issue.assignee ? (
					<div className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px]">
						<Avatar className="h-7 w-7">
							<AvatarImage src={issue.assignee.avatar_url} />
							<AvatarFallback className="text-[11px] font-bold text-white bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)]">
								{issue.assignee.name.charAt(0).toUpperCase()}
							</AvatarFallback>
						</Avatar>
						<span className="text-[14px]">{issue.assignee.name}</span>
					</div>
				) : (
					<p className="text-[14px] text-t3 py-[7px] px-[10px]">Не назначен</p>
				)}
			</div>

			{/* Автор */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Автор
				</div>
				{issue.reporter ? (
					<div className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px]">
						<Avatar className="h-7 w-7">
							<AvatarImage src={issue.reporter.avatar_url} />
							<AvatarFallback className="text-[11px] font-bold text-white bg-gradient-to-br from-[var(--greenS)] to-[#047857]">
								{issue.reporter.name.charAt(0).toUpperCase()}
							</AvatarFallback>
						</Avatar>
						<span className="text-[14px]">{issue.reporter.name}</span>
					</div>
				) : (
					<p className="text-[14px] text-t3 py-[7px] px-[10px]">--</p>
				)}
			</div>

			{/* Метки */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Метки
				</div>
				<div className="flex flex-wrap gap-[6px]">
					{issue.labels && issue.labels.length > 0 ? (
						issue.labels.map((label) => (
							<span
								key={label.id}
								className="px-3 py-[3px] rounded-full text-[12px] font-semibold border"
								style={{ borderColor: label.color, color: label.color }}
							>
								{label.name}
							</span>
						))
					) : (
						<p className="text-[14px] text-t3">Нет меток</p>
					)}
				</div>
			</div>

			{/* Оценка сложности (SP) */}
			<div className="mb-[18px]">
				<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
					Story Points
				</div>
				<Select
					value={
						issue.story_points != null ? String(issue.story_points) : "none"
					}
					onValueChange={(v) =>
						onUpdateField({
							story_points: v === "none" ? undefined : Number(v),
						})
					}
				>
					<SelectTrigger className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px] w-full border-0 bg-transparent">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="none">--</SelectItem>
						{[1, 2, 3, 5, 8, 13, 21].map((sp) => (
							<SelectItem key={sp} value={String(sp)}>
								{sp}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>

			{issue.due_date && (
				<div className="mb-[18px]">
					<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
						Срок
					</div>
					<p className="text-[14px] py-[7px] px-[10px]">
						{new Date(issue.due_date).toLocaleDateString("ru-RU")}
					</p>
				</div>
			)}

			{/* Родительская задача */}
			{issue.parent && (
				<div className="mb-[18px]">
					<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
						Родительская задача
					</div>
					<Link
						to="/projects/$projectId/issues/$issueKey"
						params={{ projectId, issueKey: issue.parent.issue_key }}
						className="flex items-center gap-[9px] py-[7px] px-[10px] rounded-[var(--r2)] cursor-pointer transition-colors duration-180 hover:bg-[var(--hover)] text-[14px] text-blue"
					>
						{issue.parent.issue_key}: {issue.parent.title}
					</Link>
				</div>
			)}

			{/* Дочерние задачи */}
			{issue.children_count > 0 && (
				<div className="mb-[18px]">
					<div className="text-[10.5px] font-bold uppercase tracking-[0.06em] text-t4 mb-[7px]">
						Подзадачи ({issue.children_count})
					</div>
				</div>
			)}

			{/* Мета */}
			<div className="mt-[22px] pt-[18px] border-t border-b1">
				<div className="text-[12px] text-t4 py-1">
					Создана: {new Date(issue.created_at).toLocaleDateString("ru-RU")}
				</div>
				<div className="text-[12px] text-t4 py-1">
					Обновлена: {new Date(issue.updated_at).toLocaleDateString("ru-RU")}
				</div>
				<div className="text-[12px] text-t4 py-1">Версия: {issue.version}</div>
			</div>
		</div>
	);
}
