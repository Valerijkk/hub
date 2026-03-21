import { useState, useEffect } from "react";
import { getLimits, updateLimit, type ResourceLimit } from "@/api/admin";

/** Описания лимитов на русском. */
const LIMIT_LABELS: Record<string, string> = {
  "limit.file.max_size_mb": "Максимальный размер файла (МБ)",
  "limit.user.max_projects": "Максимум проектов на пользователя",
  "limit.project.max_storage_mb": "Максимум хранилища на проект (МБ)",
};

/**
 * ResourceLimits -- страница администратора для настройки лимитов ресурсов.
 * Отображает форму с числовыми полями и кнопкой сохранения.
 */
export function ResourceLimits() {
  const [limits, setLimits] = useState<ResourceLimit[]>([]);
  const [values, setValues] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [toast, setToast] = useState("");

  useEffect(() => {
    setLoading(true);
    getLimits()
      .then((data) => {
        setLimits(data);
        const vals: Record<string, string> = {};
        for (const l of data) {
          vals[l.key] = l.value;
        }
        setValues(vals);
      })
      .catch((err) => console.error("Ошибка загрузки лимитов:", err))
      .finally(() => setLoading(false));
  }, []);

  /** Сохранить все изменённые лимиты. */
  const handleSave = async () => {
    setSaving(true);
    try {
      for (const limit of limits) {
        if (values[limit.key] !== limit.value) {
          await updateLimit(limit.key, values[limit.key] ?? limit.value);
        }
      }
      setToast("Лимиты сохранены");
      setTimeout(() => setToast(""), 3000);
      // Обновляем
      const data = await getLimits();
      setLimits(data);
    } catch (err) {
      console.error("Ошибка сохранения лимитов:", err);
      setToast("Ошибка сохранения");
      setTimeout(() => setToast(""), 3000);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return <div className="py-8 text-center text-muted-foreground">Загрузка...</div>;
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold">Лимиты ресурсов</h2>

      <div className="space-y-4">
        {limits.map((limit) => (
          <div key={limit.key} className="flex items-center gap-4">
            <label className="min-w-64 text-sm font-medium">
              {LIMIT_LABELS[limit.key] ?? limit.key}
            </label>
            <input
              type="number"
              min={0}
              value={values[limit.key] ?? ""}
              onChange={(e) =>
                setValues((v) => ({ ...v, [limit.key]: e.target.value }))
              }
              className="w-32 rounded border px-3 py-1.5 text-sm"
            />
            <span className="text-xs text-muted-foreground">
              Текущее: {limit.value}
            </span>
          </div>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <button
          onClick={handleSave}
          disabled={saving}
          className="rounded bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {saving ? "Сохранение..." : "Сохранить"}
        </button>
        {toast && (
          <span className="text-sm text-green-600">{toast}</span>
        )}
      </div>
    </div>
  );
}
