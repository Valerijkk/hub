import { createFileRoute } from "@tanstack/react-router";
import { EnvironmentComparison } from "@/modules/tms/components/environment-comparison";

/**
 * Страница сравнения тестовых окружений.
 * Позволяет выбрать два окружения и сравнить результаты тестов.
 */
function EnvironmentsPage() {
  const { projectId } = Route.useParams();

  return <EnvironmentComparison projectId={projectId} />;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/environments",
)({
  component: EnvironmentsPage,
});
