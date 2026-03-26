import { Link, createFileRoute } from "@tanstack/react-router";
import {
	ArrowLeft,
	Calendar,
	CheckCircle2,
	Clock,
	Flame,
	ListTodo,
	Target,
	Zap,
} from "lucide-react";
import { useEffect, useState } from "react";
import { getSprint, type SprintResponse } from "@/api/sprint";
import { BurndownChart } from "@/modules/tracker/burndown-chart";
import { SprintRetro } from "@/modules/tracker/sprint-retro";

/** Получить конфигурацию бейджа по статусу спринта. */
function getStatusBadge(status: string) {
	switch (status) {
		case "active":
			return {
				label: "Активный",
				className:
					"bg-emerald-500/10 text-emerald-400 border border-emerald-500/20",
				icon: Zap,
			};
		case "planning":
			return {
				label: "Планирование",
				className: "bg-blue-500/10 text-blue-400 border border-blue-500/20",
				icon: Clock,
			};
		case "completed":
			return {
				label: "Завершён",
				className:
					"bg-zinc-500/10 text-zinc-400 border border-zinc-500/20",
				icon: CheckCircle2,
			};
		default:
			return {
				label: status,
				className: "bg-zinc-500/10 text-zinc-400 border border-zinc-500/20",
				icon: Clock,
			};
	}
}

/** Вычислить дней осталось до конца спринта. */
function getDaysRemaining(endDate: string): number {
	const now = new Date();
	const end = new Date(endDate);
	const diff = Math.ceil(
		(end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24),
	);
	return Math.max(0, diff);
}

/** Вычислить длительность спринта в днях. */
function getSprintDuration(startDate: string, endDate: string): number {
	const start = new Date(startDate);
	const end = new Date(endDate);
	return Math.ceil(
		(end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24),
	);
}

/** Вычислить прогресс спринта по дням. */
function getProgressPercent(startDate: string, endDate: string): number {
	const now = new Date().getTime();
	const start = new Date(startDate).getTime();
	const end = new Date(endDate).getTime();
	if (now >= end) return 100;
	if (now <= start) return 0;
	return Math.round(((now - start) / (end - start)) * 100);
}

/**
 * Страница детального просмотра спринта.
 * Включает burndown-график и ретроспективу.
 */
function SprintDetailPage() {
	const { projectId, sprintId } = Route.useParams();
	const [sprint, setSprint] = useState<SprintResponse | null>(null);

	useEffect(() => {
		getSprint(projectId, sprintId).then(setSprint);
	}, [projectId, sprintId]);

	if (!sprint) {
		return (
			<div className="flex items-center justify-center p-16">
				<div className="flex flex-col items-center gap-3">
					<div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
					<p className="text-sm text-muted-foreground">Загрузка...</p>
				</div>
			</div>
		);
	}

	const badge = getStatusBadge(sprint.status);
	const BadgeIcon = badge.icon;
	const daysLeft = getDaysRemaining(sprint.end_date);
	const duration = getSprintDuration(sprint.start_date, sprint.end_date);
	const progress = getProgressPercent(sprint.start_date, sprint.end_date);

	const retroNotes = sprint.retro_notes as
		| {
				went_well?: string[];
				to_improve?: string[];
				actions?: string[];
		  }
		| undefined;

	return (
		<div className="mx-auto max-w-5xl space-y-6 animate-fade-in">
			{/* Back link */}
			<Link
				to="/projects/$projectId/sprints"
				params={{ projectId }}
				className="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
			>
				<ArrowLeft className="h-3.5 w-3.5" />
				Все спринты
			</Link>

			{/* Sprint header card */}
			<div className="card-premium rounded-2xl border border-white/[0.06] bg-card p-6">
				<div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
					<div className="min-w-0 flex-1">
						<div className="flex items-center gap-3">
							<div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-blue-500/20 to-purple-500/20 text-blue-500 shadow-sm">
								<Flame className="h-5 w-5" />
							</div>
							<div>
								<div className="flex items-center gap-2.5">
									<h1 className="text-xl font-bold text-foreground">
										{sprint.name}
									</h1>
									<span
										className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${badge.className}`}
									>
										<BadgeIcon className="h-3 w-3" />
										{badge.label}
									</span>
								</div>
								{sprint.goal && (
									<p className="mt-1 text-sm text-muted-foreground">
										<Target className="mr-1 inline h-3 w-3 align-[-2px] opacity-50" />
										{sprint.goal}
									</p>
								)}
							</div>
						</div>
					</div>
				</div>

				{/* Stats row */}
				<div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-4">
					{[
						{
							label: "Задач",
							value: sprint.issue_count,
							icon: ListTodo,
							color: "text-blue-400",
							bg: "from-blue-500/10 to-blue-500/5",
						},
						{
							label: "Длительность",
							value: `${duration} дн.`,
							icon: Calendar,
							color: "text-purple-400",
							bg: "from-purple-500/10 to-purple-500/5",
						},
						{
							label: "Осталось",
							value:
								sprint.status === "completed"
									? "Завершён"
									: `${daysLeft} дн.`,
							icon: Clock,
							color: "text-amber-400",
							bg: "from-amber-500/10 to-amber-500/5",
						},
						{
							label: "Период",
							value: `${new Date(sprint.start_date).toLocaleDateString("ru-RU", { day: "numeric", month: "short" })} — ${new Date(sprint.end_date).toLocaleDateString("ru-RU", { day: "numeric", month: "short" })}`,
							icon: Calendar,
							color: "text-emerald-400",
							bg: "from-emerald-500/10 to-emerald-500/5",
						},
					].map((stat) => (
						<div
							key={stat.label}
							className="widget-card rounded-xl border border-white/[0.06] bg-card p-3"
						>
							<div className="flex items-center gap-2.5">
								<div
									className={`flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br ${stat.bg}`}
								>
									<stat.icon
										className={`h-3.5 w-3.5 ${stat.color}`}
									/>
								</div>
								<div>
									<p className="text-sm font-bold tabular-nums leading-tight">
										{stat.value}
									</p>
									<p className="text-[11px] text-muted-foreground">
										{stat.label}
									</p>
								</div>
							</div>
						</div>
					))}
				</div>

				{/* Progress bar */}
				{sprint.status !== "completed" ? (
					<div className="mt-4">
						<div className="mb-1.5 flex justify-between text-xs text-muted-foreground">
							<span>Прогресс по времени</span>
							<span className="tabular-nums">{progress}%</span>
						</div>
						<div className="h-1.5 w-full overflow-hidden rounded-full bg-white/[0.06]">
							<div
								className="animate-progress-fill h-full rounded-full bg-gradient-to-r from-emerald-500 to-emerald-400"
								style={{ width: `${progress}%` }}
							/>
						</div>
					</div>
				) : (
					<div className="mt-4">
						<div className="mb-1.5 flex justify-between text-xs text-muted-foreground">
							<span>Спринт завершён</span>
							<span className="tabular-nums">100%</span>
						</div>
						<div className="h-1.5 w-full overflow-hidden rounded-full bg-white/[0.06]">
							<div className="h-full w-full rounded-full bg-zinc-500/40" />
						</div>
					</div>
				)}
			</div>

			{/* Burndown chart */}
			<BurndownChart projectId={projectId} sprintId={sprintId} />

			{/* Retro section */}
			{(sprint.status === "active" || sprint.status === "completed") && (
				<SprintRetro
					projectId={projectId}
					sprintId={sprintId}
					initialNotes={retroNotes}
				/>
			)}
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/sprints/$sprintId",
)({
	component: SprintDetailPage,
});
