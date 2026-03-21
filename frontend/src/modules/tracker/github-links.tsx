import { useState, useEffect, useCallback } from "react";
import {
  listGitHubLinks,
  listIntegrations,
  createBranch,
  type GitHubLinkResponse,
  type IntegrationResponse,
} from "@/api/github";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

interface GitHubLinksProps {
  /** Ключ задачи (например "HUB-123"). */
  issueKey: string;
  /** ID проекта для получения интеграций. */
  projectId: string;
}

// ---------------------------------------------------------------------------
// Компонент статуса PR
// ---------------------------------------------------------------------------

/** Бейдж статуса PR с цветовой индикацией. */
function PRStatusBadge({ status }: { status: string }) {
  const styles: Record<string, { bg: string; text: string; label: string }> = {
    open: { bg: "bg-green-100", text: "text-green-800", label: "Open" },
    merged: { bg: "bg-purple-100", text: "text-purple-800", label: "Merged" },
    closed: { bg: "bg-red-100", text: "text-red-800", label: "Closed" },
    none: { bg: "bg-gray-100", text: "text-gray-600", label: "" },
  };

  const defaultStyle = { bg: "bg-gray-100", text: "text-gray-600", label: "" };
  const style = styles[status] ?? defaultStyle;
  if (!style.label) return null;

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${style.bg} ${style.text}`}
    >
      {style.label}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Иконки
// ---------------------------------------------------------------------------

/** Иконка Git Pull Request. */
function GitPullRequestIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="currentColor"
    >
      <path d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Zm5.677-.177L9.573.677A.25.25 0 0 1 10 .854V2.5h1A2.5 2.5 0 0 1 13.5 5v5.628a2.251 2.251 0 1 1-1.5 0V5a1 1 0 0 0-1-1h-1v1.646a.25.25 0 0 1-.427.177L7.177 3.427a.25.25 0 0 1 0-.354ZM3.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm0 9.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm8.25.75a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0Z" />
    </svg>
  );
}

/** Иконка Git Commit. */
function GitCommitIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="currentColor"
    >
      <path d="M11.93 8.5a4.002 4.002 0 0 1-7.86 0H.75a.75.75 0 0 1 0-1.5h3.32a4.002 4.002 0 0 1 7.86 0h3.32a.75.75 0 0 1 0 1.5Zm-1.43-.5a2.5 2.5 0 1 0-5 0 2.5 2.5 0 0 0 5 0Z" />
    </svg>
  );
}

// ---------------------------------------------------------------------------
// Диалог создания ветки
// ---------------------------------------------------------------------------

interface CreateBranchDialogProps {
  issueKey: string;
  integrations: IntegrationResponse[];
  onClose: () => void;
  onCreated: (branchName: string) => void;
}

/** Диалог создания ветки из задачи (GHUB-05). */
function CreateBranchDialog({
  issueKey,
  integrations,
  onClose,
  onCreated,
}: CreateBranchDialogProps) {
  const defaultBranchName = issueKey.toLowerCase().replace(/[^a-z0-9-]/g, "-");
  const [branchName, setBranchName] = useState(defaultBranchName);
  const [selectedIntegration, setSelectedIntegration] = useState(
    integrations[0]?.id ?? "",
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleCreate = async () => {
    if (!selectedIntegration) return;
    setLoading(true);
    setError("");
    try {
      const result = await createBranch(issueKey, {
        integration_id: selectedIntegration,
        branch_name: branchName || undefined,
      });
      onCreated(result.branch_name);
    } catch (err: unknown) {
      const msg =
        err instanceof Error ? err.message : "Ошибка создания ветки";
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl dark:bg-gray-800">
        <h3 className="mb-4 text-lg font-semibold">Создать ветку</h3>

        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium">Репозиторий</label>
          <select
            className="w-full rounded-md border px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700"
            value={selectedIntegration}
            onChange={(e) => setSelectedIntegration(e.target.value)}
          >
            {integrations.map((ig) => (
              <option key={ig.id} value={ig.id}>
                {ig.repo_owner}/{ig.repo_name}
              </option>
            ))}
          </select>
        </div>

        <div className="mb-4">
          <label className="mb-1 block text-sm font-medium">Имя ветки</label>
          <input
            className="w-full rounded-md border px-3 py-2 text-sm dark:border-gray-600 dark:bg-gray-700"
            value={branchName}
            onChange={(e) => setBranchName(e.target.value)}
            placeholder={defaultBranchName}
          />
        </div>

        {error && (
          <p className="mb-4 text-sm text-red-600 dark:text-red-400">
            {error}
          </p>
        )}

        <div className="flex justify-end gap-2">
          <button
            className="rounded-md px-4 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
            onClick={onClose}
            disabled={loading}
          >
            Отмена
          </button>
          <button
            className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            onClick={handleCreate}
            disabled={loading || !selectedIntegration}
          >
            {loading ? "Создание..." : "Создать"}
          </button>
        </div>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Основной компонент
// ---------------------------------------------------------------------------

/**
 * GitHubLinks -- секция GitHub на странице задачи.
 * Отображает привязанные PR/коммиты, статус PR, кнопку создания ветки.
 */
export function GitHubLinks({ issueKey, projectId }: GitHubLinksProps) {
  const [links, setLinks] = useState<GitHubLinkResponse[]>([]);
  const [integrations, setIntegrations] = useState<IntegrationResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [showBranchDialog, setShowBranchDialog] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [linksData, integrationsData] = await Promise.all([
        listGitHubLinks(issueKey),
        listIntegrations(projectId),
      ]);
      setLinks(linksData);
      setIntegrations(integrationsData);
    } catch {
      // Ошибка загрузки -- показываем пустой список
    } finally {
      setLoading(false);
    }
  }, [issueKey, projectId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="animate-pulse space-y-2">
        <div className="h-4 w-24 rounded bg-gray-200 dark:bg-gray-700" />
        <div className="h-8 w-full rounded bg-gray-200 dark:bg-gray-700" />
      </div>
    );
  }

  const pullRequests = links.filter((l) => l.link_type === "pull_request");
  const commits = links.filter((l) => l.link_type === "commit");

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
          GitHub
        </h4>
        {integrations.length > 0 && (
          <button
            className="rounded-md px-2 py-1 text-xs text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20"
            onClick={() => setShowBranchDialog(true)}
          >
            + Создать ветку
          </button>
        )}
      </div>

      {links.length === 0 && (
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Нет привязанных PR или коммитов
        </p>
      )}

      {/* Pull Requests */}
      {pullRequests.length > 0 && (
        <div className="space-y-1">
          {pullRequests.map((link) => (
            <a
              key={link.id}
              href={link.url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <GitPullRequestIcon className="shrink-0 text-gray-500" />
              <span className="min-w-0 flex-1 truncate">{link.title}</span>
              <PRStatusBadge status={link.status} />
            </a>
          ))}
        </div>
      )}

      {/* Commits */}
      {commits.length > 0 && (
        <div className="space-y-1">
          {commits.map((link) => (
            <a
              key={link.id}
              href={link.url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
            >
              <GitCommitIcon className="shrink-0 text-gray-500" />
              <span className="min-w-0 flex-1 truncate" title={link.title}>
                {link.external_id.substring(0, 7)}
              </span>
              <span className="shrink-0 text-xs text-gray-400">
                {link.author}
              </span>
            </a>
          ))}
        </div>
      )}

      {/* Диалог создания ветки */}
      {showBranchDialog && (
        <CreateBranchDialog
          issueKey={issueKey}
          integrations={integrations}
          onClose={() => setShowBranchDialog(false)}
          onCreated={() => {
            setShowBranchDialog(false);
            fetchData();
          }}
        />
      )}
    </div>
  );
}
