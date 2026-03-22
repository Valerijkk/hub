import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { getSprint, type SprintResponse } from "@/api/sprint";
import { BurndownChart } from "@/modules/tracker/burndown-chart";
import { SprintRetro } from "@/modules/tracker/sprint-retro";

/**
 * Страница детального просмотра спринта.
 * Включает burndown-график и ретроспективу.
 */
function SprintDetailPage() {
	const { projectId, sprintId } = Route.useParams();
	const [sprint, setSprint] = useState<SprintResponse | null>(null);

	useEffect(() => {
		getSprint(projectId, sprintId).then(setSprint);
	}, [projectId, sprintId]);

	if (!sprint) {
		return (
			<div className="flex items-center justify-center p-8">
				<p className="text-muted-foreground">Загрузка...</p>
			</div>
		);
	}

	const retroNotes = sprint.retro_notes as
		| {
				went_well?: string[];
				to_improve?: string[];
				actions?: string[];
		  }
		| undefined;

	return (
		<div className="container mx-auto space-y-6 p-6">
			<div>
				<h1 className="text-2xl font-bold">{sprint.name}</h1>
				{sprint.goal && <p className="text-muted-foreground">{sprint.goal}</p>}
				<p className="text-sm text-muted-foreground">
					{new Date(sprint.start_date).toLocaleDateString()} -{" "}
					{new Date(sprint.end_date).toLocaleDateString()} | Задач:{" "}
					{sprint.issue_count}
				</p>
			</div>

			<BurndownChart projectId={projectId} sprintId={sprintId} />

			{(sprint.status === "active" || sprint.status === "completed") && (
				<SprintRetro
					projectId={projectId}
					sprintId={sprintId}
					initialNotes={retroNotes}
				/>
			)}
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/sprints/$sprintId",
)({
	component: SprintDetailPage,
});
