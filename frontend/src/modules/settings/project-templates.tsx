import { useState, useEffect, useCallback } from "react";
import { api } from "@/api/client";

/** Шаблон проекта. */
interface ProjectTemplate {
  id: string;
  name: string;
  description?: string;
  template_type: string;
  config: Record<string, unknown>;
  is_system: boolean;
  created_at: string;
}

/** Компонент выбора шаблона при создании проекта. */
export function TemplateSelector({
  onSelect,
}: {
  onSelect: (templateId: string) => void;
}) {
  const [templates, setTemplates] = useState<ProjectTemplate[]>([]);
  const [selected, setSelected] = useState<string | null>(null);

  useEffect(() => {
    api.get<ProjectTemplate[]>("/project-templates").then((data) => {
      setTemplates(data ?? []);
    });
  }, []);

  const handleSelect = (id: string) => {
    setSelected(id);
    onSelect(id);
  };

  const typeIcon = (type: string) => {
    const icons: Record<string, string> = {
      empty: "[ ]",
      software_dev: "[DEV]",
      qa: "[QA]",
      kanban_simple: "[KAN]",
      custom: "[*]",
    };
    return icons[type] ?? "[?]";
  };

  return (
    <div className="space-y-2">
      <p className="text-sm font-medium">Выберите шаблон проекта:</p>
      <div className="grid grid-cols-2 gap-2">
        {templates.map((t) => (
          <button
            key={t.id}
            onClick={() => handleSelect(t.id)}
            className={`rounded-md border p-3 text-left text-sm ${
              selected === t.id
                ? "border-blue-500 bg-blue-50 dark:bg-blue-950"
                : "hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800"
            }`}
          >
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs text-gray-400">
                {typeIcon(t.template_type)}
              </span>
              <span className="font-medium">{t.name}</span>
              {t.is_system && (
                <span className="text-xs text-gray-400">Системный</span>
              )}
            </div>
            {t.description && (
              <p className="mt-1 text-xs text-gray-500">{t.description}</p>
            )}
          </button>
        ))}
      </div>
    </div>
  );
}

/** Страница управления шаблонами проектов (admin). */
export function ProjectTemplatesPage() {
  const [templates, setTemplates] = useState<ProjectTemplate[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchTemplates = useCallback(async () => {
    try {
      setLoading(true);
      const data = await api.get<ProjectTemplate[]>("/project-templates");
      setTemplates(data ?? []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTemplates();
  }, [fetchTemplates]);

  const handleDelete = async (id: string) => {
    if (!confirm("Удалить этот шаблон?")) return;
    await api.delete(`/project-templates/${id}`);
    fetchTemplates();
  };

  return (
    <div className="space-y-6 p-6">
      <h2 className="text-xl font-semibold">Шаблоны проектов</h2>

      {loading ? (
        <p className="text-gray-500">Загрузка...</p>
      ) : (
        <div className="space-y-3">
          {templates.map((t) => (
            <div
              key={t.id}
              className="flex items-center justify-between rounded-md border p-4 dark:border-gray-700"
            >
              <div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{t.name}</span>
                  <span className="rounded bg-gray-100 px-2 py-0.5 text-xs dark:bg-gray-800">
                    {t.template_type}
                  </span>
                  {t.is_system && (
                    <span className="text-xs text-gray-400">Системный</span>
                  )}
                </div>
                {t.description && (
                  <p className="mt-1 text-sm text-gray-500">{t.description}</p>
                )}
              </div>
              {!t.is_system && (
                <button
                  onClick={() => handleDelete(t.id)}
                  className="text-sm text-red-600 hover:underline"
                >
                  Удалить
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
