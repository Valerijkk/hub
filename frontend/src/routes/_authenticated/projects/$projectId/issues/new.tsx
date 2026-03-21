import { createFileRoute } from "@tanstack/react-router";

/** Страница создания задачи (заглушка). */
function NewIssuePage() {
  return <div>Загрузка...</div>;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/issues/new",
)({
  component: NewIssuePage,
});
