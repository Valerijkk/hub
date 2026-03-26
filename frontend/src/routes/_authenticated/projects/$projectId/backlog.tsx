import { createFileRoute } from "@tanstack/react-router";
import { SprintBacklog } from "@/modules/tracker/sprint-backlog";

/**
 * Страница бэклога проекта.
 * Отображает задачи без спринта и позволяет перемещать их в спринты.
 */
function BacklogPage() {
	const { projectId } = Route.useParams();

	return (
		<div className="mx-auto max-w-5xl px-4 py-6">
			<SprintBacklog projectId={projectId} />
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/backlog",
)({
	component: BacklogPage,
});
