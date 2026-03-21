import { createFileRoute } from "@tanstack/react-router";

/** Страница управления метками (заглушка). */
function LabelsSettingsPage() {
  return <div>Загрузка...</div>;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/settings/labels",
)({
  component: LabelsSettingsPage,
});
