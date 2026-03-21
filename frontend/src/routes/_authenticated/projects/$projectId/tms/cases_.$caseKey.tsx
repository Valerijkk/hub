import { useState } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getTestCase,
  listVersions,
  rollbackVersion,
  type TestCaseVersionResponse,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

/** Цвета приоритетов. */
const PRIORITY_COLORS: Record<string, string> = {
  critical: "bg-red-500 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-500 text-black",
  low: "bg-blue-500 text-white",
  trivial: "bg-gray-400 text-white",
};

/** Цвета статусов. */
const STATUS_COLORS: Record<string, string> = {
  draft: "bg-gray-200 text-gray-700",
  active: "bg-green-100 text-green-800",
  deprecated: "bg-red-100 text-red-800",
};

/**
 * Страница деталей тест-кейса.
 * Отображает заголовок, шаги, метаданные и историю версий.
 */
function TestCaseDetailPage() {
  const { projectId, caseKey } = Route.useParams();
  const queryClient = useQueryClient();
  const [versionsOpen, setVersionsOpen] = useState(false);

  // Загрузка тест-кейса
  const { data: tc, isLoading } = useQuery({
    queryKey: ["tms-case", projectId, caseKey],
    queryFn: () => getTestCase(projectId, caseKey),
  });

  // Загрузка версий
  const { data: versions } = useQuery({
    queryKey: ["tms-case-versions", projectId, caseKey],
    queryFn: () => listVersions(projectId, caseKey),
    enabled: versionsOpen,
  });

  // Откат
  const rollbackMutation = useMutation({
    mutationFn: (versionId: string) =>
      rollbackVersion(projectId, caseKey, versionId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-case", projectId, caseKey] });
      void queryClient.invalidateQueries({ queryKey: ["tms-case-versions", projectId, caseKey] });
      toast.success("Тест-кейс откачен к версии");
    },
    onError: () => toast.error("Ошибка отката"),
  });

  if (isLoading || !tc) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl">
      {/* Навигация */}
      <div className="mb-4">
        <Link
          to="/projects/$projectId/tms/cases"
          params={{ projectId }}
          className="text-sm text-muted-foreground hover:underline"
        >
          Тест-кейсы
        </Link>
        <span className="mx-2 text-muted-foreground">/</span>
        <span className="text-sm font-medium">{tc.case_key}</span>
      </div>

      <div className="flex gap-6">
        {/* Основная секция */}
        <div className="flex-1 space-y-6">
          {/* Заголовок */}
          <div>
            <div className="mb-2 flex items-center gap-3">
              <span className="text-lg font-bold text-primary">{tc.case_key}</span>
              <Badge
                variant="secondary"
                className={cn("text-xs", PRIORITY_COLORS[tc.priority])}
              >
                {tc.priority}
              </Badge>
              <Badge
                variant="secondary"
                className={cn("text-xs", STATUS_COLORS[tc.status])}
              >
                {tc.status}
              </Badge>
              <Badge variant="outline" className="text-xs">
                {tc.case_type === "bdd" ? "BDD" : "Classic"}
              </Badge>
            </div>
            <h2 className="text-xl font-semibold">{tc.title}</h2>
            {tc.description && (
              <p className="mt-2 text-sm text-muted-foreground">{tc.description}</p>
            )}
          </div>

          {/* Предусловия */}
          {tc.precondition && (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Предусловие</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm">{tc.precondition}</p>
              </CardContent>
            </Card>
          )}

          {/* Шаги */}
          {tc.steps && tc.steps.length > 0 && (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Шаги</CardTitle>
              </CardHeader>
              <CardContent>
                <table className="w-full">
                  <thead>
                    <tr className="border-b">
                      <th className="px-2 py-2 text-left text-xs font-medium text-muted-foreground">
                        #
                      </th>
                      {tc.case_type === "bdd" && (
                        <th className="px-2 py-2 text-left text-xs font-medium text-muted-foreground">
                          Ключевое слово
                        </th>
                      )}
                      <th className="px-2 py-2 text-left text-xs font-medium text-muted-foreground">
                        Действие
                      </th>
                      <th className="px-2 py-2 text-left text-xs font-medium text-muted-foreground">
                        Ожидаемый результат
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {tc.steps.map((step) => (
                      <tr key={step.order} className="border-b last:border-0">
                        <td className="px-2 py-2 text-sm text-muted-foreground">
                          {step.order}
                        </td>
                        {tc.case_type === "bdd" && (
                          <td className="px-2 py-2">
                            <Badge variant="outline" className="text-xs">
                              {step.keyword}
                            </Badge>
                          </td>
                        )}
                        <td className="px-2 py-2 text-sm">{step.action}</td>
                        <td className="px-2 py-2 text-sm text-muted-foreground">
                          {step.expected_result}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          )}

          {/* Постусловие */}
          {tc.postcondition && (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Постусловие</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm">{tc.postcondition}</p>
              </CardContent>
            </Card>
          )}

          {/* История версий */}
          <Card>
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm">
                  История версий (v{tc.version})
                </CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setVersionsOpen(!versionsOpen)}
                >
                  {versionsOpen ? "Скрыть" : "Показать"}
                </Button>
              </div>
            </CardHeader>
            {versionsOpen && versions && (
              <CardContent>
                {versions.length === 0 ? (
                  <p className="text-sm text-muted-foreground">
                    Нет предыдущих версий
                  </p>
                ) : (
                  <div className="space-y-2">
                    {versions.map((v: TestCaseVersionResponse) => (
                      <div
                        key={v.id}
                        className="flex items-center justify-between rounded border p-2"
                      >
                        <div>
                          <span className="text-sm font-medium">
                            Версия {v.version_number}
                          </span>
                          {v.change_description && (
                            <span className="ml-2 text-xs text-muted-foreground">
                              {v.change_description}
                            </span>
                          )}
                          <div className="text-xs text-muted-foreground">
                            {v.created_by?.name} -{" "}
                            {new Date(v.created_at).toLocaleString("ru-RU")}
                          </div>
                        </div>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={rollbackMutation.isPending}
                          onClick={() => rollbackMutation.mutate(v.id)}
                        >
                          Откатить
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            )}
          </Card>
        </div>

        {/* Боковая панель метаданных */}
        <div className="w-64 shrink-0 space-y-4">
          <Card>
            <CardContent className="space-y-3 pt-4">
              {tc.created_by && (
                <div>
                  <div className="text-xs text-muted-foreground">Создал</div>
                  <div className="text-sm">{tc.created_by.name}</div>
                </div>
              )}
              <div>
                <div className="text-xs text-muted-foreground">Создан</div>
                <div className="text-sm">
                  {new Date(tc.created_at).toLocaleString("ru-RU")}
                </div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Обновлён</div>
                <div className="text-sm">
                  {new Date(tc.updated_at).toLocaleString("ru-RU")}
                </div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Версия</div>
                <div className="text-sm">v{tc.version}</div>
              </div>
              {tc.suite && (
                <div>
                  <div className="text-xs text-muted-foreground">Папка</div>
                  <div className="text-sm">{tc.suite.name}</div>
                </div>
              )}
              {tc.components && tc.components.length > 0 && (
                <div>
                  <div className="text-xs text-muted-foreground">Компоненты</div>
                  <div className="flex flex-wrap gap-1">
                    {tc.components.map((c) => (
                      <Badge key={c.id} variant="outline" className="text-xs">
                        {c.name}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
              {tc.labels && tc.labels.length > 0 && (
                <div>
                  <div className="text-xs text-muted-foreground">Метки</div>
                  <div className="flex flex-wrap gap-1">
                    {tc.labels.map((l) => (
                      <Badge
                        key={l.id}
                        variant="secondary"
                        className="text-xs"
                        style={{ backgroundColor: l.color + "20", color: l.color }}
                      >
                        {l.name}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
              {tc.linked_issues && tc.linked_issues.length > 0 && (
                <div>
                  <div className="text-xs text-muted-foreground">
                    Связанные задачи
                  </div>
                  <div className="space-y-1">
                    {tc.linked_issues.map((issue) => (
                      <div key={issue.id} className="text-sm">
                        <Link
                          to="/projects/$projectId/issues/$issueKey"
                          params={{ projectId, issueKey: issue.issue_key }}
                          className="text-primary hover:underline"
                        >
                          {issue.issue_key}
                        </Link>
                        <span className="ml-1 text-muted-foreground">
                          {issue.title}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export const Route = createFileRoute(
  "/_authenticated/projects/$projectId/tms/cases_/$caseKey",
)({
  component: TestCaseDetailPage,
});
