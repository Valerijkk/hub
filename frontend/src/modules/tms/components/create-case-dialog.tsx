/**
 * Диалог создания нового тест-кейса.
 */
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createTestCase,
  type TestCaseResponse,
  type TestCaseType,
  type SuiteResponse,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "sonner";

interface CreateCaseDialogProps {
  projectId: string;
  suites?: SuiteResponse[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  trigger: React.ReactNode;
}

export function CreateCaseDialog({
  projectId,
  suites,
  open,
  onOpenChange,
  trigger,
}: CreateCaseDialogProps) {
  const queryClient = useQueryClient();
  const [newTitle, setNewTitle] = useState("");
  const [newType, setNewType] = useState<TestCaseType>("classic");
  const [newPriority, setNewPriority] = useState("medium");
  const [newSuiteId, setNewSuiteId] = useState("");
  const [newDescription, setNewDescription] = useState("");

  const createMutation = useMutation({
    mutationFn: () =>
      createTestCase(projectId, {
        title: newTitle,
        case_type: newType,
        priority: newPriority as TestCaseResponse["priority"],
        suite_id: newSuiteId || undefined,
        description: newDescription || undefined,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["tms-cases", projectId] });
      onOpenChange(false);
      setNewTitle("");
      setNewDescription("");
      toast.success("Тест-кейс создан");
    },
    onError: () => toast.error("Ошибка создания тест-кейса"),
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>
        {trigger}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Новый тест-кейс</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <Input
            placeholder="Название тест-кейса"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
          />
          <div className="grid grid-cols-2 gap-3">
            <select
              className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
              value={newType}
              onChange={(e) => setNewType(e.target.value as TestCaseType)}
            >
              <option value="classic">Classic</option>
              <option value="bdd">BDD</option>
            </select>
            <select
              className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
              value={newPriority}
              onChange={(e) => setNewPriority(e.target.value)}
            >
              <option value="critical">Критический</option>
              <option value="high">Высокий</option>
              <option value="medium">Средний</option>
              <option value="low">Низкий</option>
              <option value="trivial">Тривиальный</option>
            </select>
          </div>
          {suites && suites.length > 0 && (
            <select
              className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
              value={newSuiteId}
              onChange={(e) => setNewSuiteId(e.target.value)}
            >
              <option value="">Без папки</option>
              {suites.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
          )}
          <textarea
            className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
            placeholder="Описание (необязательно)"
            rows={3}
            value={newDescription}
            onChange={(e) => setNewDescription(e.target.value)}
          />
          <Button
            className="w-full"
            disabled={!newTitle.trim() || createMutation.isPending}
            onClick={() => createMutation.mutate()}
          >
            {createMutation.isPending ? "Создание..." : "Создать"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
