/**
 * Панель выполнения тест-кейса -- результат, комментарий, кнопки статусов.
 */

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import {
	type StepResult,
	type TestResultDetailResponse,
	type TestResultStatus,
	type UpdateTestResultRequest,
	updateTestResult,
} from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { toastApiError } from "@/lib/format-api-error";
import { cn } from "@/lib/utils";

/** Конфигурация кнопок статусов выполнения (passed, failed, broken, skipped). */
const STATUS_BUTTONS: {
	status: TestResultStatus;
	label: string;
	color: string;
	activeColor: string;
}[] = [
	{
		status: "passed",
		label: "Passed",
		color:
			"hover:bg-green-500/10 hover:text-green-700 hover:border-green-500/30",
		activeColor:
			"bg-green-500/15 text-green-700 dark:text-green-400 border-green-500/30",
	},
	{
		status: "failed",
		label: "Failed",
		color: "hover:bg-red-500/10 hover:text-red-700 hover:border-red-500/30",
		activeColor:
			"bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/30",
	},
	{
		status: "broken",
		label: "Broken",
		color:
			"hover:bg-orange-500/10 hover:text-orange-700 hover:border-orange-500/30",
		activeColor:
			"bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/30",
	},
	{
		status: "skipped",
		label: "Skip",
		color: "hover:bg-gray-500/10 hover:text-gray-700 hover:border-gray-500/30",
		activeColor:
			"bg-gray-500/15 text-gray-700 dark:text-gray-400 border-gray-500/30",
	},
];

/** Панель выполнения тест-кейса: выбор статуса, комментарий, сохранение результата. */
export function RunExecutionPanel({
	result,
	projectId,
	runId,
	onClose,
	onNext,
}: {
	result: TestResultDetailResponse;
	projectId: string;
	runId: string;
	onClose: () => void;
	onNext: () => void;
}) {
	const queryClient = useQueryClient();
	const tc = result.test_case;
	const [comment, setComment] = useState(result.comment ?? "");
	const [overallStatus, setOverallStatus] = useState<TestResultStatus>(
		result.status === "unknown" ? "passed" : result.status,
	);
	const [stepResults] = useState<StepResult[]>(() => {
		if (result.step_results && result.step_results.length > 0) {
			return result.step_results;
		}
		return [];
	});

	const submitMutation = useMutation({
		mutationFn: () => {
			const data: UpdateTestResultRequest = {
				status: overallStatus,
				comment: comment || undefined,
				step_results: stepResults.length > 0 ? stepResults : undefined,
			};
			return updateTestResult(projectId, runId, result.id, data);
		},
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: ["tms-run-detail", projectId, runId],
			});
			toast.success("Результат сохранён");
			onNext();
		},
		onError: (err) => toastApiError(err, "Ошибка сохранения результата"),
	});

	return (
		<div className="space-y-4">
			{/* Информация о тест-кейсе */}
			{tc && (
				<div className="rounded-lg border bg-raised/20 p-3">
					<div className="flex items-center gap-2 mb-1">
						<span className="font-mono text-xs font-bold text-primary">
							{tc.case_key}
						</span>
						<Badge variant="outline" className="text-[10px]">
							{tc.case_type === "bdd" ? "BDD" : "Classic"}
						</Badge>
					</div>
					<div className="text-sm font-medium leading-tight">{tc.title}</div>
				</div>
			)}

			{/* Кнопки статусов */}
			<div>
				<div className="mb-2 text-xs font-semibold text-t2 uppercase tracking-wider">
					Результат
				</div>
				<div className="grid grid-cols-4 gap-1.5">
					{STATUS_BUTTONS.map((btn) => (
						<button
							key={btn.status}
							type="button"
							className={cn(
								"rounded-lg border px-3 py-2 text-xs font-semibold transition-colors duration-150",
								overallStatus === btn.status
									? btn.activeColor
									: cn("border-border text-t2", btn.color),
							)}
							onClick={() => setOverallStatus(btn.status)}
						>
							{btn.label}
						</button>
					))}
				</div>
			</div>

			{/* Комментарий */}
			<div>
				<div className="mb-1.5 text-xs font-semibold text-t2 uppercase tracking-wider">
					Комментарий
				</div>
				<textarea
					className="w-full rounded-lg border bg-card px-3 py-2 text-sm"
					rows={3}
					value={comment}
					onChange={(e) => setComment(e.target.value)}
					placeholder="Комментарий к результату..."
				/>
			</div>

			{/* Кнопки управления */}
			<div className="flex gap-2">
				<Button
					className="flex-1 gradient-tms text-white"
					disabled={submitMutation.isPending}
					onClick={() => submitMutation.mutate()}
				>
					{submitMutation.isPending ? "Сохранение..." : "Сохранить"}
				</Button>
				<Button variant="outline" onClick={onClose} className="shrink-0">
					Закрыть
				</Button>
			</div>
		</div>
	);
}
