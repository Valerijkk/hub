import { createFileRoute } from "@tanstack/react-router";
import { GanttChart } from "@/modules/tracker/gantt-chart";

function GanttPage() {
	const { projectId } = Route.useParams();
	return (
		<div className="mx-auto max-w-5xl space-y-6 px-4 py-6">
			<div>
				<h1 className="text-2xl font-bold tracking-tight">Gantt</h1>
				<p className="text-sm text-t2 mt-1">
					Сроки задач и зависимости (blocked_by)
				</p>
			</div>
			<GanttChart projectId={projectId} />
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/gantt",
)({
	component: GanttPage,
});
