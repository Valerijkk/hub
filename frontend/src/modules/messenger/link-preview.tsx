import { useState, useEffect } from "react";
import { getLinkPreview, type LinkPreviewResponse } from "@/api/github";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

interface LinkPreviewProps {
  /** URL для получения превью. */
  url: string;
}

// ---------------------------------------------------------------------------
// Иконки для типов задач
// ---------------------------------------------------------------------------

/** Иконка типа задачи. */
function IssueTypeIcon({ type }: { type?: string }) {
  const colors: Record<string, string> = {
    bug: "text-red-500",
    task: "text-blue-500",
    story: "text-green-500",
    epic: "text-purple-500",
    improvement: "text-orange-500",
    tech_debt: "text-yellow-500",
    spike: "text-cyan-500",
    sub_task: "text-gray-500",
    release: "text-indigo-500",
  };

  const color = colors[type ?? ""] ?? "text-gray-500";

  return (
    <svg
      className={`h-4 w-4 shrink-0 ${color}`}
      viewBox="0 0 16 16"
      fill="currentColor"
    >
      <path d="M8 9.5a1.5 1.5 0 1 0 0-3 1.5 1.5 0 0 0 0 3Z" />
      <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Z" />
    </svg>
  );
}

/** Иконка тест-кейса. */
function TestCaseIcon() {
  return (
    <svg
      className="h-4 w-4 shrink-0 text-teal-500"
      viewBox="0 0 16 16"
      fill="currentColor"
    >
      <path d="M8 16A8 8 0 1 0 8 0a8 8 0 0 0 0 16Zm3.78-9.72a.751.751 0 0 0-1.042-.018.751.751 0 0 0-.018 1.042L7.28 10.72a.75.75 0 0 1-1.06 0L4.22 8.72a.751.751 0 0 1 1.042-1.042l1.47 1.47 2.53-2.87Z" />
    </svg>
  );
}

/** Иконка внешней ссылки. */
function ExternalLinkIcon() {
  return (
    <svg
      className="h-4 w-4 shrink-0 text-gray-400"
      viewBox="0 0 16 16"
      fill="currentColor"
    >
      <path d="M3.75 2h3.5a.75.75 0 0 1 0 1.5h-3.5a.25.25 0 0 0-.25.25v8.5c0 .138.112.25.25.25h8.5a.25.25 0 0 0 .25-.25v-3.5a.75.75 0 0 1 1.5 0v3.5A1.75 1.75 0 0 1 12.25 14h-8.5A1.75 1.75 0 0 1 2 12.25v-8.5C2 2.784 2.784 2 3.75 2Zm6.854-1h4.146a.25.25 0 0 1 .25.25v4.146a.25.25 0 0 1-.427.177L13.03 4.03 9.28 7.78a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042l3.75-3.75-1.543-1.543A.25.25 0 0 1 10.604 1Z" />
    </svg>
  );
}

// ---------------------------------------------------------------------------
// Бейдж статуса
// ---------------------------------------------------------------------------

/** Бейдж статуса (для задач/тест-кейсов). */
function StatusBadge({ status }: { status?: string }) {
  if (!status) return null;
  return (
    <span className="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 dark:bg-gray-700 dark:text-gray-300">
      {status}
    </span>
  );
}

// ---------------------------------------------------------------------------
// Превью для различных типов ссылок
// ---------------------------------------------------------------------------

/** Превью внутренней задачи. */
function IssuePreview({ data }: { data: LinkPreviewResponse }) {
  return (
    <div className="flex items-start gap-2">
      <IssueTypeIcon type={data.issue_type} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-gray-500">
            {data.issue_key}
          </span>
          <StatusBadge status={data.status} />
        </div>
        <p className="truncate text-sm font-medium">{data.title}</p>
        {data.assignee_name && (
          <div className="mt-1 flex items-center gap-1">
            {data.avatar_url ? (
              <img
                src={data.avatar_url}
                alt={data.assignee_name}
                className="h-4 w-4 rounded-full"
              />
            ) : (
              <div className="flex h-4 w-4 items-center justify-center rounded-full bg-gray-300 text-[8px] font-medium text-gray-600">
                {data.assignee_name[0]?.toUpperCase()}
              </div>
            )}
            <span className="text-xs text-gray-500">{data.assignee_name}</span>
          </div>
        )}
      </div>
    </div>
  );
}

/** Превью тест-кейса. */
function TestCasePreview({ data }: { data: LinkPreviewResponse }) {
  return (
    <div className="flex items-start gap-2">
      <TestCaseIcon />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{data.title}</p>
        {data.status && <StatusBadge status={data.status} />}
      </div>
    </div>
  );
}

/** Превью внешнего URL (OpenGraph). */
function ExternalPreview({ data }: { data: LinkPreviewResponse }) {
  return (
    <div className="flex gap-3">
      {data.image_url && (
        <img
          src={data.image_url}
          alt=""
          className="h-16 w-24 shrink-0 rounded object-cover"
          onError={(e) => {
            (e.target as HTMLImageElement).style.display = "none";
          }}
        />
      )}
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{data.title || data.url}</p>
        {data.description && (
          <p className="mt-0.5 line-clamp-2 text-xs text-gray-500 dark:text-gray-400">
            {data.description}
          </p>
        )}
        {data.domain && (
          <div className="mt-1 flex items-center gap-1">
            <ExternalLinkIcon />
            <span className="text-xs text-gray-400">{data.domain}</span>
          </div>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Основной компонент
// ---------------------------------------------------------------------------

/**
 * LinkPreview -- универсальный компонент превью ссылок (MSG-11).
 * Определяет тип ссылки и отображает соответствующую карточку:
 * - Внутренние задачи: тип + ключ + title + статус + assignee
 * - Тест-кейсы: иконка + title + статус
 * - Внешние URL: OG image + title + description + domain
 */
export function LinkPreview({ url }: LinkPreviewProps) {
  const [data, setData] = useState<LinkPreviewResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function fetchPreview() {
      try {
        const result = await getLinkPreview(url);
        if (!cancelled) {
          setData(result);
        }
      } catch {
        if (!cancelled) {
          setError(true);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    fetchPreview();
    return () => {
      cancelled = true;
    };
  }, [url]);

  if (loading) {
    return (
      <div className="mt-1 max-w-[400px] animate-pulse rounded-lg border border-gray-200 p-3 dark:border-gray-700">
        <div className="h-4 w-3/4 rounded bg-gray-200 dark:bg-gray-700" />
        <div className="mt-2 h-3 w-1/2 rounded bg-gray-200 dark:bg-gray-700" />
      </div>
    );
  }

  if (error || !data) return null;

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="mt-1 block max-w-[400px] rounded-lg border border-gray-200 p-3 transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-gray-800/50"
    >
      {data.type === "issue" && <IssuePreview data={data} />}
      {data.type === "test_case" && <TestCasePreview data={data} />}
      {(data.type === "external" || data.type === "github") && (
        <ExternalPreview data={data} />
      )}
    </a>
  );
}
