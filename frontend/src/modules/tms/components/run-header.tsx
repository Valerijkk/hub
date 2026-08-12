/**
 * Заголовок страницы прогона тестов -- breadcrumb, название, статус, прогресс-бар, действия.
 */
import { Link } from "@tanstack/react-router";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/** Конфигурация статусов прогона. */
export const RUN_STATUS_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string; borderColor: string }
> = {
	created: {
		bg: "bg-gray/[0.08]",
		text: "text-gray",
		label: "Создан",
		borderColor: "border-gray/[0.15]",
	},
	in_progress: {
		bg: "bg-blue/[0.08]",
		text: "text-blue",
		label: "Выполняется",
		borderColor: "border-blue/[0.15]",
	},
	completed: {
		bg: "bg-green/[0.08]",
		text: "text-green",
		label: "Завершён",
		borderColor: "border-green/[0.15]",
	},
	aborted: {
		bg: "bg-red/[0.08]",
		text: "text-red",
		label: "Прерван",
		borderColor: "border-red/[0.15]",
	},
};

/** Пропсы компонента RunHeader. */
interface RunHeaderProps {
	projectId: string;
	run: {
		title: string;
		description?: string | null;
		status: string;
		total_count: number;
		passed_count: number;
		failed_count: number;
		broken_count: number;
		skipped_count: number;
		assignees?: { id: string; name: string }[];
		results?: { id: string }[];
	};
	onStart: () => void;
	onComplete: () => void;
	onReopen: () => void;
	onAddCases: () => void;
	isStarting: boolean;
	isCompleting: boolean;
	isReopening: boolean;
	/** false = скрыть Start/Complete/Reopen/Add cases (viewer/guest, P2-FE-16). */
	canMutate?: boolean;
}

/** Заголовок прогона тестов: breadcrumb, название, статус, прогресс-бар, кнопки управления. */
export function RunHeader({
	projectId,
	run,
	onStart,
	onComplete,
	onReopen,
	onAddCases,
	isStarting,
	isCompleting,
	isReopening,
	canMutate = true,
}: RunHeaderProps) {
	// biome-ignore lint/style/noNonNullAssertion: RUN_STATUS_CONFIG.created is always defined
	const statusCfg = (RUN_STATUS_CONFIG[run.status] ??
		RUN_STATUS_CONFIG.created)!;
	const passRate =
		run.total_count > 0
			? Math.round((run.passed_count / run.total_count) * 100)
			: 0;

	return (
		<>
			{/* Навигация (breadcrumb) */}
			<div className="mb-4 flex items-center gap-1.5 text-sm">
				<Link
					to="/projects/$projectId/tms/runs"
					params={{ projectId }}
					className="text-t2 hover:text-t1 transition-colors"
				>
					Прогоны
				</Link>
				<svg
					aria-hidden="true"
					className="h-3.5 w-3.5 text-t2/40"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
				</svg>
				<span className="font-medium line-clamp-1">{run.title}</span>
			</div>

			{/* Заголовок и действия */}
			<div className="mb-6 flex items-start justify-between gap-4">
				<div className="min-w-0">
					<div className="flex items-center gap-3 mb-2">
						<h2 className="text-xl font-semibold leading-tight">{run.title}</h2>
						<Badge
							variant="secondary"
							className={cn(
								"text-[11px] font-semibold shrink-0 border",
								statusCfg.bg,
								statusCfg.text,
								statusCfg.borderColor,
							)}
						>
							{statusCfg.label}
						</Badge>
					</div>
					{run.description && (
						<p className="text-sm text-t2 mb-2">{run.description}</p>
					)}
					{/* Прогресс-бар в заголовке */}
					{run.total_count > 0 && (
						<div className="flex items-center gap-3">
							<div className="flex h-[3px] w-56 overflow-hidden rounded-full bg-white/[0.04] gap-[2px]">
								{run.passed_count > 0 && (
									<div
										className="h-full rounded-full bg-green shadow-[0_0_8px_var(--gG)] animate-[progressGrow_800ms_var(--ease)]"
										style={{
											width: `${Math.max((run.passed_count / run.total_count) * 100, 4)}%`,
										}}
									/>
								)}
								{run.failed_count > 0 && (
									<div
										className="h-full rounded-full bg-red shadow-[0_0_8px_var(--gR)] animate-[progressGrow_800ms_var(--ease)]"
										style={{
											width: `${Math.max((run.failed_count / run.total_count) * 100, 4)}%`,
										}}
									/>
								)}
								{run.broken_count > 0 && (
									<div
										className="h-full rounded-full bg-orange animate-[progressGrow_800ms_var(--ease)]"
										style={{
											width: `${Math.max((run.broken_count / run.total_count) * 100, 4)}%`,
										}}
									/>
								)}
								{run.skipped_count > 0 && (
									<div
										className="h-full rounded-full bg-gray opacity-40 animate-[progressGrow_800ms_var(--ease)]"
										style={{
											width: `${Math.max((run.skipped_count / run.total_count) * 100, 4)}%`,
										}}
									/>
								)}
							</div>
							<span className="text-xs text-t2 tabular-nums font-medium">
								{passRate}% passed
							</span>
						</div>
					)}
					{/* Назначенные */}
					{(run.assignees ?? []).length > 0 && (
						<div className="mt-2 flex items-center gap-1.5">
							<span className="text-[10px] text-t2 uppercase tracking-wider font-medium">
								Назначены:
							</span>
							{(run.assignees ?? []).map((u) => (
								<Badge
									key={u.id}
									variant="outline"
									className="text-[10px] font-medium"
								>
									{u.name}
								</Badge>
							))}
						</div>
					)}
				</div>

				{canMutate ? (
					<div className="flex gap-2 shrink-0">
						{run.status === "created" && (
							<>
								<Button
									variant="outline"
									size="sm"
									className="h-8"
									onClick={onAddCases}
								>
									<svg
										aria-hidden="true"
										className="mr-1 h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M12 4v16m8-8H4"
										/>
									</svg>
									Добавить кейсы
								</Button>
								<Button
									size="sm"
									className="h-8 gradient-tms text-white"
									disabled={isStarting || (run.results ?? []).length === 0}
									onClick={onStart}
								>
									<svg
										aria-hidden="true"
										className="mr-1 h-3.5 w-3.5"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
										strokeWidth={2}
									>
										<path
											strokeLinecap="round"
											strokeLinejoin="round"
											d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
										/>
									</svg>
									Запустить
								</Button>
							</>
						)}
						{run.status === "in_progress" && (
							<Button
								size="sm"
								className="h-8"
								disabled={isCompleting}
								onClick={onComplete}
							>
								<svg
									aria-hidden="true"
									className="mr-1 h-3.5 w-3.5"
									fill="none"
									viewBox="0 0 24 24"
									stroke="currentColor"
									strokeWidth={2}
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M5 13l4 4L19 7"
									/>
								</svg>
								Завершить
							</Button>
						)}
						{(run.status === "completed" || run.status === "aborted") && (
							<Button
								variant="outline"
								size="sm"
								className="h-8"
								disabled={isReopening}
								onClick={onReopen}
							>
								Переоткрыть
							</Button>
						)}
					</div>
				) : null}
			</div>
		</>
	);
}
