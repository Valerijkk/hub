import { createFileRoute } from "@tanstack/react-router";

/** Страница детальной информации о задаче (заглушка). */
function IssueDetailPage() {
  return <div>Загрузка...</div>;
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/issues/$issueKey",
)({
  component: IssueDetailPage,
});
