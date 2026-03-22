/**
 * Виджет "Дедлайны" -- список задач с приближающимися сроками.
 */
import type { PersonalData, TaskSummary } from "@/api/dashboard";
import { Badge } from "@/components/ui/badge";

interface Props {
	data?: unknown;
}

/** Рассчитывает количество дней до дедлайна. */
function daysRemaining(dueDate: string): number {
	const due = new Date(dueDate);
	const now = new Date();
	return Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

/** Виджет дедлайнов. */
export function DeadlinesWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const deadlines = personalData?.deadlines ?? [];

	if (deadlines.length === 0) {
		return (
			<p className="text-sm text-muted-foreground">Нет ближайших дедлайнов</p>
		);
	}

	return (
		<div className="space-y-2">
			{deadlines.map((task: TaskSummary) => {
				const days = task.due_date ? daysRemaining(task.due_date) : null;
				return (
					<div
						key={task.id}
						className="flex items-center justify-between gap-2 rounded-md border p-2 text-sm"
					>
						<div className="min-w-0">
							<p className="truncate">
								<span className="font-mono text-xs text-muted-foreground mr-1">
									{task.key}
								</span>
								{task.title}
							</p>
						</div>
						{days !== null && (
							<Badge
								variant={
									days <= 1 ? "destructive" : days <= 3 ? "default" : "outline"
								}
								className="shrink-0"
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
