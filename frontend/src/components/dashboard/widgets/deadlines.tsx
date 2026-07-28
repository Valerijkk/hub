/**
 * Виджет «Дедлайны» — задачи с приближающимся due_date.
 */
import { CalendarCheck, CalendarClock } from "lucide-react";
import type { TaskSummary } from "@/api/dashboard";
import {
	type DashboardWidgetData,
	isPersonalData,
} from "@/components/dashboard/widget-data";
import { Badge } from "@/components/ui/badge";

interface Props {
	data?: DashboardWidgetData;
}

/** Дней до дедлайна (отрицательное — просрочка). */
function daysRemaining(dueDate: string): number {
	const due = new Date(dueDate);
	const now = new Date();
	return Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

/** Стили срочности по оставшимся дням. */
function getDeadlineStyle(days: number): {
	variant: "destructive" | "default" | "outline" | "secondary";
	className: string;
} {
	if (days <= 0)
		return {
			variant: "destructive",
			className:
				"bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20",
		};
	if (days <= 1) return { variant: "destructive", className: "" };
	if (days <= 3)
		return {
			variant: "default",
			className:
				"bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20",
		};
	return { variant: "outline", className: "" };
}

/** Виджет дедлайнов. */
export function DeadlinesWidget({ data }: Props) {
	const personalData = isPersonalData(data) ? data : undefined;
	const deadlines = personalData?.deadlines ?? [];

	if (deadlines.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<CalendarCheck className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет ближайших дедлайнов</p>
				<p className="text-xs opacity-60">Все сроки в порядке</p>
			</div>
		);
	}

	return (
		<div className="space-y-1.5">
			{deadlines.map((task: TaskSummary) => {
				const days = task.due_date ? daysRemaining(task.due_date) : null;
				const style = days !== null ? getDeadlineStyle(days) : null;

				return (
					<div
						key={task.id}
						className="flex items-center justify-between gap-2 rounded-lg border border-border/50 p-2.5 text-sm"
					>
						<div className="flex items-center gap-2.5 min-w-0">
							<CalendarClock
								className={`h-4 w-4 shrink-0 ${days !== null && days <= 1 ? "text-red-500" : "text-t2"}`}
							/>
							<div className="min-w-0">
								<p className="truncate">
									<span className="font-mono text-xs text-t2 mr-1.5">
										{task.key}
									</span>
									{task.title}
								</p>
							</div>
						</div>
						{days !== null && style && (
							<Badge
								variant={style.variant}
								className={`shrink-0 tabular-nums ${style.className}`}
							>
								{days <= 0 ? "Просрочено" : `${days}д`}
							</Badge>
						)}
					</div>
				);
			})}
		</div>
	);
}
