import { useQuery } from "@tanstack/react-query";
import { fetchLinkPreview, type LinkPreviewData } from "@/api/messenger";

/**
 * Пропсы компонента LinkPreview.
 * Отображает превью ссылки: внутренняя задача, GitHub или generic OG-preview.
 */
interface LinkPreviewProps {
  /** URL для отображения превью. */
  url: string;
}

/** Регулярное выражение для обнаружения внутренних ссылок на задачи. */
const INTERNAL_ISSUE_RE = /\/projects\/[^/]+\/issues\/([A-Z]+-\d+)/;

/** Регулярное выражение для обнаружения GitHub ссылок. */
const GITHUB_RE = /github\.com\/([^/]+)\/([^/]+)\/?(?:(?:issues|pull)\/(\d+))?/;

/**
 * Компонент превью внутренней задачи Hub.
 * Показывает ключ задачи в стиле бейджа со ссылкой.
 */
function InternalIssuePreview({ url, issueKey }: { url: string; issueKey: string }) {
  return (
    <a
      href={url}
      className="mt-2 flex items-center gap-2 rounded-lg border border-blue-500/20 bg-blue-500/5 p-3 transition-colors hover:bg-blue-500/10"
    >
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-500/10">
        <svg className="h-4 w-4 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      </div>
      <div>
        <span className="font-mono text-sm font-bold text-blue-600 dark:text-blue-400">{issueKey}</span>
        <p className="text-xs text-muted-foreground">Задача в Hub</p>
      </div>
    </a>
  );
}

/**
 * Компонент превью GitHub ссылки.
 * Отображает owner/repo и номер issue/PR если указан.
 */
function GitHubPreview({
  url,
  owner,
  repo,
  number,
}: {
  url: string;
  owner: string;
  repo: string;
  number?: string;
}) {
  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="mt-2 flex items-center gap-2 rounded-lg border border-gray-500/20 bg-gray-500/5 p-3 transition-colors hover:bg-gray-500/10"
    >
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gray-500/10">
        <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
        </svg>
      </div>
      <div>
        <span className="text-sm font-semibold">
          {owner}/{repo}
          {number && <span className="text-muted-foreground"> #{number}</span>}
        </span>
        <p className="text-xs text-muted-foreground">GitHub</p>
      </div>
    </a>
  );
}

/**
 * Компонент generic превью ссылки.
 * Загружает OG-данные через GET /api/link-preview?url=...
 * и отображает title, description, image.
 */
function GenericPreview({ url }: { url: string }) {
  const { data, isLoading, isError } = useQuery<LinkPreviewData>({
    queryKey: ["link-preview", url],
    queryFn: () => fetchLinkPreview(url),
    staleTime: 5 * 60 * 1000,
    retry: 1,
  });

  if (isLoading) {
    return (
      <div className="mt-2 h-16 animate-pulse rounded-lg border bg-muted/30" />
    );
  }

  if (isError || !data || (!data.title && !data.description)) {
    return null;
  }

  return (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="mt-2 flex overflow-hidden rounded-lg border transition-colors hover:bg-muted/30"
    >
      {data.image && (
        <div className="w-20 shrink-0 bg-muted">
          <img
            src={data.image}
            alt=""
            className="h-full w-full object-cover"
            loading="lazy"
          />
        </div>
      )}
      <div className="flex flex-col justify-center gap-0.5 p-3 min-w-0">
        {data.title && (
          <span className="text-sm font-medium truncate">{data.title}</span>
        )}
        {data.description && (
          <span className="text-xs text-muted-foreground line-clamp-2">
            {data.description}
          </span>
        )}
        <span className="text-[10px] text-muted-foreground/60 truncate mt-0.5">
          {(() => { try { return new URL(url).hostname; } catch { return url; } })()}
        </span>
      </div>
    </a>
  );
}

/**
 * LinkPreview — определяет тип ссылки и рендерит соответствующее превью.
 * Поддерживает: внутренние задачи Hub, GitHub, generic OG-preview.
 */
export function LinkPreview({ url }: LinkPreviewProps) {
  // Проверяем внутреннюю ссылку на задачу
  const internalMatch = url.match(INTERNAL_ISSUE_RE);
  if (internalMatch) {
    return <InternalIssuePreview url={url} issueKey={internalMatch[1]!} />;
  }

  // Проверяем GitHub ссылку
  const ghMatch = url.match(GITHUB_RE);
  if (ghMatch) {
    return (
      <GitHubPreview
        url={url}
        owner={ghMatch[1]!}
        repo={ghMatch[2]!}
        number={ghMatch[3]}
      />
    );
  }

  // Generic превью через OG-теги
  return <GenericPreview url={url} />;
}
