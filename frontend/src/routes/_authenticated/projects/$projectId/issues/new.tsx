import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { IssueForm } from "@/components/tracker/issue-form";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

/**
 * Страница создания новой задачи.
 * Отображает форму создания и перенаправляет на детали после успеха.
 */
function NewIssuePage() {
  const { projectId } = Route.useParams();
  const navigate = useNavigate();

  return (
    <div className="mx-auto max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>Новая задача</CardTitle>
        </CardHeader>
        <CardContent>
          <IssueForm
            projectId={projectId}
            mode="create"
            onSuccess={(issue) => {
              void navigate({
                to: "/projects/$projectId/issues/$issueKey",
                params: { projectId, issueKey: issue.issue_key },
              });
            }}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/issues/new",
)({
  component: NewIssuePage,
});
