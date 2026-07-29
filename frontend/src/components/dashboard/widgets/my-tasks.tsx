/**
 * Виджет «Мои задачи» — назначенные задачи с приоритетом и статусом.
 */
import { Circle, CircleCheck, Clock, Inbox } from "lucide-react";
import type { TaskSummary } from "@/api/dashboard";
import {
	type DashboardWidgetData,
	isPersonalData,
} from "@/components/dashboard/widget-data";
import { TaskRowLink } from "@/components/dashboard/widgets/task-row-link";
import { Badge } from "@/components/ui/badge";

/** Цветовые точки приоритетов — semantic tokens (A3W9-2-004). */
const PRIORITY_COLORS: Record<number, string> = {
	5: "bg-[var(--red)]",
	4: "bg-[var(--orange)]",
	3: "bg-[var(--amber)]",
	2: "bg-[var(--blue)]",
	1: "bg-[var(--t3)]",
};

const PRIORITY_LABELS: Record<number, string> = {
	5: "Критический",
	4: "Высокий",
	3: "Средний",
	2: "Низкий",
	1: "Минимальный",
};

/** Стили бейджа статуса. */
function getStatusVariant(
	status: string,
): "default" | "secondary" | "destructive" | "outline" {
	const s = status.toLowerCase();
	if (s.includes("done") || s.includes("закрыт") || s.includes("готов"))
		return "default";
	if (s.includes("progress") || s.includes("работ")) return "secondary";
	if (s.includes("block") || s.includes("блок")) return "destructive";
	return "outline";
}

/** Иконка статуса. */
function StatusIcon({ status }: { status: string }) {
	const s = status.toLowerCase();
	if (s.includes("done") || s.includes("закрыт") || s.includes("готов"))
		return <CircleCheck className="h-3 w-3" />;
	if (s.includes("progress") || s.includes("работ"))
		return <Clock className="h-3 w-3" />;
	return <Circle className="h-3 w-3" />;
}

interface Props {
	data?: DashboardWidgetData;
}

/** Виджет списка задач пользователя. */
export function MyTasksWidget({ data }: Props) {
	const personalData = isPersonalData(data) ? data : undefined;
	const tasks = personalData?.my_tasks ?? [];

	if (tasks.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-t2">
				<Inbox className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет активных задач</p>
				<p className="text-xs opacity-60">Все задачи выполнены</p>
			</div>
		);
	}

	return (
		<div className="space-y-0.5">
			{tasks.slice(0, 10).map((task: TaskSummary) => (
				<TaskRowLink
					key={task.id}
					task={task}
					className="group flex items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-sm no-underline text-inherit transition-colors duration-150 hover:bg-white/[0.04]"
				>
					<div className="flex items-center gap-2.5 min-w-0">
						{/* Точка статуса (8px) */}
						<div
							className={`h-2 w-2 shrink-0 rounded-full ${PRIORITY_COLORS[task.priority] ?? "bg-[var(--t3)]"}`}
							role="img"
							aria-label={PRIORITY_LABELS[task.priority] ?? "Приоритет"}
							title={PRIORITY_LABELS[task.priority] ?? ""}
						/>
						<span className="font-mono text-[11px] text-t2/70 shrink-0 tracking-wide">
							{task.key}
						</span>
						<span className="truncate text-[13px]">{task.title}</span>
					</div>
					<Badge
						variant={getStatusVariant(task.status)}
						className="shrink-0 gap-1 text-[10px] uppercase tracking-wider"
					>
						<StatusIcon status={task.status} />
						{task.status}
					</Badge>
				</TaskRowLink>
			))}
		</div>
	);
}
