import { createFileRoute } from "@tanstack/react-router";

/** Страница настроек workflow (заглушка). */
function WorkflowSettingsPage() {
  return <div>Загрузка...</div>;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/settings/workflow",
)({
  component: WorkflowSettingsPage,
});
