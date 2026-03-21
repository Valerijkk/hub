import { createFileRoute } from "@tanstack/react-router";

/** Страница списка задач проекта (заглушка). */
function IssueListPage() {
  return <div>Загрузка...</div>;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/issues/",
)({
  component: IssueListPage,
});
