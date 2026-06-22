import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import {
	createTestRun,
	type ListTestRunsParams,
	listTestRuns,
	type TestRunType,
} from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";
import { formatRelativeTime } from "@/lib/format-time";
import { cn } from "@/lib/utils";

/** Конфиг статусов прогона с иконками. */
const RUN_STATUS_CONFIG: Record<
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

/** Конфиг типов прогона. */
const RUN_TYPE_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string }
> = {
	manual: {
		bg: "bg-blue/[0.08]",
		text: "text-blue",
		label: "Ручной",
	},
	automatic: {
		bg: "bg-purple/[0.08]",
		text: "text-purple",
		label: "Авто",
	},
	mixed: {
		bg: "bg-orange/[0.08]",
		text: "text-orange",
		label: "Смешанный",
	},
};

/**
 * Список прогонов — UI уровня Allure TestOps.
 */
function TestRunsPage() {
	const { projectId } = Route.useParams();
	const queryClient = useQueryClient();
	const [statusFilter, setStatusFilter] = useState("");
	const [offset, setOffset] = useState(0);
	const [createOpen, setCreateOpen] = useState(false);
	const [newTitle, setNewTitle] = useState("");
	const [newType, setNewType] = useState<TestRunType>("manual");
	const [newDescription, setNewDescription] = useState("");

	const params: ListTestRunsParams = {
		offset,
		limit: 50,
		status: (statusFilter || undefined) as ListTestRunsParams["status"],
	};

	const { data, isLoading } = useQuery({
		queryKey: ["tms-runs", projectId, params],
		queryFn: () => listTestRuns(projectId, params),
	});

	const createMutation = useMutation({
		mutationFn: () =>
			createTestRun(projectId, {
				title: newTitle,
				run_type: newType,
				description: newDescription || undefined,
			}),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["tms-runs", projectId] });
			setCreateOpen(false);
			setNewTitle("");
			setNewDescription("");
			toast.success("Прогон создан");
		},
		onError: (err) => toastApiError(err, "Ошибка создания прогона"),
	});

	const runs = data?.items ?? [];
	const total = data?.total ?? 0;

	function passRate(r: (typeof runs)[number]): number {
		if (r.total_count === 0) return 0;
		return Math.round((r.passed_count / r.total_count) * 100);
	}

	return (
		<div className="flex flex-col flex-1 animate-[enterUp_450ms_var(--spring)_both]">
			{/* Верхняя панель */}
			<div className="px-6 py-4 border-b border-b1 flex items-center gap-[14px] bg-[rgba(13,14,18,0.5)] backdrop-blur-[12px]">
				<h2 className="text-[15px] font-bold tracking-[-0.015em]">Прогоны</h2>
				<span className="text-[11.5px] font-semibold text-t3 bg-white/[0.035] px-3 py-[3px] rounded-full">
					{total} {total === 1 ? "прогон" : total < 5 ? "прогона" : "прогонов"}
				</span>
				<div className="flex-1" />

				<Select
					value={statusFilter || "all"}
					onValueChange={(v) => {
						setStatusFilter(v === "all" ? "" : v);
						setOffset(0);
					}}
				>
					<SelectTrigger className="w-[140px] h-8 text-xs rounded-[var(--r2)] border-b1 bg-white/[0.03]">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">Все статусы</SelectItem>
						<SelectItem value="created">Создан</SelectItem>
						<SelectItem value="in_progress">Выполняется</SelectItem>
						<SelectItem value="completed">Завершён</SelectItem>
						<SelectItem value="aborted">Прерван</SelectItem>
					</SelectContent>
				</Select>

				{statusFilter && (
					<Button
						variant="ghost"
						size="sm"
						className="h-8 text-xs text-t3"
						onClick={() => {
							setStatusFilter("");
							setOffset(0);
						}}
					>
						<svg
							aria-hidden="true"
							className="mr-1 h-3 w-3"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={2}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
						Сбросить
					</Button>
				)}

				<Dialog open={createOpen} onOpenChange={setCreateOpen}>
					<DialogTrigger asChild>
						<Button size="sm" className="gradient-tms text-white shadow-sm h-8">
							<svg
								aria-hidden="true"
								className="mr-1.5 h-3.5 w-3.5"
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
							Новый прогон
						</Button>
					</DialogTrigger>
					<DialogContent>
						<DialogHeader>
							<DialogTitle>Новый прогон тестов</DialogTitle>
						</DialogHeader>
						<div className="space-y-4">
							<Input
								placeholder="Название прогона"
								value={newTitle}
								onChange={(e) => setNewTitle(e.target.value)}
							/>
							<Select
								value={newType}
								onValueChange={(v) => setNewType(v as TestRunType)}
							>
								<SelectTrigger className="w-full rounded-xl border bg-card px-3 py-2 text-sm">
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="manual">Ручной</SelectItem>
									<SelectItem value="automatic">Автоматический</SelectItem>
									<SelectItem value="mixed">Смешанный</SelectItem>
								</SelectContent>
							</Select>
							<textarea
								className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
								placeholder="Описание (необязательно)"
								rows={3}
								value={newDescription}
								onChange={(e) => setNewDescription(e.target.value)}
							/>
							<Button
								className="w-full"
								disabled={!newTitle.trim() || createMutation.isPending}
								onClick={() => createMutation.mutate()}
							>
								{createMutation.isPending ? "Создание..." : "Создать"}
							</Button>
						</div>
					</DialogContent>
				</Dialog>
			</div>

			{/* Список прогонов */}
			{isLoading ? (
				<div className="grid grid-cols-[repeat(auto-fill,minmax(350px,1fr))] gap-[18px] p-6">
					{[1, 2, 3, 4, 5, 6].map((i) => (
						<div
							key={`run-sk-${i}`}
							className="tms-run-card bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] p-6 space-y-3"
						>
							<div className="flex justify-between">
								<div className="h-5 w-3/4 rounded animate-shimmer" />
								<div className="h-5 w-16 rounded-full animate-shimmer" />
							</div>
							<div className="h-3 w-1/3 rounded animate-shimmer" />
							<div className="h-[3px] w-full rounded-full animate-shimmer" />
							<div className="flex justify-between">
								<div className="h-3 w-20 rounded animate-shimmer" />
								<div className="h-3 w-16 rounded animate-shimmer" />
							</div>
						</div>
					))}
				</div>
			) : runs.length === 0 ? (
				<div className="flex flex-col items-center justify-center py-20 flex-1 animate-[enterUp_450ms_var(--spring)_both]">
					<div className="mx-auto mb-[22px] flex h-24 w-24 items-center justify-center rounded-full bg-gradient-to-br from-blue/[0.06] to-purple/[0.06] shadow-[0_0_60px_rgba(167,139,250,0.06)] animate-[breathe_4s_ease-in-out_infinite] text-[38px]">
						⚡
					</div>
					<h3 className="text-[19px] font-bold mb-2 tracking-[-0.02em]">
						Нет тестовых прогонов
					</h3>
					<p className="text-[14px] text-t2 mb-6 max-w-[340px] text-center leading-[1.65]">
						{statusFilter
							? "По выбранному фильтру прогоны не найдены. Попробуйте изменить статус."
							: "Создайте первый прогон для проверки тест-кейсов и отслеживания качества продукта"}
					</p>
					{statusFilter ? (
						<Button
							variant="outline"
							onClick={() => {
								setStatusFilter("");
								setOffset(0);
							}}
						>
							Сбросить фильтр
						</Button>
					) : (
						<Button
							className="gradient-tms text-white shadow-md"
							onClick={() => setCreateOpen(true)}
						>
							<svg
								aria-hidden="true"
								className="mr-1.5 h-4 w-4"
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
							Создать прогон
						</Button>
					)}
				</div>
			) : (
				<>
					{/* Сетка карточек прогонов */}
					<div className="grid grid-cols-[repeat(auto-fill,minmax(350px,1fr))] gap-[18px] p-6 flex-1 overflow-y-auto">
						{runs.map((run) => {
							const rate = passRate(run);
							// biome-ignore lint/style/noNonNullAssertion: RUN_STATUS_CONFIG.created is always defined
							const statusCfg = (RUN_STATUS_CONFIG[run.status] ??
								RUN_STATUS_CONFIG.created)!;
							const typeCfg = RUN_TYPE_CONFIG[run.run_type];
							const hasResults = run.total_count > 0;

							return (
								<Link
									key={run.id}
									to="/projects/$projectId/tms/runs/$runId"
									params={{ projectId, runId: run.id }}
									className="block group"
								>
									<div className="tms-run-card bg-gradient-to-b from-card to-raised border border-b1 rounded-[var(--r4)] overflow-hidden transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-[250ms] hover:border-b2 hover:-translate-y-[3px] hover:shadow-[var(--sh2)]">
										{/* Заголовок карточки */}
										<div className="flex items-center gap-3 px-5 py-4">
											<Badge
												variant="secondary"
												className={cn(
													"text-[11px] font-semibold border",
													statusCfg.bg,
													statusCfg.text,
													statusCfg.borderColor,
												)}
											>
												{statusCfg.label}
											</Badge>
											<div className="flex-1" />
											<span className="text-[11.5px] text-t3">
												{hasResults ? `${rate}%` : ""}
											</span>
										</div>

										{/* Название + дата */}
										<div className="px-5 pb-1">
											<div className="text-[14px] font-bold tracking-[-0.01em] mb-[2px] group-hover:text-white transition-colors">
												{run.title}
											</div>
											<div className="text-[11.5px] text-t3">
												{formatRelativeTime(run.created_at)}
												{typeCfg && (
													<span className="ml-2">
														<Badge
															variant="secondary"
															className={cn(
																"text-[10px] font-medium",
																typeCfg.bg,
																typeCfg.text,
															)}
														>
															{typeCfg.label}
														</Badge>
													</span>
												)}
											</div>
										</div>

										{/* Прогресс-бар */}
										{hasResults && (
											<div className="px-5 pb-4 pt-3">
												<div className="h-[3px] bg-white/[0.04] rounded-full overflow-hidden flex gap-[2px]">
													{run.passed_count > 0 && (
														<div
															className="h-full rounded-full bg-green shadow-[0_0_8px_var(--gG)] animate-[progressGrow_800ms_var(--ease)]"
															style={{
																width: `${(run.passed_count / run.total_count) * 100}%`,
															}}
														/>
													)}
													{run.failed_count > 0 && (
														<div
															className="h-full rounded-full bg-red shadow-[0_0_8px_var(--gR)] animate-[progressGrow_800ms_var(--ease)]"
															style={{
																width: `${(run.failed_count / run.total_count) * 100}%`,
															}}
														/>
													)}
													{run.broken_count > 0 && (
														<div
															className="h-full rounded-full bg-orange animate-[progressGrow_800ms_var(--ease)]"
															style={{
																width: `${(run.broken_count / run.total_count) * 100}%`,
															}}
														/>
													)}
													{run.skipped_count > 0 && (
														<div
															className="h-full rounded-full bg-gray opacity-40 animate-[progressGrow_800ms_var(--ease)]"
															style={{
																width: `${(run.skipped_count / run.total_count) * 100}%`,
															}}
														/>
													)}
												</div>
											</div>
										)}

										{/* Строка статистики */}
										{hasResults && (
											<div className="flex gap-3 px-5 pb-4">
												{run.passed_count > 0 && (
													<span className="flex-1 text-center py-[6px] rounded-[var(--r2)] text-[11.5px] font-semibold bg-green/[0.06] text-green">
														{run.passed_count}
													</span>
												)}
												{run.failed_count > 0 && (
													<span className="flex-1 text-center py-[6px] rounded-[var(--r2)] text-[11.5px] font-semibold bg-red/[0.06] text-red">
														{run.failed_count}
													</span>
												)}
												{run.skipped_count > 0 && (
													<span className="flex-1 text-center py-[6px] rounded-[var(--r2)] text-[11.5px] font-semibold bg-white/[0.03] text-t3">
														{run.skipped_count}
													</span>
												)}
											</div>
										)}

										{/* Футер диалога */}
										<div className="px-5 py-3 border-t border-b1 flex items-center justify-between text-[12px] text-t3">
											<span>{run.total_count} тестов</span>
											<span>{formatRelativeTime(run.created_at)}</span>
										</div>
									</div>
								</Link>
							);
						})}
					</div>

					{/* Пагинация */}
					{total > 50 && (
						<div className="px-6 py-3 border-t border-b1 flex items-center justify-between">
							<span className="text-[12px] text-t3 tabular-nums">
								Показано {offset + 1}--{Math.min(offset + 50, total)} из {total}
							</span>
							<div className="flex gap-2">
								<Button
									variant="outline"
									size="sm"
									className="h-7"
									disabled={offset === 0}
									onClick={() => setOffset(Math.max(0, offset - 50))}
								>
									Назад
								</Button>
								<Button
									variant="outline"
									size="sm"
									className="h-7"
									disabled={offset + 50 >= total}
									onClick={() => setOffset(offset + 50)}
								>
									Далее
								</Button>
							</div>
						</div>
					)}
				</>
			)}
		</div>
	);
}

export const Route = createLazyFileRoute(
	"/_authenticated/projects/$projectId/tms/runs",
)({
	component: TestRunsPage,
});
