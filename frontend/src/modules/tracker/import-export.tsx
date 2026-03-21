import { useState, useRef } from "react";

/** Диалог импорта задач из CSV/JSON. */
export function ImportExportToolbar({ projectId }: { projectId: string }) {
  const [showImport, setShowImport] = useState(false);
  const [preview, setPreview] = useState<{
    headers: string[];
    rows: { fields: Record<string, string> }[];
    total: number;
  } | null>(null);
  const [importing, setImporting] = useState(false);
  const [result, setResult] = useState<{
    imported: number;
    skipped: number;
  } | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  const handleFilePreview = async (file: File) => {
    const formData = new FormData();
    formData.append("file", file);

    const resp = await fetch(
      `/api/projects/${projectId}/import/issues?mode=preview`,
      {
        method: "POST",
        body: formData,
      },
    );
    if (resp.ok) {
      setPreview(await resp.json());
    }
  };

  const handleImport = async () => {
    if (!fileRef.current?.files?.[0]) return;
    setImporting(true);
    try {
      const formData = new FormData();
      formData.append("file", fileRef.current.files[0]);

      const resp = await fetch(
        `/api/projects/${projectId}/import/issues?mode=commit`,
        {
          method: "POST",
          body: formData,
        },
      );
      if (resp.ok) {
        const data = await resp.json();
        setResult(data);
        setPreview(null);
      }
    } finally {
      setImporting(false);
    }
  };

  const handleExport = async (format: "csv" | "excel") => {
    window.open(
      `/api/projects/${projectId}/export/issues?format=${format}`,
      "_blank",
    );
  };

  return (
    <div className="flex items-center gap-2">
      <button
        onClick={() => setShowImport(true)}
        className="rounded-md border px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
      >
        Импорт
      </button>
      <div className="relative">
        <button
          onClick={() => handleExport("csv")}
          className="rounded-md border px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
        >
          Экспорт CSV
        </button>
      </div>
      <button
        onClick={() => handleExport("excel")}
        className="rounded-md border px-3 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
      >
        Экспорт Excel
      </button>

      {showImport && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-2xl rounded-lg bg-white p-6 dark:bg-gray-900">
            <h3 className="mb-4 text-lg font-semibold">Импорт задач</h3>

            <input
              ref={fileRef}
              type="file"
              accept=".csv,.json"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) handleFilePreview(file);
              }}
              className="mb-4"
            />

            {preview && (
              <div className="mb-4">
                <p className="mb-2 text-sm text-gray-500">
                  Найдено {preview.total} строк. Предпросмотр:
                </p>
                <div className="max-h-48 overflow-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        {preview.headers.map((h) => (
                          <th key={h} className="p-1 text-left">
                            {h}
                          </th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {preview.rows.map((row, i) => (
                        <tr key={i} className="border-b">
                          {preview.headers.map((h) => (
                            <td key={h} className="p-1">
                              {row.fields[h] ?? ""}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {result && (
              <div className="mb-4 rounded-md bg-green-50 p-3 text-sm dark:bg-green-950">
                Импортировано: {result.imported}, пропущено: {result.skipped}
              </div>
            )}

            <div className="flex justify-end gap-2">
              {preview && (
                <button
                  onClick={handleImport}
                  disabled={importing}
                  className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white disabled:opacity-50"
                >
                  {importing ? "Импорт..." : "Импортировать"}
                </button>
              )}
              <button
                onClick={() => {
                  setShowImport(false);
                  setPreview(null);
                  setResult(null);
                }}
                className="rounded-md border px-4 py-2 text-sm"
              >
                Закрыть
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
