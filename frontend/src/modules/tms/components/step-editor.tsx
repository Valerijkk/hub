import type { TestStep, TestCaseResponse } from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

/** Цвета BDD-ключевых слов для бейджей (Given, When, Then, And, But). */
const BDD_KEYWORD_COLORS: Record<string, string> = {
  Given: "text-blue-600 dark:text-blue-400 bg-blue-500/10",
  When: "text-amber-600 dark:text-amber-400 bg-amber-500/10",
  Then: "text-green-600 dark:text-green-400 bg-green-500/10",
  And: "text-gray-600 dark:text-gray-400 bg-gray-500/10",
  But: "text-red-600 dark:text-red-400 bg-red-500/10",
};

/** Пропсы редактора шагов в режиме редактирования. */
interface StepEditorEditProps {
  editing: true;
  editSteps: TestStep[];
  onAddStep: () => void;
  onRemoveStep: (index: number) => void;
  onUpdateStep: (index: number, field: "action" | "expected_result", value: string) => void;
}

/** Пропсы отображения шагов в режиме чтения. */
interface StepEditorReadProps {
  editing: false;
  tc: TestCaseResponse;
  onStartEditing: () => void;
}

export type StepEditorProps = StepEditorEditProps | StepEditorReadProps;

/** Редактор и отображение шагов тест-кейса (режим редактирования и чтения). */
export function StepEditor(props: StepEditorProps) {
  if (props.editing) {
    return (
      <div>
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm font-semibold">Шаги теста</span>
          <Button variant="outline" size="sm" onClick={props.onAddStep} className="h-7 text-xs">
            <svg className="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            Добавить шаг
          </Button>
        </div>
        {props.editSteps.length === 0 ? (
          <div className="text-center py-8 text-sm text-muted-foreground border rounded-lg border-dashed">
            Нет шагов. Нажмите "Добавить шаг".
          </div>
        ) : (
          <div className="space-y-2">
            {props.editSteps.map((step, idx) => (
              <div key={idx} className="flex gap-2 items-start rounded-lg border p-3 bg-card/50 hover:bg-card transition-colors">
                <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary mt-1">
                  {step.order}
                </span>
                <div className="flex-1 space-y-2">
                  <Input
                    placeholder="Действие"
                    value={step.action}
                    onChange={(e) => props.onUpdateStep(idx, "action", e.target.value)}
                    className="text-sm"
                  />
                  <Input
                    placeholder="Ожидаемый результат"
                    value={step.expected_result ?? ""}
                    onChange={(e) => props.onUpdateStep(idx, "expected_result", e.target.value)}
                    className="text-sm"
                  />
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  className="mt-1 h-6 w-6 p-0 text-muted-foreground hover:text-destructive"
                  onClick={() => props.onRemoveStep(idx)}
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
    );
  }

  // Read mode
  const { tc, onStartEditing } = props;

  if (tc.steps && tc.steps.length > 0) {
    return (
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
    );
  }

  return (
    <div className="flex flex-col items-center justify-center py-12 text-center border rounded-xl border-dashed">
      <svg className="h-10 w-10 text-muted-foreground/20 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <p className="text-sm text-muted-foreground mb-2">Шаги не заданы</p>
      <Button variant="outline" size="sm" onClick={onStartEditing}>
        Добавить шаги
      </Button>
    </div>
  );
}
