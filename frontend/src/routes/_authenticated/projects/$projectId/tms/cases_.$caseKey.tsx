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

/** Цвета приоритетов — бейджи. */
const PRIORITY_CONFIG: Record<string, { dot: string; bg: string; text: string; label: string }> = {
  critical: { dot: "bg-red-500", bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Критический" },
  high: { dot: "bg-orange-500", bg: "bg-orange-500/10", text: "text-orange-600 dark:text-orange-400", label: "Высокий" },
  medium: { dot: "bg-yellow-500", bg: "bg-yellow-500/10", text: "text-yellow-700 dark:text-yellow-400", label: "Средний" },
  low: { dot: "bg-blue-500", bg: "bg-blue-500/10", text: "text-blue-600 dark:text-blue-400", label: "Низкий" },
  trivial: { dot: "bg-gray-400", bg: "bg-gray-500/10", text: "text-gray-500", label: "Тривиальный" },
};

/** Конфигурация статусов. */
const STATUS_CONFIG: Record<string, { bg: string; text: string; label: string }> = {
  draft: { bg: "bg-gray-500/10", text: "text-gray-600 dark:text-gray-400", label: "Черновик" },
  active: { bg: "bg-green-500/10", text: "text-green-700 dark:text-green-400", label: "Активный" },
  deprecated: { bg: "bg-red-500/10", text: "text-red-600 dark:text-red-400", label: "Устаревший" },
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

/** Цвета ключевых слов BDD. */
const BDD_KEYWORD_COLORS: Record<string, string> = {
  Given: "text-blue-600 dark:text-blue-400 bg-blue-500/10",
  When: "text-amber-600 dark:text-amber-400 bg-amber-500/10",
  Then: "text-green-600 dark:text-green-400 bg-green-500/10",
  And: "text-gray-600 dark:text-gray-400 bg-gray-500/10",
  But: "text-red-600 dark:text-red-400 bg-red-500/10",
};

/**
 * Форматирует относительное время на русском языке.
 */
function formatRelativeTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMin < 1) return "только что";
  if (diffMin < 60) return `${diffMin} мин. назад`;
  if (diffHours < 24) return `${diffHours} ч. назад`;
  if (diffDays === 1) return "вчера";
  if (diffDays < 7) return `${diffDays} дн. назад`;
  return date.toLocaleDateString("ru-RU");
}

/**
 * Страница деталей тест-кейса — дизайн уровня Allure TestOps.
 * Заголовок с ключом и статусами, табы (Шаги / Связанные задачи / Автотест),
 * правый сайдбар с метаданными и историей версий.
 */
function TestCaseDetailPage() {
  const { projectId, caseKey } = Route.useParams();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<"steps" | "links" | "autotest">("steps");
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

  // Откат версии
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

  /** Начать редактирование. */
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
  const cancelEditing = () => setEditing(false);

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
        i === index ? { ...s, [field]: value } : s,
      ),
    );
  };

  if (isLoading || !tc) {
    return (
      <div className="mx-auto max-w-6xl animate-fade-in">
        <div className="mb-4 flex items-center gap-2">
          <div className="h-4 w-20 rounded animate-shimmer" />
          <div className="h-4 w-4 rounded animate-shimmer" />
          <div className="h-4 w-24 rounded animate-shimmer" />
        </div>
        <div className="flex gap-6">
          <div className="flex-1 space-y-4">
            <div className="h-8 w-48 rounded animate-shimmer" />
            <div className="h-6 w-96 rounded animate-shimmer" />
            <div className="h-64 w-full rounded-xl animate-shimmer" />
          </div>
          <div className="w-72 shrink-0">
            <div className="h-80 w-full rounded-xl animate-shimmer" />
          </div>
        </div>
      </div>
    );
  }

  const priorityCfg = (PRIORITY_CONFIG[tc.priority] ?? PRIORITY_CONFIG.medium)!;
  const statusCfg = (STATUS_CONFIG[tc.status] ?? STATUS_CONFIG.draft)!;
  const hasAutotest = !!(tc.full_name || tc.class_name || tc.external_id);

  /** Конфигурация табов детальной страницы. */
  const detailTabs = [
    { id: "steps" as const, label: "Шаги", count: tc.steps?.length ?? 0 },
    { id: "links" as const, label: "Связанные задачи", count: tc.linked_issues?.length ?? 0 },
    { id: "autotest" as const, label: "Автотест", show: hasAutotest || tc.case_type === "bdd" },
  ];

  return (
    <div className="mx-auto max-w-6xl animate-fade-in">
      {/* Навигация (breadcrumb) */}
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-1.5 text-sm">
          <Link
            to="/projects/$projectId/tms/cases"
            params={{ projectId }}
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Тест-кейсы
          </Link>
          <svg className="h-3.5 w-3.5 text-muted-foreground/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
          </svg>
          <span className="font-mono font-medium text-primary">{tc.case_key}</span>
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
                className="gradient-tms text-white"
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
        <div className="flex-1 min-w-0 space-y-5">
          {/* Заголовок кейса */}
          <div>
            <div className="mb-2 flex items-center gap-2.5 flex-wrap">
              <span className="font-mono text-lg font-bold text-primary">{tc.case_key}</span>
              {editing ? (
                <div className="flex items-center gap-2">
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
                </div>
              ) : (
                <>
                  <Badge variant="secondary" className={cn("text-[10px] font-semibold px-2", statusCfg.bg, statusCfg.text)}>
                    {statusCfg.label}
                  </Badge>
                  <div className="flex items-center gap-1">
                    <span className={cn("h-2 w-2 rounded-full", priorityCfg.dot)} />
                    <span className={cn("text-xs font-medium", priorityCfg.text)}>{priorityCfg.label}</span>
                  </div>
                  <Badge variant="outline" className="text-[10px] font-medium">
                    {tc.case_type === "bdd" ? "BDD" : "Classic"}
                  </Badge>
                </>
              )}
            </div>
            {editing ? (
              <Input
                className="text-xl font-semibold"
                value={editTitle}
                onChange={(e) => setEditTitle(e.target.value)}
                placeholder="Название тест-кейса"
              />
            ) : (
              <h2 className="text-xl font-semibold leading-tight">{tc.title}</h2>
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
                <p className="mt-2 text-sm text-muted-foreground leading-relaxed">{tc.description}</p>
              )
            )}
          </div>

          {/* Предусловия (только в режиме редактирования или если заполнены) */}
          {editing ? (
            <div className="rounded-lg border bg-card/50 p-3">
              <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Предусловие</div>
              <textarea
                className="w-full rounded-md border bg-card px-3 py-2 text-sm"
                placeholder="Предусловие (необязательно)"
                rows={2}
                value={editPrecondition}
                onChange={(e) => setEditPrecondition(e.target.value)}
              />
            </div>
          ) : (
            tc.precondition && (
              <div className="rounded-lg border border-blue-500/20 bg-blue-500/5 p-3">
                <div className="text-[10px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-wider mb-1">Предусловие</div>
                <p className="text-sm">{tc.precondition}</p>
              </div>
            )
          )}

          {/* Табы */}
          <div className="border-b">
            <div className="flex gap-0.5">
              {detailTabs.map((tab) => {
                if (tab.show === false) return null;
                return (
                  <button
                    key={tab.id}
                    className={cn(
                      "relative px-4 py-2 text-sm font-medium transition-all duration-200 border-b-2",
                      activeTab === tab.id
                        ? "border-primary text-primary"
                        : "border-transparent text-muted-foreground hover:text-foreground hover:border-muted-foreground/30",
                    )}
                    onClick={() => setActiveTab(tab.id)}
                  >
                    <span className="flex items-center gap-1.5">
                      {tab.label}
                      {"count" in tab && (tab.count ?? 0) > 0 && (
                        <span className={cn(
                          "text-[10px] font-semibold px-1.5 py-0 rounded-full",
                          activeTab === tab.id
                            ? "bg-primary/10 text-primary"
                            : "bg-muted text-muted-foreground",
                        )}>
                          {tab.count}
                        </span>
                      )}
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Таб: Шаги */}
          {activeTab === "steps" && (
            <div className="space-y-4">
              {editing ? (
                <div>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-semibold">Шаги теста</span>
                    <Button variant="outline" size="sm" onClick={addStep} className="h-7 text-xs">
                      <svg className="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
                      </svg>
                      Добавить шаг
                    </Button>
                  </div>
                  {editSteps.length === 0 ? (
                    <div className="text-center py-8 text-sm text-muted-foreground border rounded-lg border-dashed">
                      Нет шагов. Нажмите "Добавить шаг".
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {editSteps.map((step, idx) => (
                        <div key={idx} className="flex gap-2 items-start rounded-lg border p-3 bg-card/50 hover:bg-card transition-colors">
                          <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary mt-1">
                            {step.order}
                          </span>
                          <div className="flex-1 space-y-2">
                            <Input
                              placeholder="Действие"
                              value={step.action}
                              onChange={(e) => updateStep(idx, "action", e.target.value)}
                              className="text-sm"
                            />
                            <Input
                              placeholder="Ожидаемый результат"
                              value={step.expected_result ?? ""}
                              onChange={(e) => updateStep(idx, "expected_result", e.target.value)}
                              className="text-sm"
                            />
                          </div>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="mt-1 h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                            onClick={() => removeStep(idx)}
                          >
                            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ) : (
                <>
                  {tc.steps && tc.steps.length > 0 ? (
                    <div className="rounded-xl border overflow-hidden shadow-card">
                      <table className="w-full">
                        <thead>
                          <tr className="border-b bg-muted/30">
                            <th className="w-12 px-3 py-2.5 text-center text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                              #
                            </th>
                            {tc.case_type === "bdd" && (
                              <th className="w-24 px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                                Keyword
                              </th>
                            )}
                            <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                              Действие
                            </th>
                            <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                              Ожидаемый результат
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {tc.steps.map((step) => (
                            <tr key={step.order} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                              <td className="px-3 py-2.5 text-center">
                                <span className="inline-flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary">
                                  {step.order}
                                </span>
                              </td>
                              {tc.case_type === "bdd" && (
                                <td className="px-3 py-2.5">
                                  {step.keyword && (
                                    <Badge
                                      variant="secondary"
                                      className={cn(
                                        "text-[10px] font-bold font-mono",
                                        BDD_KEYWORD_COLORS[step.keyword] ?? "text-gray-600 bg-gray-500/10",
                                      )}
                                    >
                                      {step.keyword}
                                    </Badge>
                                  )}
                                </td>
                              )}
                              <td className="px-3 py-2.5 text-sm">{step.action}</td>
                              <td className="px-3 py-2.5 text-sm text-muted-foreground">
                                {step.expected_result || (
                                  <span className="italic text-muted-foreground/40">--</span>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ) : (
                    <div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
                      <svg className="h-10 w-10 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                      <p className="text-sm text-muted-foreground mb-2">Шаги не заданы</p>
                      <Button variant="outline" size="sm" onClick={startEditing}>
                        Добавить шаги
                      </Button>
                    </div>
                  )}
                </>
              )}

              {/* Постусловие */}
              {editing ? (
                <div className="rounded-lg border bg-card/50 p-3">
                  <div className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Постусловие</div>
                  <textarea
                    className="w-full rounded-md border bg-card px-3 py-2 text-sm"
                    placeholder="Постусловие (необязательно)"
                    rows={2}
                    value={editPostcondition}
                    onChange={(e) => setEditPostcondition(e.target.value)}
                  />
                </div>
              ) : (
                tc.postcondition && (
                  <div className="rounded-lg border border-amber-500/20 bg-amber-500/5 p-3">
                    <div className="text-[10px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-wider mb-1">Постусловие</div>
                    <p className="text-sm">{tc.postcondition}</p>
                  </div>
                )
              )}
            </div>
          )}

          {/* Таб: Связанные задачи */}
          {activeTab === "links" && (
            <div className="space-y-3">
              {tc.linked_issues && tc.linked_issues.length > 0 ? (
                <div className="rounded-xl border overflow-hidden shadow-card">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b bg-muted/30">
                        <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Ключ</th>
                        <th className="px-3 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Название</th>
                        <th className="w-10 px-3 py-2.5" />
                      </tr>
                    </thead>
                    <tbody>
                      {tc.linked_issues.map((issue) => (
                        <tr key={issue.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                          <td className="px-3 py-2.5">
                            <Link
                              to="/projects/$projectId/issues/$issueKey"
                              params={{ projectId, issueKey: issue.issue_key }}
                              className="font-mono text-xs font-bold text-primary hover:underline"
                            >
                              {issue.issue_key}
                            </Link>
                          </td>
                          <td className="px-3 py-2.5 text-sm">{issue.title}</td>
                          <td className="px-3 py-2.5">
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                              disabled={unlinkMutation.isPending}
                              onClick={() => unlinkMutation.mutate(issue.id)}
                            >
                              <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                              </svg>
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
                  <svg className="h-10 w-10 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                  <p className="text-sm text-muted-foreground">Нет связанных задач</p>
                </div>
              )}

              <div className="flex gap-2">
                <Input
                  placeholder="ID или ключ задачи"
                  className="flex-1 h-8 text-sm"
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
                  className="h-8"
                  disabled={!linkIssueId.trim() || linkMutation.isPending}
                  onClick={() => linkMutation.mutate(linkIssueId.trim())}
                >
                  <svg className="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                  </svg>
                  Привязать
                </Button>
              </div>
            </div>
          )}

          {/* Таб: Автотест */}
          {activeTab === "autotest" && (
            <div className="space-y-3">
              {hasAutotest ? (
                <div className="rounded-xl border overflow-hidden shadow-card">
                  <div className="p-4 space-y-3">
                    {tc.full_name && (
                      <div>
                        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Полное имя</div>
                        <code className="text-sm font-mono bg-muted/50 px-2 py-1 rounded">{tc.full_name}</code>
                      </div>
                    )}
                    {tc.class_name && (
                      <div>
                        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Класс</div>
                        <code className="text-sm font-mono bg-muted/50 px-2 py-1 rounded">{tc.class_name}</code>
                      </div>
                    )}
                    {tc.method_name && (
                      <div>
                        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">Метод</div>
                        <code className="text-sm font-mono bg-muted/50 px-2 py-1 rounded">{tc.method_name}</code>
                      </div>
                    )}
                    {tc.external_id && (
                      <div>
                        <div className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">External ID</div>
                        <code className="text-sm font-mono bg-muted/50 px-2 py-1 rounded">{tc.external_id}</code>
                      </div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
                  <svg className="h-10 w-10 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  <p className="text-sm text-muted-foreground">Автотест не привязан</p>
                  <p className="text-xs text-muted-foreground/60 mt-1">Свяжите автотест через импорт результатов</p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Правый сайдбар — метаданные */}
        <div className="w-72 shrink-0 space-y-4">
          {/* Метаданные */}
          <Card className="shadow-card">
            <CardHeader className="pb-2 px-4 pt-4">
              <CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Метаданные</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 px-4 pb-4">
              <div className="grid grid-cols-2 gap-x-3 gap-y-2.5">
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Приоритет</div>
                  <div className="flex items-center gap-1 mt-0.5">
                    <span className={cn("h-2 w-2 rounded-full", priorityCfg.dot)} />
                    <span className="text-xs font-medium">{priorityCfg.label}</span>
                  </div>
                </div>
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Статус</div>
                  <Badge variant="secondary" className={cn("text-[10px] font-medium mt-0.5", statusCfg.bg, statusCfg.text)}>
                    {statusCfg.label}
                  </Badge>
                </div>
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Тип</div>
                  <span className="text-xs font-medium mt-0.5">{tc.case_type === "bdd" ? "BDD" : "Classic"}</span>
                </div>
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Версия</div>
                  <span className="text-xs font-medium mt-0.5 tabular-nums">v{tc.version}</span>
                </div>
              </div>

              <div className="border-t pt-2.5 space-y-2.5">
                {tc.created_by && (
                  <div>
                    <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Автор</div>
                    <div className="text-xs font-medium mt-0.5">{tc.created_by.name}</div>
                  </div>
                )}
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Создан</div>
                  <div className="text-xs text-muted-foreground mt-0.5">{formatRelativeTime(tc.created_at)}</div>
                </div>
                <div>
                  <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Обновлён</div>
                  <div className="text-xs text-muted-foreground mt-0.5">{formatRelativeTime(tc.updated_at)}</div>
                </div>
                {tc.suite && (
                  <div>
                    <div className="text-[10px] text-muted-foreground/70 uppercase tracking-wider font-medium">Папка</div>
                    <div className="text-xs font-medium mt-0.5 flex items-center gap-1">
                      <svg className="h-3 w-3 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                      </svg>
                      {tc.suite.name}
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Компоненты */}
          {tc.components && tc.components.length > 0 && (
            <Card className="shadow-card">
              <CardHeader className="pb-2 px-4 pt-4">
                <CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Компоненты</CardTitle>
              </CardHeader>
              <CardContent className="px-4 pb-4">
                <div className="flex flex-wrap gap-1.5">
                  {tc.components.map((c) => (
                    <Badge key={c.id} variant="outline" className="text-[10px] font-medium">
                      {c.name}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Метки */}
          {tc.labels && tc.labels.length > 0 && (
            <Card className="shadow-card">
              <CardHeader className="pb-2 px-4 pt-4">
                <CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">Метки</CardTitle>
              </CardHeader>
              <CardContent className="px-4 pb-4">
                <div className="flex flex-wrap gap-1.5">
                  {tc.labels.map((l) => (
                    <Badge
                      key={l.id}
                      variant="secondary"
                      className="text-[10px] font-medium"
                      style={{ backgroundColor: l.color + "20", color: l.color }}
                    >
                      {l.name}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* История версий */}
          <Card className="shadow-card">
            <CardHeader className="pb-2 px-4 pt-4">
              <div className="flex items-center justify-between">
                <CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
                  История версий
                </CardTitle>
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 text-[11px] px-2"
                  onClick={() => setVersionsOpen(!versionsOpen)}
                >
                  {versionsOpen ? "Скрыть" : "Показать"}
                </Button>
              </div>
            </CardHeader>
            {versionsOpen && versions && (
              <CardContent className="px-4 pb-4">
                {versions.length === 0 ? (
                  <p className="text-xs text-muted-foreground">
                    Нет предыдущих версий
                  </p>
                ) : (
                  <div className="space-y-2">
                    {versions.map((v: TestCaseVersionResponse) => (
                      <div
                        key={v.id}
                        className="rounded-lg border p-2.5 hover:bg-muted/30 transition-colors"
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-xs font-semibold tabular-nums">
                            v{v.version_number}
                          </span>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-5 text-[10px] px-1.5"
                            disabled={rollbackMutation.isPending}
                            onClick={() => rollbackMutation.mutate(v.id)}
                          >
                            Откатить
                          </Button>
                        </div>
                        {v.change_description && (
                          <p className="text-[11px] text-muted-foreground mt-0.5">{v.change_description}</p>
                        )}
                        <div className="text-[10px] text-muted-foreground/60 mt-1">
                          {v.created_by?.name} -- {formatRelativeTime(v.created_at)}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            )}
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
