import { createFileRoute } from "@tanstack/react-router";
import { TraceabilityMatrix } from "@/modules/tms/components/traceability-matrix";

/**
 * Страница матрицы трассировки.
 * Отображает связь задач (строки) с тест-кейсами (столбцы)
 * и статус последнего выполнения в каждой ячейке.
 */
function TraceabilityPage() {
  const { projectId } = Route.useParams();

  return <TraceabilityMatrix projectId={projectId} />;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/traceability",
)({
  component: TraceabilityPage,
});
