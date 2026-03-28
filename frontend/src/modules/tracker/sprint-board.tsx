import { Link } from "@tanstack/react-router";
import {
	Calendar,
	CheckCircle2,
	ChevronRight,
	Clock,
	Flame,
	ListTodo,
	PlayCircle,
	Target,
	Zap,
} from "lucide-react";
import { useEffect, useState } from "react";
import type { SprintResponse } from "@/api/sprint";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { useSprintStore } from "@/stores/sprint-store";

/** Пропсы компонента SprintBoard. */
interface SprintBoardProps {
	/** ID проекта. */
	projectId: string;
}

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

/** Вычислить прогресс спринта по дням (для UI). */
function getProgressPercent(startDate: string, endDate: string): number {
	const now = new Date().getTime();
	const start = new Date(startDate).getTime();
	const end = new Date(endDate).getTime();
	if (now >= end) return 100;
	if (now <= start) return 0;
	return Math.round(((now - start) / (end - start)) * 100);
}

/**
 * SprintBoard отображает все спринты проекта в виде карточек.
 * Кнопки "Начать спринт" и "Завершить спринт" управляют жизненным циклом.
 */
export function SprintBoard({ projectId }: SprintBoardProps) {
	const {
		sprints,
		activeSprint,
		fetchSprints,
		startSprint,
		completeSprint,
		loading,
	} = useSprintStore();

	const [confirmComplete, setConfirmComplete] = useState(false);
	const [selectedSprint, setSelectedSprint] =
		useState<SprintResponse | null>(null);

	useEffect(() => {
		fetchSprints(projectId);
	}, [projectId, fetchSprints]);

	const handleStartSprint = async (
		e: React.MouseEvent,
		sprintId: string,
	) => {
		e.preventDefault();
		e.stopPropagation();
		await startSprint(projectId, sprintId);
	};

	const handleCompleteSprint = async () => {
		if (!selectedSprint) return;
		await completeSprint(projectId, selectedSprint.id);
		setConfirmComplete(false);
		setSelectedSprint(null);
	};

	const openCompleteDialog = (e: React.MouseEvent, sprint: SprintResponse) => {
		e.preventDefault();
		e.stopPropagation();
		setSelectedSprint(sprint);
		setConfirmComplete(true);
	};

	if (loading && sprints.length === 0) {
		return (
			<div className="flex items-center justify-center p-16">
				<div className="flex flex-col items-center gap-3">
					<div className="h-8 w-8 animate-spin rounded-full border-2 border-blue border-t-transparent" />
					<p className="text-[13px] text-t2">
						Загрузка спринтов...
					</p>
				</div>
			</div>
		);
	}

	if (sprints.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center py-20 text-center">
				<div className="mb-4 flex h-16 w-16 items-center justify-center rounded-[var(--r4)] bg-gradient-to-br from-[var(--blueS)]/20 to-[var(--purpleS)]/20">
					<Flame className="h-8 w-8 text-blue" />
				</div>
				<h3 className="text-[14px] font-bold text-t1">
					Нет спринтов
				</h3>
				<p className="mt-1 max-w-sm text-[13px] text-t2">
					Создайте первый спринт, чтобы организовать работу команды по
					итерациям.
				</p>
			</div>
		);
	}

	// Sort: active first, then planning, then completed
	const sorted = [...sprints].sort((a, b) => {
		const order = { active: 0, planning: 1, completed: 2 };
		return (
			(order[a.status] ?? 3) - (order[b.status] ?? 3) ||
			new Date(b.start_date).getTime() - new Date(a.start_date).getTime()
		);
	});

	return (
		<div className="space-y-6 animate-[eu_450ms_var(--spring)_both]">
			{/* Page header */}
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<div className="flex h-10 w-10 items-center justify-center rounded-[var(--r3)] bg-gradient-to-br from-[var(--blueS)]/20 to-[var(--purpleS)]/20 text-blue">
						<Flame className="h-5 w-5" />
					</div>
					<div>
						<h1 className="text-[18px] font-extrabold tracking-[-0.03em] text-t1">Спринты</h1>
						<p className="mt-0.5 text-[13px] text-t2">
							{sprints.length} спринт
							{sprints.length === 1
								? ""
								: sprints.length < 5
									? "а"
									: "ов"}{" "}
							в проекте
						</p>
					</div>
				</div>
			</div>

			{/* Sprint stat bar */}
			<div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
				{[
					{
						label: "Активных",
						value: sprints.filter((s) => s.status === "active").length,
						icon: Zap,
						color: "text-emerald-400",
						bg: "from-emerald-500/10 to-emerald-500/5",
					},
					{
						label: "В планировании",
						value: sprints.filter((s) => s.status === "planning").length,
						icon: Clock,
						color: "text-blue",
						bg: "from-blue-500/10 to-blue-500/5",
					},
					{
						label: "Завершённых",
						value: sprints.filter((s) => s.status === "completed").length,
						icon: CheckCircle2,
						color: "text-zinc-400",
						bg: "from-zinc-500/10 to-zinc-500/5",
					},
				].map((stat) => (
					<div
						key={stat.label}
						className="bg-gradient-to-b from-card to-raised rounded-[var(--r4)] border border-b1 p-4 transition-all duration-250 hover:border-b2 hover:shadow-[var(--sh2)]"
					>
						<div className="flex items-center gap-3">
							<div
								className={`flex h-9 w-9 items-center justify-center rounded-[var(--r3)] bg-gradient-to-br ${stat.bg}`}
							>
								<stat.icon className={`h-4 w-4 ${stat.color}`} />
							</div>
							<div>
								<p className="text-[24px] font-black tracking-[-0.03em] tabular-nums text-t1">
									{stat.value}
								</p>
								<p className="text-[12px] text-t3 font-medium">
									{stat.label}
								</p>
							</div>
						</div>
					</div>
				))}
			</div>

			{/* Sprint cards grid */}
			<div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
				{sorted.map((sprint) => {
					const badge = getStatusBadge(sprint.status);
					const BadgeIcon = badge.icon;
					const progress = getProgressPercent(
						sprint.start_date,
						sprint.end_date,
					);
					const daysLeft = getDaysRemaining(sprint.end_date);

					return (
						<Link
							key={sprint.id}
							to="/projects/$projectId/sprints/$sprintId"
							params={{ projectId, sprintId: sprint.id }}
							className="group block"
						>
							<div
								className={`bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-5 transition-all duration-250 relative overflow-hidden hover:border-b2 hover:-translate-y-[2px] hover:shadow-[var(--sh2)] ${
									sprint.status === "active"
										? "ring-1 ring-emerald-500/20"
										: ""
								}`}
							>
								{/* Header row */}
								<div className="flex items-start justify-between gap-3">
									<div className="min-w-0 flex-1">
										<div className="flex items-center gap-2.5">
											<h3 className="truncate text-[14px] font-bold text-t1">
												{sprint.name}
											</h3>
											<span
												className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ${badge.className}`}
											>
												<BadgeIcon className="h-3 w-3" />
												{badge.label}
											</span>
										</div>
										{sprint.goal && (
											<p className="mt-1 line-clamp-2 text-[13px] text-t2">
												<Target className="mr-1 inline h-3 w-3 align-[-2px] opacity-50" />
												{sprint.goal}
											</p>
										)}
									</div>
									<ChevronRight className="mt-0.5 h-4 w-4 flex-shrink-0 text-t3 opacity-0 transition-opacity group-hover:opacity-100" />
								</div>

								{/* Date + stats row */}
								<div className="mt-4 flex flex-wrap items-center gap-4 text-[12px] text-t3">
									<span className="inline-flex items-center gap-1.5">
										<Calendar className="h-3.5 w-3.5 opacity-60" />
										{new Date(sprint.start_date).toLocaleDateString(
											"ru-RU",
											{
												day: "numeric",
												month: "short",
											},
										)}{" "}
										&mdash;{" "}
										{new Date(sprint.end_date).toLocaleDateString(
											"ru-RU",
											{
												day: "numeric",
												month: "short",
											},
										)}
									</span>
									<span className="inline-flex items-center gap-1.5">
										<ListTodo className="h-3.5 w-3.5 opacity-60" />
										{sprint.issue_count} задач
									</span>
									{sprint.status === "active" && (
										<span className="inline-flex items-center gap-1.5">
											<Clock className="h-3.5 w-3.5 opacity-60" />
											{daysLeft} дн. осталось
										</span>
									)}
								</div>

								{/* Progress bar */}
								{sprint.status !== "completed" && (
									<div className="mt-3">
										<div className="mb-1 flex justify-between text-[11px] text-t3">
											<span>Прогресс по времени</span>
											<span className="tabular-nums">
												{progress}%
											</span>
										</div>
										<div className="h-1.5 w-full overflow-hidden rounded-full bg-white/[0.06]">
											<div
												className="animate-progress-fill h-full rounded-full bg-gradient-to-r from-emerald-500 to-emerald-400"
												style={{ width: `${progress}%` }}
											/>
										</div>
									</div>
								)}

								{sprint.status === "completed" && (
									<div className="mt-3">
										<div className="h-1.5 w-full overflow-hidden rounded-full bg-white/[0.06]">
											<div className="h-full w-full rounded-full bg-zinc-500/40" />
										</div>
									</div>
								)}

								{/* Actions */}
								<div className="mt-4 flex items-center gap-2">
									{sprint.status === "planning" && (
										<Button
											size="sm"
											disabled={!!activeSprint}
											className="btn-glow bg-gradient-to-r from-blue-600 to-blue-500 text-white shadow-md shadow-blue-500/20 hover:from-blue-500 hover:to-blue-400"
											onClick={(e) =>
												handleStartSprint(e, sprint.id)
											}
										>
											<PlayCircle className="mr-1.5 h-3.5 w-3.5" />
											Начать спринт
										</Button>
									)}
									{sprint.status === "active" && (
										<Button
											size="sm"
											variant="outline"
											className="border-b1 text-t2 hover:bg-[var(--hover)] hover:text-t1"
											onClick={(e) =>
												openCompleteDialog(e, sprint)
											}
										>
											<CheckCircle2 className="mr-1.5 h-3.5 w-3.5" />
											Завершить спринт
										</Button>
									)}
								</div>
							</div>
						</Link>
					);
				})}
			</div>

			{/* Confirm complete dialog */}
			<Dialog open={confirmComplete} onOpenChange={setConfirmComplete}>
				<DialogContent className="rounded-[var(--r4)] border-b1">
					<DialogHeader>
						<DialogTitle>Завершить спринт?</DialogTitle>
						<DialogDescription>
							Незавершённые задачи будут автоматически перенесены в
							следующий спринт.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setConfirmComplete(false)}
							className="border-b1 text-t2 hover:text-t1"
						>
							Отмена
						</Button>
						<Button
							onClick={handleCompleteSprint}
							className="btn-glow bg-gradient-to-r from-blue-600 to-blue-500 text-white"
						>
							Завершить
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
