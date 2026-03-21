import { createFileRoute } from "@tanstack/react-router";
import { SprintBoard } from "@/modules/tracker/sprint-board";

/**
 * Страница спринтов проекта.
 * Отображает список спринтов с управлением жизненным циклом.
 */
function SprintsPage() {
  const { projectId } = Route.useParams();

  return (
    <div className="container mx-auto p-6">
      <SprintBoard projectId={projectId} />
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/sprints/",
)({
  component: SprintsPage,
});
