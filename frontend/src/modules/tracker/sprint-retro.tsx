import { useState } from "react";
import { useSprintStore } from "@/stores/sprint-store";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";

/** Пропсы компонента ретроспективы спринта. */
interface SprintRetroProps {
  /** ID проекта. */
  projectId: string;
  /** ID спринта. */
  sprintId: string;
  /** Начальные данные ретроспективы. */
  initialNotes?: {
    went_well?: string[];
    to_improve?: string[];
    actions?: string[];
  };
}

/**
 * SprintRetro предоставляет форму ретроспективы спринта (SCRM-03).
 * Три секции: "Что прошло хорошо", "Что можно улучшить", "Действия".
 * Сохраняет как JSONB через PUT /sprints/:id/retro.
 */
export function SprintRetro({
  projectId,
  sprintId,
  initialNotes,
}: SprintRetroProps) {
  const { saveRetroNotes } = useSprintStore();

  const [wentWell, setWentWell] = useState(
    initialNotes?.went_well?.join("\n") ?? "",
  );
  const [toImprove, setToImprove] = useState(
    initialNotes?.to_improve?.join("\n") ?? "",
  );
  const [actions, setActions] = useState(
    initialNotes?.actions?.join("\n") ?? "",
  );
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      await saveRetroNotes(projectId, sprintId, {
        went_well: wentWell
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
        to_improve: toImprove
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
        actions: actions
          .split("\n")
          .map((s) => s.trim())
          .filter(Boolean),
      });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ретроспектива</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium text-green-700">
            Что прошло хорошо
          </label>
          <Textarea
            value={wentWell}
            onChange={(e) => setWentWell(e.target.value)}
            placeholder="Каждый пункт с новой строки..."
            rows={4}
            className="border-green-200 focus:border-green-400"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-orange-700">
            Что можно улучшить
          </label>
          <Textarea
            value={toImprove}
            onChange={(e) => setToImprove(e.target.value)}
            placeholder="Каждый пункт с новой строки..."
            rows={4}
            className="border-orange-200 focus:border-orange-400"
          />
        </div>

        <div>
          <label className="mb-1 block text-sm font-medium text-blue-700">
            Действия
          </label>
          <Textarea
            value={actions}
            onChange={(e) => setActions(e.target.value)}
            placeholder="Каждый пункт с новой строки..."
            rows={4}
            className="border-blue-200 focus:border-blue-400"
          />
        </div>

        <div className="flex items-center gap-3">
          <Button onClick={handleSave} disabled={saving}>
            {saving ? "Сохранение..." : "Сохранить"}
          </Button>
          {saved && (
            <span className="text-sm text-green-600">Сохранено</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
