import { createFileRoute } from "@tanstack/react-router";
import { SprintBoard } from "@/modules/tracker/sprint-board";

/**
 * Страница спринтов проекта.
 * Отображает список спринтов с управлением жизненным циклом.
 */
function SprintsPage() {
	const { projectId } = Route.useParams();

	return (
		<div className="mx-auto max-w-5xl px-4 py-6 animate-[eu_450ms_var(--spring)_both]">
			<SprintBoard projectId={projectId} />
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/sprints/",
)({
	component: SprintsPage,
});
