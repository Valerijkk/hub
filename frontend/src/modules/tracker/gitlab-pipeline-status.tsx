import { useQuery } from "@tanstack/react-query";
import { type GitLabPipelineResponse, listGitLabPipelines } from "@/api/gitlab";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

/** Пропсы компонента GitLabPipelineStatus. */
interface GitLabPipelineStatusProps {
	/** ID проекта для получения пайплайнов. */
	projectId: string;
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

/**
 * Возвращает конфигурацию бейджа статуса GitLab CI/CD пайплайна.
 */
function getStatusConfig(pipeline: GitLabPipelineResponse): {
	icon: "check" | "x" | "spinner" | "clock" | "minus";
	color: string;
	bgColor: string;
	label: string;
} {
	switch (pipeline.status) {
		case "success":
			return {
				icon: "check",
				color: "text-green-600 dark:text-green-400",
				bgColor: "bg-green-100 dark:bg-green-900/30",
				label: "Успешно",
			};
		case "failed":
			return {
				icon: "x",
				color: "text-red-600 dark:text-red-400",
				bgColor: "bg-red-100 dark:bg-red-900/30",
				label: "Ошибка",
			};
		case "running":
			return {
				icon: "spinner",
				color: "text-yellow-600 dark:text-yellow-400",
				bgColor: "bg-yellow-100 dark:bg-yellow-900/30",
				label: "Выполняется",
			};
		case "pending":
		case "waiting_for_resource":
		case "preparing":
		case "created":
			return {
				icon: "clock",
				color: "text-gray-500 dark:text-gray-400",
				bgColor: "bg-gray-100 dark:bg-gray-800",
				label: "Ожидание",
			};
		case "canceled":
			return {
				icon: "minus",
				color: "text-gray-500 dark:text-gray-400",
				bgColor: "bg-gray-100 dark:bg-gray-800",
				label: "Отменён",
			};
		case "skipped":
			return {
				icon: "minus",
				color: "text-gray-400 dark:text-gray-500",
				bgColor: "bg-gray-50 dark:bg-gray-800",
				label: "Пропущен",
			};
		case "manual":
			return {
				icon: "clock",
				color: "text-blue-500 dark:text-blue-400",
				bgColor: "bg-blue-100 dark:bg-blue-900/30",
				label: "Ручной",
			};
		case "scheduled":
			return {
				icon: "clock",
				color: "text-indigo-500 dark:text-indigo-400",
				bgColor: "bg-indigo-100 dark:bg-indigo-900/30",
				label: "Запланирован",
			};
		default:
			return {
				icon: "minus",
				color: "text-gray-500 dark:text-gray-400",
				bgColor: "bg-gray-100 dark:bg-gray-800",
				label: pipeline.status,
			};
	}
}

/** Форматирует относительное время на русском. */
function formatRelativeTime(dateStr: string): string {
	const now = Date.now();
	const date = new Date(dateStr).getTime();
	const diffMs = now - date;
	const diffMin = Math.floor(diffMs / 60000);

	if (diffMin < 1) return "только что";
	if (diffMin < 60) return `${diffMin} мин. назад`;
	const diffHours = Math.floor(diffMin / 60);
	if (diffHours < 24) return `${diffHours} ч. назад`;
	const diffDays = Math.floor(diffHours / 24);
	if (diffDays < 7) return `${diffDays} дн. назад`;
	return new Date(dateStr).toLocaleDateString("ru-RU", {
		day: "numeric",
		month: "short",
	});
}

// ---------------------------------------------------------------------------
// Иконки статусов
// ---------------------------------------------------------------------------

/** Иконка статуса CI/CD пайплайна. */
function StatusIcon({
	icon,
	className,
}: {
	icon: "check" | "x" | "spinner" | "clock" | "minus";
	className?: string;
}) {
	switch (icon) {
		case "check":
			return (
				<svg
					aria-hidden="true"
					className={className}
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="currentColor"
				>
					<path d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z" />
				</svg>
			);
		case "x":
			return (
				<svg
					aria-hidden="true"
					className={className}
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="currentColor"
				>
					<path d="M3.72 3.72a.75.75 0 0 1 1.06 0L8 6.94l3.22-3.22a.749.749 0 0 1 1.275.326.749.749 0 0 1-.215.734L9.06 8l3.22 3.22a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215L8 9.06l-3.22 3.22a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L6.94 8 3.72 4.78a.75.75 0 0 1 0-1.06Z" />
				</svg>
			);
		case "spinner":
			return (
				<svg
					aria-hidden="true"
					className={`${className ?? ""} animate-spin`}
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="none"
					stroke="currentColor"
					strokeWidth="2"
				>
					<circle cx="8" cy="8" r="6" opacity="0.25" />
					<path d="M8 2a6 6 0 0 1 6 6" strokeLinecap="round" />
				</svg>
			);
		case "clock":
			return (
				<svg
					aria-hidden="true"
					className={className}
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="currentColor"
				>
					<path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0ZM1.5 8a6.5 6.5 0 1 0 13 0 6.5 6.5 0 0 0-13 0Zm7-3.25v2.992l2.028.812a.75.75 0 0 1-.557 1.392l-2.5-1A.751.751 0 0 1 7 8.25v-3.5a.75.75 0 0 1 1.5 0Z" />
				</svg>
			);
		case "minus":
			return (
				<svg
					aria-hidden="true"
					className={className}
					width="16"
					height="16"
					viewBox="0 0 16 16"
					fill="currentColor"
				>
					<path d="M2 7.75A.75.75 0 0 1 2.75 7h10a.75.75 0 0 1 0 1.5h-10A.75.75 0 0 1 2 7.75Z" />
				</svg>
			);
	}
}

// ---------------------------------------------------------------------------
// Основной компонент
// ---------------------------------------------------------------------------

/**
 * GitLabPipelineStatus -- секция GitLab CI/CD на странице проекта.
 * Отображает последние пайплайны GitLab с бейджами статуса.
 */
export function GitLabPipelineStatus({ projectId }: GitLabPipelineStatusProps) {
	const {
		data: pipelines,
		isLoading,
		error,
	} = useQuery({
		queryKey: ["gitlab-pipelines", projectId],
		queryFn: () => listGitLabPipelines(projectId),
		refetchInterval: 30_000,
	});

	if (isLoading) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					GitLab CI/CD
				</h4>
				<div className="animate-pulse space-y-2">
					{["gl-sk-1", "gl-sk-2", "gl-sk-3"].map((sk) => (
						<div
							key={sk}
							className="h-12 w-full rounded-lg bg-gray-200 dark:bg-gray-700"
						/>
					))}
				</div>
			</div>
		);
	}

	if (error || !pipelines) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					GitLab CI/CD
				</h4>
				<p className="text-sm text-gray-500 dark:text-gray-400">
					Не удалось загрузить данные пайплайнов GitLab
				</p>
			</div>
		);
	}

	if (pipelines.length === 0) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					GitLab CI/CD
				</h4>
				<p className="text-sm text-gray-500 dark:text-gray-400">
					Нет запусков пайплайнов GitLab
				</p>
			</div>
		);
	}

	return (
		<div className="space-y-3">
			<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
				GitLab CI/CD
			</h4>
			<div className="space-y-1.5">
				{pipelines.slice(0, 10).map((pipeline) => {
					const cfg = getStatusConfig(pipeline);
					return (
						<a
							key={pipeline.id}
							href={pipeline.web_url}
							target="_blank"
							rel="noopener noreferrer"
							className="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm transition-colors hover:bg-gray-50 dark:hover:bg-gray-800"
						>
							{/* Иконка статуса */}
							<span
								className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${cfg.bgColor}`}
							>
								<StatusIcon icon={cfg.icon} className={cfg.color} />
							</span>

							{/* Информация о пайплайне */}
							<div className="min-w-0 flex-1">
								<div className="flex items-center gap-1.5">
									<span className="truncate font-medium">
										Pipeline #{pipeline.id}
									</span>
								</div>
								<div className="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
									<span>{pipeline.ref}</span>
									<span>-</span>
									<span>{pipeline.sha.substring(0, 8)}</span>
									<span>-</span>
									<span>{formatRelativeTime(pipeline.created_at)}</span>
								</div>
							</div>

							{/* Бейдж статуса */}
							<span
								className={`shrink-0 rounded-full px-2 py-0.5 text-[10px] font-semibold ${cfg.bgColor} ${cfg.color}`}
							>
								{cfg.label}
							</span>
						</a>
					);
				})}
			</div>
		</div>
	);
}
