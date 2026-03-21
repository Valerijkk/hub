import { useState, useEffect } from "react";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getTestCase,
  updateTestCase,
  listVersions,
  rollbackVersion,
  linkTestCaseToIssue,
  unlinkTestCaseFromIssue,
  type TestCaseVersionResponse,
  type TestStep,
  type TestCasePriority,
  type TestCaseStatus,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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

/** Лейблы приоритетов. */
const PRIORITY_LABELS: Record<string, string> = {
  critical: "Критический",
  high: "Высокий",
  medium: "Средний",
  low: "Низкий",
  trivial: "Тривиальный",
};

/** Лейблы статусов. */
const STATUS_LABELS: Record<string, string> = {
  draft: "Черновик",
  active: "Активный",
  deprecated: "Устаревший",
};

/**
 * Страница деталей тест-кейса.
 * Отображает заголовок, шаги, метаданные и историю версий.
 * Поддерживает редактирование всех полей тест-кейса.
 */
function TestCaseDetailPage() {
  const { projectId, caseKey } = Route.useParams();
  const queryClient = useQueryClient();
  const [versionsOpen, setVersionsOpen] = useState(false);
  const [linkIssueId, setLinkIssueId] = useState("");
  const [editing, setEditing] = useState(false);

  // Поля редактирования
  const [editTitle, setEditTitle] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editPrecondition, setEditPrecondition] = useState("");
  const [editPostcondition, setEditPostcondition] = useState("");
  const [editPriority, setEditPriority] = useState<TestCasePriority>("medium");
  const [editStatus, setEditStatus] = useState<TestCaseStatus>("draft");
  const [editSteps, setEditSteps] = useState<TestStep[]>([]);

  // Загрузка тест-кейса
  const { data: tc, isLoading } = useQuery({
    queryKey: ["tms-case", projectId, caseKey],
    queryFn: () => getTestCase(projectId, caseKey),
  });

  // Инициализация формы при загрузке данных
  useEffect(() => {
    if (tc) {
      setEditTitle(tc.title);
      setEditDescription(tc.description ?? "");
      setEditPrecondition(tc.precondition ?? "");
      setEditPostcondition(tc.postcondition ?? "");
      setEditPriority(tc.priority);
      setEditStatus(tc.status);
      setEditSteps(tc.steps ? tc.steps.map((s) => ({ ...s })) : []);
    }
  }, [tc]);

  // Загрузка версий
  const { data: versions } = useQuery({
    queryKey: ["tms-case-versions", projectId, caseKey],
    queryFn: () => listVersions(projectId, caseKey),
    enabled: versionsOpen,
  });

  // Обновление тест-кейса
  const updateMutation = useMutation({
    mutationFn: () =>
      updateTestCase(projectId, caseKey, {
        title: editTitle,
        description: editDescription || undefined,
        precondition: editPrecondition || undefined,
        postcondition: editPostcondition || undefined,
        priority: editPriority,
        status: editStatus,
        steps: editSteps.length > 0 ? editSteps : undefined,
        version: tc!.version,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-case", projectId, caseKey] });
      void queryClient.invalidateQueries({ queryKey: ["tms-cases", projectId] });
      setEditing(false);
      toast.success("Тест-кейс обновлён");
    },
    onError: () => toast.error("Ошибка обновления тест-кейса"),
  });

  // Привязка к задаче
  const linkMutation = useMutation({
    mutationFn: (issueId: string) =>
      linkTestCaseToIssue(projectId, caseKey, issueId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-case", projectId, caseKey] });
      setLinkIssueId("");
      toast.success("Задача привязана");
    },
    onError: () => toast.error("Ошибка привязки"),
  });

  // Отвязка от задачи
  const unlinkMutation = useMutation({
    mutationFn: (issueId: string) =>
      unlinkTestCaseFromIssue(projectId, caseKey, issueId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-case", projectId, caseKey] });
      toast.success("Задача отвязана");
    },
    onError: () => toast.error("Ошибка отвязки"),
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

  /** Начать редактирование — инициализировать форму из текущих данных. */
  const startEditing = () => {
    if (!tc) return;
    setEditTitle(tc.title);
    setEditDescription(tc.description ?? "");
    setEditPrecondition(tc.precondition ?? "");
    setEditPostcondition(tc.postcondition ?? "");
    setEditPriority(tc.priority);
    setEditStatus(tc.status);
    setEditSteps(tc.steps ? tc.steps.map((s) => ({ ...s })) : []);
    setEditing(true);
  };

  /** Отменить редактирование. */
  const cancelEditing = () => {
    setEditing(false);
  };

  /** Добавить шаг в конец списка. */
  const addStep = () => {
    setEditSteps([
      ...editSteps,
      { order: editSteps.length + 1, action: "", expected_result: "" },
    ]);
  };

  /** Удалить шаг и пересчитать порядок. */
  const removeStep = (index: number) => {
    const updated = editSteps.filter((_, i) => i !== index);
    setEditSteps(updated.map((s, i) => ({ ...s, order: i + 1 })));
  };

  /** Обновить поле шага. */
  const updateStep = (index: number, field: "action" | "expected_result", value: string) => {
    setEditSteps((prev) =>
      prev.map((s, i) =>
        i === index ? { order: s.order, action: field === "action" ? value : s.action, expected_result: field === "expected_result" ? value : s.expected_result, keyword: s.keyword } : s,
      ),
    );
  };

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
      <div className="mb-4 flex items-center justify-between">
        <div>
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
        <div className="flex gap-2">
          {editing ? (
            <>
              <Button
                variant="outline"
                size="sm"
                onClick={cancelEditing}
                disabled={updateMutation.isPending}
              >
                Отмена
              </Button>
              <Button
                size="sm"
                onClick={() => updateMutation.mutate()}
                disabled={!editTitle.trim() || updateMutation.isPending}
              >
                {updateMutation.isPending ? "Сохранение..." : "Сохранить"}
              </Button>
            </>
          ) : (
            <Button variant="outline" size="sm" onClick={startEditing}>
              <svg className="mr-1.5 h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
              Редактировать
            </Button>
          )}
        </div>
      </div>

      <div className="flex gap-6">
        {/* Основная секция */}
        <div className="flex-1 space-y-6">
          {/* Заголовок */}
          <div>
            <div className="mb-2 flex items-center gap-3">
              <span className="text-lg font-bold text-primary">{tc.case_key}</span>
              {editing ? (
                <>
                  <select
                    className="rounded-md border bg-card px-2 py-1 text-xs"
                    value={editPriority}
                    onChange={(e) => setEditPriority(e.target.value as TestCasePriority)}
                  >
                    {Object.entries(PRIORITY_LABELS).map(([value, label]) => (
                      <option key={value} value={value}>{label}</option>
                    ))}
                  </select>
                  <select
                    className="rounded-md border bg-card px-2 py-1 text-xs"
                    value={editStatus}
                    onChange={(e) => setEditStatus(e.target.value as TestCaseStatus)}
                  >
                    {Object.entries(STATUS_LABELS).map(([value, label]) => (
                      <option key={value} value={value}>{label}</option>
                    ))}
                  </select>
                </>
              ) : (
                <>
                  <Badge
                    variant="secondary"
                    className={cn("text-xs", PRIORITY_COLORS[tc.priority])}
                  >
                    {PRIORITY_LABELS[tc.priority] ?? tc.priority}
                  </Badge>
                  <Badge
                    variant="secondary"
                    className={cn("text-xs", STATUS_COLORS[tc.status])}
                  >
                    {STATUS_LABELS[tc.status] ?? tc.status}
                  </Badge>
                </>
              )}
              <Badge variant="outline" className="text-xs">
                {tc.case_type === "bdd" ? "BDD" : "Classic"}
              </Badge>
            </div>
            {editing ? (
              <Input
                className="text-xl font-semibold"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                placeholder="Название тест-кейса"
              />
            ) : (
              <h2 className="text-xl font-semibold">{tc.title}</h2>
            )}
            {editing ? (
              <textarea
                className="mt-2 w-full rounded-lg border bg-card px-3 py-2 text-sm"
                placeholder="Описание (необязательно)"
                rows={3}
                value={editDescription}
                onChange={(e) => setEditDescription(e.target.value)}
              />
            ) : (
              tc.description && (
                <p className="mt-2 text-sm text-muted-foreground">{tc.description}</p>
              )
            )}
          </div>

          {/* Предусловия */}
          {editing ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Предусловие</CardTitle>
              </CardHeader>
              <CardContent>
                <textarea
                  className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                  placeholder="Предусловие (необязательно)"
                  rows={2}
                  value={editPrecondition}
                  onChange={(e) => setEditPrecondition(e.target.value)}
                />
              </CardContent>
            </Card>
          ) : (
            tc.precondition && (
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Предусловие</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{tc.precondition}</p>
                </CardContent>
              </Card>
            )
          )}

          {/* Шаги */}
          {editing ? (
            <Card>
              <CardHeader className="pb-2">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-sm">Шаги</CardTitle>
                  <Button variant="outline" size="sm" onClick={addStep}>
                    + Добавить шаг
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {editSteps.length === 0 ? (
                  <p className="text-sm text-muted-foreground">
                    Нет шагов. Нажмите "Добавить шаг".
                  </p>
                ) : (
                  <div className="space-y-3">
                    {editSteps.map((step, idx) => (
                      <div key={idx} className="flex gap-2 items-start rounded border p-2">
                        <span className="mt-2 text-sm font-medium text-muted-foreground w-6 shrink-0">
                          {step.order}
                        </span>
                        <div className="flex-1 space-y-2">
                          <Input
                            placeholder="Действие"
                            value={step.action}
                            onChange={(e) => updateStep(idx, "action", e.target.value)}
                          />
                          <Input
                            placeholder="Ожидаемый результат"
                            value={step.expected_result ?? ""}
                            onChange={(e) => updateStep(idx, "expected_result", e.target.value)}
                          />
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="mt-1 h-7 px-2 text-xs text-destructive"
                          onClick={() => removeStep(idx)}
                        >
                          x
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ) : (
            tc.steps && tc.steps.length > 0 && (
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
            )
          )}

          {/* Постусловие */}
          {editing ? (
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm">Постусловие</CardTitle>
              </CardHeader>
              <CardContent>
                <textarea
                  className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
                  placeholder="Постусловие (необязательно)"
                  rows={2}
                  value={editPostcondition}
                  onChange={(e) => setEditPostcondition(e.target.value)}
                />
              </CardContent>
            </Card>
          ) : (
            tc.postcondition && (
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">Постусловие</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{tc.postcondition}</p>
                </CardContent>
              </Card>
            )
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
              {/* Связанные задачи */}
              <div>
                <div className="mb-1 flex items-center justify-between">
                  <span className="text-xs text-muted-foreground">
                    Связанные задачи
                  </span>
                </div>
                {tc.linked_issues && tc.linked_issues.length > 0 && (
                  <div className="mb-2 space-y-1">
                    {tc.linked_issues.map((issue) => (
                      <div
                        key={issue.id}
                        className="flex items-center justify-between text-sm"
                      >
                        <div>
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
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-5 px-1 text-xs text-muted-foreground"
                          disabled={unlinkMutation.isPending}
                          onClick={() => unlinkMutation.mutate(issue.id)}
                        >
                          x
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
                <div className="flex gap-1">
                  <input
                    type="text"
                    className="flex-1 rounded-md border bg-background px-2 py-1 text-xs"
                    placeholder="ID задачи"
                    value={linkIssueId}
                    onChange={(e) => setLinkIssueId(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && linkIssueId.trim()) {
                        linkMutation.mutate(linkIssueId.trim());
                      }
                    }}
                  />
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-7 px-2 text-xs"
                    disabled={!linkIssueId.trim() || linkMutation.isPending}
                    onClick={() => linkMutation.mutate(linkIssueId.trim())}
                  >
                    +
                  </Button>
                </div>
              </div>
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
