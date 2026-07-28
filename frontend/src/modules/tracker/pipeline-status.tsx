import { useQuery } from "@tanstack/react-query";
import { listWorkflowRuns, type WorkflowRunResponse } from "@/api/github";

// ---------------------------------------------------------------------------
// Типы
// ---------------------------------------------------------------------------

/** Пропсы компонента PipelineStatus. */
interface PipelineStatusProps {
	/** ID проекта для получения запусков CI/CD. */
	projectId: string;
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

/**
 * Возвращает конфигурацию бейджа статуса CI/CD пайплайна.
 * Зелёная галочка для success, красный крестик для failure,
 * жёлтый спиннер для in_progress, серый для остальных.
 */
function getStatusConfig(run: WorkflowRunResponse): {
	icon: "check" | "x" | "spinner" | "clock" | "minus";
	color: string;
	bgColor: string;
	label: string;
} {
	if (run.status === "in_progress") {
		return {
			icon: "spinner",
			color: "text-yellow-600 dark:text-yellow-400",
			bgColor: "bg-yellow-100 dark:bg-yellow-900/30",
			label: "В процессе",
		};
	}
	if (run.status === "queued") {
		return {
			icon: "clock",
			color: "text-gray-500 dark:text-gray-400",
			bgColor: "bg-gray-100 dark:bg-gray-800",
			label: "В очереди",
		};
	}
	// завершённые
	switch (run.conclusion) {
		case "success":
			return {
				icon: "check",
				color: "text-green-600 dark:text-green-400",
				bgColor: "bg-green-100 dark:bg-green-900/30",
				label: "Успешно",
			};
		case "failure":
			return {
				icon: "x",
				color: "text-red-600 dark:text-red-400",
				bgColor: "bg-red-100 dark:bg-red-900/30",
				label: "Ошибка",
			};
		case "cancelled":
			return {
				icon: "minus",
				color: "text-gray-500 dark:text-gray-400",
				bgColor: "bg-gray-100 dark:bg-gray-800",
				label: "Отменён",
			};
		default:
			return {
				icon: "minus",
				color: "text-gray-500 dark:text-gray-400",
				bgColor: "bg-gray-100 dark:bg-gray-800",
				label: run.conclusion || run.status,
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
 * PipelineStatus -- секция CI/CD на странице проекта.
 * Отображает последние запуски GitHub Actions workflow с бейджами статуса
 * (зелёная галочка / красный крестик / жёлтый спиннер).
 * Ссылается на страницу GitHub Actions для полных деталей.
 */
export function PipelineStatus({ projectId }: PipelineStatusProps) {
	const {
		data: runs,
		isLoading,
		error,
	} = useQuery({
		queryKey: ["github-workflows", projectId],
		queryFn: () => listWorkflowRuns(projectId),
		// Poll only while CI runs are active (queued / in_progress).
		refetchInterval: (query) => {
			const data = query.state.data;
			if (!data?.length) return false;
			const active = data.some(
				(r) => r.status === "queued" || r.status === "in_progress",
			);
			return active ? 30_000 : false;
		},
	});

	if (isLoading) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					CI/CD
				</h4>
				<div className="animate-pulse space-y-2">
					{["pipe-sk-1", "pipe-sk-2", "pipe-sk-3"].map((sk) => (
						<div
							key={sk}
							className="h-12 w-full rounded-lg bg-gray-200 dark:bg-gray-700"
						/>
					))}
				</div>
			</div>
		);
	}

	if (error || !runs) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					CI/CD
				</h4>
				<p className="text-sm text-gray-500 dark:text-gray-400">
					Не удалось загрузить данные пайплайнов
				</p>
			</div>
		);
	}

	if (runs.length === 0) {
		return (
			<div className="space-y-3">
				<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
					CI/CD
				</h4>
				<p className="text-sm text-gray-500 dark:text-gray-400">
					Нет запусков CI/CD пайплайнов
				</p>
			</div>
		);
	}

	return (
		<div className="space-y-3">
			<h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
				CI/CD
			</h4>
			<div className="space-y-1.5">
				{runs.slice(0, 10).map((run) => {
					const cfg = getStatusConfig(run);
					return (
						<a
							key={run.id}
							href={run.html_url}
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

							{/* Информация о запуске */}
							<div className="min-w-0 flex-1">
								<div className="flex items-center gap-1.5">
									<span className="truncate font-medium">{run.name}</span>
									<span className="shrink-0 text-xs text-gray-400">
										#{run.run_number}
									</span>
								</div>
								<div className="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
									<span>{run.branch}</span>
									<span>-</span>
									<span>{formatRelativeTime(run.created_at)}</span>
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
