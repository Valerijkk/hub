/**
 * Виджет "Мои задачи" -- отображает список назначенных задач с приоритетом и статусом.
 */
import type { PersonalData, TaskSummary } from "@/api/dashboard";
import { Badge } from "@/components/ui/badge";

/** Иконки приоритетов. */
const PRIORITY_ICONS: Record<number, string> = {
	5: "!!!",
	4: "!!",
	3: "!",
	2: "-",
	1: ".",
};

interface Props {
	data?: unknown;
}

/** Виджет списка задач пользователя. */
export function MyTasksWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const tasks = personalData?.my_tasks ?? [];

	if (tasks.length === 0) {
		return <p className="text-sm text-muted-foreground">Нет активных задач</p>;
	}

	return (
		<div className="space-y-2">
			{tasks.slice(0, 10).map((task: TaskSummary) => (
				<div
					key={task.id}
					className="flex items-center justify-between gap-2 rounded-md border p-2 text-sm"
				>
					<div className="flex items-center gap-2 min-w-0">
						<span className="text-xs font-mono text-muted-foreground whitespace-nowrap">
							{PRIORITY_ICONS[task.priority] ?? ""}
						</span>
						<span className="font-mono text-xs text-muted-foreground">
							{task.key}
						</span>
						<span className="truncate">{task.title}</span>
					</div>
					<Badge variant="outline" className="shrink-0">
						{task.status}
					</Badge>
				</div>
			))}
		</div>
	);
}
