import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import {
	addCasesToRun,
	completeRun,
	getTestRunDetail,
	listTestCases,
	reopenRun,
	startRun,
} from "@/api/tms";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import { RunExecutionPanel } from "@/modules/tms/components/run-execution-panel";
import { RunHeader } from "@/modules/tms/components/run-header";
import { RunResultsTable } from "@/modules/tms/components/run-results-table";

/**
 * Страница деталей прогона тестов -- дизайн уровня Allure TestOps.
 */
function TestRunDetailPage() {
	const { projectId, runId } = Route.useParams();
	const queryClient = useQueryClient();
	const [executingResultId, setExecutingResultId] = useState<string | null>(
		null,
	);
	const [addCasesOpen, setAddCasesOpen] = useState(false);
	const [selectedCaseIds, setSelectedCaseIds] = useState<string[]>([]);

	// Загрузка деталей прогона
	const { data: run, isLoading } = useQuery({
		queryKey: ["tms-run-detail", projectId, runId],
		queryFn: () => getTestRunDetail(projectId, runId),
	});

	// Загрузка тест-кейсов для диалога добавления
	const { data: casesData } = useQuery({
		queryKey: ["tms-cases-for-run", projectId],
		queryFn: () => listTestCases(projectId, { limit: 100 }),
		enabled: addCasesOpen,
	});

	// Мутации управления прогоном
	const startMutation = useMutation({
		mutationFn: () => startRun(projectId, runId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-run-detail", projectId, runId],
			});
			toast.success("Прогон запущен");
		},
		onError: () => toast.error("Ошибка запуска прогона"),
	});

	const completeMutation = useMutation({
		mutationFn: () => completeRun(projectId, runId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-run-detail", projectId, runId],
			});
			toast.success("Прогон завершён");
		},
		onError: () => toast.error("Ошибка завершения прогона"),
	});

	const reopenMutation = useMutation({
		mutationFn: () => reopenRun(projectId, runId),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-run-detail", projectId, runId],
			});
			toast.success("Прогон переоткрыт");
		},
		onError: () => toast.error("Ошибка переоткрытия прогона"),
	});

	const addCasesMutation = useMutation({
		mutationFn: () => addCasesToRun(projectId, runId, selectedCaseIds),
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-run-detail", projectId, runId],
			});
			setAddCasesOpen(false);
			setSelectedCaseIds([]);
			toast.success("Тест-кейсы добавлены");
		},
		onError: () => toast.error("Ошибка добавления тест-кейсов"),
	});

	/** Переходит к следующему невыполненному результату. */
	const handleNext = useCallback(() => {
		if (!run) {
			setExecutingResultId(null);
			return;
		}
		const nextUnknown = (run.results ?? []).find(
			(r) => r.status === "unknown" && r.id !== executingResultId,
		);
		if (nextUnknown) {
			setExecutingResultId(nextUnknown.id);
		} else {
			setExecutingResultId(null);
		}
	}, [run, executingResultId]);

	/** Переключает выбор тест-кейса для добавления. */
	const toggleCaseSelection = useCallback((caseId: string) => {
		setSelectedCaseIds((prev) =>
			prev.includes(caseId)
				? prev.filter((id) => id !== caseId)
				: [...prev, caseId],
		);
	}, []);

	if (isLoading || !run) {
		return (
			<div className="mx-auto max-w-6xl animate-fade-in">
				<div className="mb-4 flex items-center gap-2">
					<div className="h-4 w-16 rounded animate-shimmer" />
					<div className="h-4 w-4 rounded animate-shimmer" />
					<div className="h-4 w-40 rounded animate-shimmer" />
				</div>
				<div className="h-8 w-64 rounded animate-shimmer mb-4" />
				<div className="grid grid-cols-4 gap-3 mb-6">
					{[1, 2, 3, 4].map((i) => (
						<div key={i} className="h-20 rounded-xl animate-shimmer" />
					))}
				</div>
				<div className="h-64 w-full rounded-xl animate-shimmer" />
			</div>
		);
	}

	const executingResult = executingResultId
		? (run.results ?? []).find((r) => r.id === executingResultId)
		: null;

	/** Конфигурация карточек статистики. */
	const statCards = [
		{
			label: "Passed",
			count: run.passed_count,
			color: "text-green-600 dark:text-green-400",
			bgColor: "bg-green-500/10",
			borderColor: "border-green-500/20",
		},
		{
			label: "Failed",
			count: run.failed_count,
			color: "text-red-600 dark:text-red-400",
			bgColor: "bg-red-500/10",
			borderColor: "border-red-500/20",
		},
		{
			label: "Broken",
			count: run.broken_count,
			color: "text-orange-600 dark:text-orange-400",
			bgColor: "bg-orange-500/10",
			borderColor: "border-orange-500/20",
		},
		{
			label: "Skipped",
			count: run.skipped_count,
			color: "text-gray-500",
			bgColor: "bg-gray-500/10",
			borderColor: "border-gray-500/20",
		},
	];

	return (
		<div className="mx-auto max-w-6xl animate-fade-in">
			<RunHeader
				projectId={projectId}
				run={run}
				onStart={() => startMutation.mutate()}
				onComplete={() => completeMutation.mutate()}
				onReopen={() => reopenMutation.mutate()}
				onAddCases={() => setAddCasesOpen(true)}
				isStarting={startMutation.isPending}
				isCompleting={completeMutation.isPending}
				isReopening={reopenMutation.isPending}
			/>

			{/* Карточки статистики -- 4 больших числа */}
			<div className="mb-6 grid grid-cols-2 sm:grid-cols-4 gap-3">
				{statCards.map((card) => (
					<div
						key={card.label}
						className={cn(
							"rounded-xl border p-5 text-center transition-all duration-200 shadow-sm hover:shadow-md hover:-translate-y-0.5",
							card.bgColor,
							card.borderColor,
						)}
					>
						<div
							className={cn(
								"text-4xl font-extrabold tabular-nums leading-none tracking-tight",
								card.color,
							)}
						>
							{card.count}
						</div>
						<div className="text-[11px] font-semibold text-muted-foreground mt-2 uppercase tracking-wider">
							{card.label}
						</div>
						{run.total_count > 0 && (
							<div className="text-[10px] text-muted-foreground/60 tabular-nums mt-1">
								{Math.round((card.count / run.total_count) * 100)}%
							</div>
						)}
					</div>
				))}
			</div>

			{/* Основной контент: таблица + панель выполнения */}
			<div className="flex gap-4">
				{/* Таблица результатов */}
				<div className="flex-1 min-w-0">
					<RunResultsTable
						results={run.results ?? []}
						runStatus={run.status}
						executingResultId={executingResultId}
						onSelectResult={(id) => setExecutingResultId(id)}
						onAddCases={() => setAddCasesOpen(true)}
					/>
				</div>

				{/* Панель выполнения -- фиксированная справа */}
				{executingResult && run.status === "in_progress" && (
					<div className="w-80 shrink-0">
						<Card className="shadow-card sticky top-4">
							<CardHeader className="pb-2 px-4 pt-4">
								<CardTitle className="text-[11px] font-bold uppercase tracking-wider text-muted-foreground">
									Выполнение теста
								</CardTitle>
							</CardHeader>
							<CardContent className="px-4 pb-4">
								<RunExecutionPanel
									result={executingResult}
									projectId={projectId}
									runId={runId}
									onClose={() => setExecutingResultId(null)}
									onNext={handleNext}
								/>
							</CardContent>
						</Card>
					</div>
				)}
			</div>

			{/* Диалог добавления тест-кейсов */}
			<Dialog open={addCasesOpen} onOpenChange={setAddCasesOpen}>
				<DialogContent className="max-h-[80vh] overflow-y-auto">
					<DialogHeader>
						<DialogTitle>Добавить тест-кейсы в прогон</DialogTitle>
					</DialogHeader>
					<div className="space-y-1.5">
						{(casesData?.items ?? []).map((tc) => {
							const existingIds = new Set(
								(run.results ?? []).map((r) => r.test_case?.id).filter(Boolean),
							);
							const alreadyAdded = existingIds.has(tc.id);
							const isSelected = selectedCaseIds.includes(tc.id);
							return (
								<label
									key={tc.id}
									className={cn(
										"flex items-center gap-3 rounded-lg border p-2.5 text-sm transition-colors cursor-pointer",
										alreadyAdded && "opacity-40 cursor-not-allowed",
										isSelected &&
											!alreadyAdded &&
											"bg-primary/5 border-primary/30",
										!alreadyAdded && !isSelected && "hover:bg-muted/50",
									)}
								>
									<input
										type="checkbox"
										checked={isSelected}
										disabled={alreadyAdded}
										onChange={() => toggleCaseSelection(tc.id)}
										className="rounded"
									/>
									<div className="min-w-0 flex-1">
										<span className="font-mono text-xs font-bold text-primary">
											{tc.case_key}
										</span>
										<span className="ml-2 text-sm">{tc.title}</span>
										{alreadyAdded && (
											<span className="ml-2 text-[10px] text-muted-foreground">
												(уже добавлен)
											</span>
										)}
									</div>
								</label>
							);
						})}
					</div>
					<Button
						className="mt-4 w-full gradient-tms text-white"
						disabled={
							selectedCaseIds.length === 0 || addCasesMutation.isPending
						}
						onClick={() => addCasesMutation.mutate()}
					>
						{addCasesMutation.isPending
							? "Добавление..."
							: `Добавить (${selectedCaseIds.length})`}
					</Button>
				</DialogContent>
			</Dialog>
		</div>
	);
}

export const Route = createFileRoute(
	"/_authenticated/projects/$projectId/tms/runs_/$runId",
)({
	component: TestRunDetailPage,
});
