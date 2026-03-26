/**
 * Widget "My Tasks" -- displays a list of assigned tasks with priority and status.
 */
import { CircleCheck, Circle, Clock, Inbox } from "lucide-react";
import type { PersonalData, TaskSummary } from "@/api/dashboard";
import { Badge } from "@/components/ui/badge";

/** Priority color dots. */
const PRIORITY_COLORS: Record<number, string> = {
	5: "bg-red-500",
	4: "bg-orange-500",
	3: "bg-yellow-500",
	2: "bg-blue-400",
	1: "bg-slate-400",
};

const PRIORITY_LABELS: Record<number, string> = {
	5: "Критический",
	4: "Высокий",
	3: "Средний",
	2: "Низкий",
	1: "Минимальный",
};

/** Status badge styling. */
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

/** Status icon. */
function StatusIcon({ status }: { status: string }) {
	const s = status.toLowerCase();
	if (s.includes("done") || s.includes("закрыт") || s.includes("готов"))
		return <CircleCheck className="h-3 w-3" />;
	if (s.includes("progress") || s.includes("работ"))
		return <Clock className="h-3 w-3" />;
	return <Circle className="h-3 w-3" />;
}

interface Props {
	data?: unknown;
}

/** User task list widget. */
export function MyTasksWidget({ data }: Props) {
	const personalData = data as PersonalData | undefined;
	const tasks = personalData?.my_tasks ?? [];

	if (tasks.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center h-full gap-2 text-muted-foreground">
				<Inbox className="h-8 w-8 opacity-40" />
				<p className="text-sm">Нет активных задач</p>
				<p className="text-xs opacity-60">
					Все задачи выполнены
				</p>
			</div>
		);
	}

	return (
		<div className="space-y-0.5">
			{tasks.slice(0, 10).map((task: TaskSummary) => (
				<div
					key={task.id}
					className="group flex items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors duration-150 hover:bg-white/[0.04]"
				>
					<div className="flex items-center gap-2.5 min-w-0">
						{/* Status dot (8px) */}
						<div
							className={`h-2 w-2 shrink-0 rounded-full ${PRIORITY_COLORS[task.priority] ?? "bg-slate-400"}`}
							title={PRIORITY_LABELS[task.priority] ?? ""}
						/>
						<span className="font-mono text-[11px] text-muted-foreground/70 shrink-0 tracking-wide">
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
				</div>
			))}
		</div>
	);
}
