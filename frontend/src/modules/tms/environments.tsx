import { useState, useEffect, useCallback } from "react";
import { api } from "@/api/client";

/** Тестовое окружение. */
interface Environment {
  id: string;
  name: string;
  label: string;
  config?: Record<string, unknown>;
  created_at: string;
}

const LABELS = ["dev", "staging", "prod", "custom"] as const;

/** Страница управления тестовыми окружениями. */
export function EnvironmentsPage({ projectId }: { projectId: string }) {
  const [envs, setEnvs] = useState<Environment[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState("");
  const [label, setLabel] = useState<string>("custom");
  const [configBrowser, setConfigBrowser] = useState("");
  const [configOS, setConfigOS] = useState("");

  const fetchEnvs = useCallback(async () => {
    try {
      setLoading(true);
      const data = await api.get<Environment[]>(
        `/projects/${projectId}/environments`,
      );
      setEnvs(data ?? []);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchEnvs();
  }, [fetchEnvs]);

  const handleCreate = async () => {
    if (!name.trim()) return;
    const config: Record<string, string> = {};
    if (configBrowser) config.browser = configBrowser;
    if (configOS) config.os = configOS;

    await api.post(`/projects/${projectId}/environments`, {
      name: name.trim(),
      label,
      config: Object.keys(config).length > 0 ? config : undefined,
    });
    setName("");
    setLabel("custom");
    setConfigBrowser("");
    setConfigOS("");
    setShowCreate(false);
    fetchEnvs();
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Удалить это окружение?")) return;
    await api.delete(`/projects/${projectId}/environments/${id}`);
    fetchEnvs();
  };

  const labelBadge = (l: string) => {
    const colors: Record<string, string> = {
      dev: "bg-blue-100 text-blue-700",
      staging: "bg-yellow-100 text-yellow-700",
      prod: "bg-red-100 text-red-700",
      custom: "bg-gray-100 text-gray-700",
    };
    return (
      <span className={`rounded-full px-2 py-1 text-xs ${colors[l] ?? colors.custom}`}>
        {l}
      </span>
    );
  };

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Тестовые окружения</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
        >
          Создать окружение
        </button>
      </div>

      {showCreate && (
        <div className="rounded-md border p-4 dark:border-gray-700">
          <h3 className="mb-3 font-medium">Новое окружение</h3>
          <div className="space-y-3">
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Название (например, Chrome Linux)"
              className="w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
            />
            <select
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              className="w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
            >
              {LABELS.map((l) => (
                <option key={l} value={l}>
                  {l}
                </option>
              ))}
            </select>
            <input
              type="text"
              value={configBrowser}
              onChange={(e) => setConfigBrowser(e.target.value)}
              placeholder="Браузер (необязательно)"
              className="w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
            />
            <input
              type="text"
              value={configOS}
              onChange={(e) => setConfigOS(e.target.value)}
              placeholder="ОС (необязательно)"
              className="w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
            />
          </div>
          <div className="mt-3 flex gap-2">
            <button
              onClick={handleCreate}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
            >
              Создать
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="rounded-md border px-4 py-2 text-sm"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-gray-500">Загрузка...</p>
      ) : envs.length === 0 ? (
        <p className="text-gray-500">Нет окружений</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-left dark:border-gray-700">
              <th className="p-2">Название</th>
              <th className="p-2">Метка</th>
              <th className="p-2">Конфигурация</th>
              <th className="p-2">Действия</th>
            </tr>
          </thead>
          <tbody>
            {envs.map((env) => (
              <tr key={env.id} className="border-b dark:border-gray-700">
                <td className="p-2">{env.name}</td>
                <td className="p-2">{labelBadge(env.label)}</td>
                <td className="p-2 text-gray-500">
                  {env.config
                    ? Object.entries(env.config)
                        .map(([k, v]) => `${k}: ${v}`)
                        .join(", ")
                    : "-"}
                </td>
                <td className="p-2">
                  <button
                    onClick={() => handleDelete(env.id)}
                    className="text-red-600 hover:underline"
                  >
                    Удалить
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
