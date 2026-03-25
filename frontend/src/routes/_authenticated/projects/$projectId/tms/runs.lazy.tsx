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
import { formatRelativeTime } from "@/lib/format-time";
import { cn } from "@/lib/utils";

/** Run status config with icons. */
const RUN_STATUS_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string; iconColor: string }
> = {
	created: {
		bg: "bg-gray-500/10",
		text: "text-gray-600 dark:text-gray-400",
		label: "Создан",
		iconColor: "text-gray-400",
	},
	in_progress: {
		bg: "bg-blue-500/10",
		text: "text-blue-700 dark:text-blue-400",
		label: "Выполняется",
		iconColor: "text-blue-500",
	},
	completed: {
		bg: "bg-green-500/10",
		text: "text-green-700 dark:text-green-400",
		label: "Завершён",
		iconColor: "text-green-500",
	},
	aborted: {
		bg: "bg-red-500/10",
		text: "text-red-600 dark:text-red-400",
		label: "Прерван",
		iconColor: "text-red-500",
	},
};

/** Run type config. */
const RUN_TYPE_CONFIG: Record<
	string,
	{ bg: string; text: string; label: string }
> = {
	manual: {
		bg: "bg-blue-500/10",
		text: "text-blue-700 dark:text-blue-400",
		label: "Ручной",
	},
	automatic: {
		bg: "bg-purple-500/10",
		text: "text-purple-700 dark:text-purple-400",
		label: "Авто",
	},
	mixed: {
		bg: "bg-orange-500/10",
		text: "text-orange-700 dark:text-orange-400",
		label: "Смешанный",
	},
};

/**
 * Test runs list page -- Allure TestOps level design.
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
		onError: () => toast.error("Ошибка создания прогона"),
	});

	const runs = data?.items ?? [];
	const total = data?.total ?? 0;

	function passRate(r: (typeof runs)[number]): number {
		if (r.total_count === 0) return 0;
		return Math.round((r.passed_count / r.total_count) * 100);
	}

	return (
		<div>
			{/* Top bar */}
			<div className="mb-4 flex flex-wrap items-center gap-2">
				<Select
					value={statusFilter || "all"}
					onValueChange={(v) => {
						setStatusFilter(v === "all" ? "" : v);
						setOffset(0);
					}}
				>
					<SelectTrigger className="rounded-xl border bg-card px-2.5 py-1.5 text-xs h-8 w-auto min-w-[130px] shadow-sm cursor-pointer hover:border-primary/40 transition-colors">
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
						className="h-8 text-xs"
						onClick={() => {
							setStatusFilter("");
							setOffset(0);
						}}
					>
						<svg
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

				<div className="ml-auto flex items-center gap-2">
					<span className="text-xs text-muted-foreground tabular-nums">
						{total}{" "}
						{total === 1 ? "прогон" : total < 5 ? "прогона" : "прогонов"}
					</span>
					<Dialog open={createOpen} onOpenChange={setCreateOpen}>
						<DialogTrigger asChild>
							<Button
								size="sm"
								className="gradient-tms text-white shadow-sm h-8"
							>
								<svg
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
			</div>

			{/* Run list */}
			{isLoading ? (
				<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
					{[1, 2, 3, 4, 5, 6].map((i) => (
						<div
							key={i}
							className="rounded-xl border p-4 space-y-3 shadow-card"
						>
							<div className="flex justify-between">
								<div className="h-5 w-3/4 rounded animate-shimmer" />
								<div className="h-5 w-16 rounded-full animate-shimmer" />
							</div>
							<div className="h-3 w-1/3 rounded animate-shimmer" />
							<div className="h-2.5 w-full rounded-full animate-shimmer" />
							<div className="flex justify-between">
								<div className="h-3 w-20 rounded animate-shimmer" />
								<div className="h-3 w-16 rounded animate-shimmer" />
							</div>
						</div>
					))}
				</div>
			) : runs.length === 0 ? (
				<div className="flex flex-col items-center justify-center py-24 animate-fade-in">
					<div className="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-3xl bg-gradient-to-br from-muted/60 to-muted/30 border border-dashed border-muted-foreground/20 shadow-sm">
						<svg
							className="h-12 w-12 text-muted-foreground/30"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M13 10V3L4 14h7v7l9-11h-7z"
							/>
						</svg>
					</div>
					<h3 className="text-xl font-semibold mb-2">Нет тестовых прогонов</h3>
					<p className="text-sm text-muted-foreground mb-6 max-w-sm text-center leading-relaxed">
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
					{/* Run cards */}
					<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
						{runs.map((run) => {
							const rate = passRate(run);
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
									<div className="rounded-xl border bg-card/80 backdrop-blur-sm p-4 shadow-card hover-lift transition-all duration-200 space-y-3 hover:border-primary/20 hover:shadow-lg">
										{/* Title + status */}
										<div className="flex items-start justify-between gap-2">
											<h3 className="text-sm font-semibold leading-tight line-clamp-2 group-hover:text-primary transition-colors">
												{run.title}
											</h3>
											<Badge
												variant="secondary"
												className={cn(
													"text-[10px] font-semibold shrink-0",
													statusCfg.bg,
													statusCfg.text,
												)}
											>
												{statusCfg.label}
											</Badge>
										</div>

										{/* Meta: type + date */}
										<div className="flex items-center gap-2">
											{typeCfg && (
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
											)}
											<span className="text-[11px] text-muted-foreground">
												{formatRelativeTime(run.created_at)}
											</span>
										</div>

										{/* Multi-segment progress bar */}
										{hasResults ? (
											<div className="space-y-2">
												<div className="flex h-2.5 overflow-hidden rounded-full bg-muted/60">
													{run.passed_count > 0 && (
														<div
															className="h-full bg-green-500 animate-progress-fill transition-all duration-500"
															style={{
																width: `${Math.max((run.passed_count / run.total_count) * 100, 4)}%`,
															}}
															title={`Passed: ${run.passed_count}`}
														/>
													)}
													{run.failed_count > 0 && (
														<div
															className="h-full bg-red-500 transition-all duration-500"
															style={{
																width: `${Math.max((run.failed_count / run.total_count) * 100, 4)}%`,
															}}
															title={`Failed: ${run.failed_count}`}
														/>
													)}
													{run.broken_count > 0 && (
														<div
															className="h-full bg-orange-500 transition-all duration-500"
															style={{
																width: `${Math.max((run.broken_count / run.total_count) * 100, 4)}%`,
															}}
															title={`Broken: ${run.broken_count}`}
														/>
													)}
													{run.skipped_count > 0 && (
														<div
															className="h-full bg-gray-400 transition-all duration-500"
															style={{
																width: `${Math.max((run.skipped_count / run.total_count) * 100, 4)}%`,
															}}
															title={`Skipped: ${run.skipped_count}`}
														/>
													)}
												</div>

												{/* Stats as colored chips */}
												<div className="flex items-center gap-1.5 flex-wrap">
													{run.passed_count > 0 && (
														<span className="flex items-center gap-1 text-[10px] font-medium text-green-600 dark:text-green-400">
															<span className="h-1.5 w-1.5 rounded-full bg-green-500" />
															{run.passed_count}
														</span>
													)}
													{run.failed_count > 0 && (
														<span className="flex items-center gap-1 text-[10px] font-medium text-red-600 dark:text-red-400">
															<span className="h-1.5 w-1.5 rounded-full bg-red-500" />
															{run.failed_count}
														</span>
													)}
													{run.broken_count > 0 && (
														<span className="flex items-center gap-1 text-[10px] font-medium text-orange-600 dark:text-orange-400">
															<span className="h-1.5 w-1.5 rounded-full bg-orange-500" />
															{run.broken_count}
														</span>
													)}
													{run.skipped_count > 0 && (
														<span className="flex items-center gap-1 text-[10px] font-medium text-gray-500">
															<span className="h-1.5 w-1.5 rounded-full bg-gray-400" />
															{run.skipped_count}
														</span>
													)}
													<span
														className={cn(
															"ml-auto text-[11px] tabular-nums font-bold",
															rate >= 80
																? "text-green-600 dark:text-green-400"
																: rate >= 50
																	? "text-yellow-600 dark:text-yellow-400"
																	: "text-red-600 dark:text-red-400",
														)}
													>
														{rate}%
													</span>
												</div>
											</div>
										) : (
											<div className="flex items-center gap-2 py-1">
												<div className="h-2.5 w-full rounded-full bg-muted" />
											</div>
										)}
									</div>
								</Link>
							);
						})}
					</div>

					{/* Pagination */}
					{total > 50 && (
						<div className="mt-6 flex items-center justify-between">
							<span className="text-xs text-muted-foreground tabular-nums">
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
