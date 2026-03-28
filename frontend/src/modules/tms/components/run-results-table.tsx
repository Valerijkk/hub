/**
 * Таблица результатов прогона тестов с фильтрацией по статусу.
 */
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

/** Конфигурация статусов результата -- цвета, лейблы и цвет точки-индикатора. */
export const RESULT_STATUS_CONFIG: Record<
	string,
	{ badgeClass: string; label: string; dotColor: string }
> = {
	passed: {
		badgeClass: "badge-passed",
		label: "Passed",
		dotColor: "bg-green-500",
	},
	failed: {
		badgeClass: "badge-failed",
		label: "Failed",
		dotColor: "bg-red-500",
	},
	broken: {
		badgeClass: "badge-broken",
		label: "Broken",
		dotColor: "bg-orange-500",
	},
	skipped: {
		badgeClass: "badge-skipped",
		label: "Skipped",
		dotColor: "bg-gray-400",
	},
	unknown: {
		badgeClass: "badge-blocked",
		label: "Unknown",
		dotColor: "bg-gray-300",
	},
};

/**
 * Форматирует длительность в человекочитаемый формат.
 */
export function formatDuration(ms?: number): string {
	if (!ms) return "--";
	if (ms < 1000) return `${ms}мс`;
	if (ms < 60000) return `${(ms / 1000).toFixed(1)}с`;
	return `${Math.floor(ms / 60000)}м ${Math.floor((ms % 60000) / 1000)}с`;
}

/** Строка результата прогона для таблицы. */
interface ResultRow {
	id: string;
	status: string;
	duration_ms?: number;
	test_case?: { case_key: string; title: string } | null;
	executed_by?: { name: string } | null;
}

/** Пропсы таблицы результатов прогона. */
interface RunResultsTableProps {
	results: ResultRow[];
	runStatus: string;
	executingResultId: string | null;
	onSelectResult: (id: string) => void;
	onAddCases: () => void;
}

/** Таблица результатов прогона тестов с фильтрацией по статусу и подсветкой текущего. */
export function RunResultsTable({
	results,
	runStatus,
	executingResultId,
	onSelectResult,
	onAddCases,
}: RunResultsTableProps) {
	const [statusFilter, setStatusFilter] = useState("");

	if (results.length === 0) {
		return (
			<div className="flex flex-col items-center justify-center py-16 text-center border rounded-xl border-dashed">
				<svg
					className="h-12 w-12 text-muted-foreground/20 mb-3"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={1.5}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
					/>
				</svg>
				<p className="text-sm text-muted-foreground mb-3">
					Нет результатов. Добавьте тест-кейсы в прогон.
				</p>
				{runStatus === "created" && (
					<Button variant="outline" size="sm" onClick={onAddCases}>
						Добавить тест-кейсы
					</Button>
				)}
			</div>
		);
	}

	const filteredResults = statusFilter
		? results.filter((r) => r.status === statusFilter)
		: results;

	return (
		<>
			{/* Фильтр по статусу результатов */}
			<div className="mb-3 flex items-center gap-2">
				<Select
					value={statusFilter || "all"}
					onValueChange={(v) => setStatusFilter(v === "all" ? "" : v)}
				>
					<SelectTrigger className="rounded-xl border bg-card px-2.5 py-1.5 text-xs h-8 w-auto min-w-[140px] shadow-sm cursor-pointer hover:border-primary/40 transition-colors">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">Все результаты</SelectItem>
						<SelectItem value="passed">Passed</SelectItem>
						<SelectItem value="failed">Failed</SelectItem>
						<SelectItem value="broken">Broken</SelectItem>
						<SelectItem value="skipped">Skipped</SelectItem>
						<SelectItem value="unknown">Unknown</SelectItem>
					</SelectContent>
				</Select>
				{statusFilter && (
					<Button
						variant="ghost"
						size="sm"
						className="h-8 text-xs"
						onClick={() => setStatusFilter("")}
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
				<span className="ml-auto text-[11px] text-muted-foreground tabular-nums">
					{statusFilter
						? `${filteredResults.length} из ${results.length}`
						: `${results.length} результатов`}
				</span>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 overflow-hidden">
				{filteredResults.length === 0 && statusFilter ? (
					<div className="flex flex-col items-center justify-center py-14 text-center">
						<svg
							className="h-10 w-10 text-muted-foreground/20 mb-3"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"
							/>
						</svg>
						<p className="text-sm font-medium text-muted-foreground mb-1">
							Нет результатов с выбранным статусом
						</p>
						<p className="text-xs text-muted-foreground/60 mb-3">
							Статус "
							{
								(RESULT_STATUS_CONFIG[statusFilter] ?? { label: statusFilter })
									.label
							}
							" не найден среди результатов
						</p>
						<Button
							variant="outline"
							size="sm"
							onClick={() => setStatusFilter("")}
						>
							Показать все результаты
						</Button>
					</div>
				) : (
					<table className="w-full border-separate border-spacing-0">
						<thead>
							<tr>
								<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2">
									Тест-кейс
								</th>
								<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2 w-24">
									Статус
								</th>
								<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2 w-24">
									Время
								</th>
								<th className="sticky top-0 z-[5] bg-base text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-5 py-[14px] border-b border-b2 w-28">
									Испол��итель
								</th>
							</tr>
						</thead>
						<tbody>
							{filteredResults.map((r) => {
								const resultCfg = (RESULT_STATUS_CONFIG[r.status] ??
									RESULT_STATUS_CONFIG.unknown)!;
								const isExecuting = executingResultId === r.id;
								const isClickable = runStatus === "in_progress";

								return (
									<tr
										key={r.id}
										className={cn(
											"group transition-all duration-[180ms]",
											isClickable && "cursor-pointer",
											isExecuting && "bg-purple/[0.05]",
										)}
										onClick={() => {
											if (isClickable) onSelectResult(r.id);
										}}
									>
										<td className="px-5 h-14 border-b border-b1 align-middle transition-all duration-[180ms] group-hover:bg-[var(--hover)]">
											<div className="flex items-center gap-2">
												<span
													className={cn(
														"h-2 w-2 rounded-[3px] shrink-0 transition-[box-shadow,transform] duration-200 group-hover:shadow-[0_0_8px_currentColor] group-hover:scale-[1.2]",
														resultCfg.dotColor,
													)}
													style={{ color: "inherit" }}
												/>
												<div className="min-w-0">
													<span className="font-mono text-[12px] font-medium text-t3">
														{r.test_case?.case_key ?? "N/A"}
													</span>
													<div className="text-[13px] text-t2 truncate max-w-xs group-hover:text-t1 transition-colors duration-150">
														{r.test_case?.title}
													</div>
												</div>
											</div>
										</td>
										<td className="px-5 h-14 border-b border-b1 align-middle transition-all duration-[180ms] group-hover:bg-[var(--hover)]">
											<Badge
												variant="secondary"
												className={cn(
													"text-[11px] font-semibold",
													resultCfg.badgeClass,
												)}
											>
												{resultCfg.label}
											</Badge>
										</td>
										<td className="px-5 h-14 border-b border-b1 align-middle transition-all duration-[180ms] group-hover:bg-[var(--hover)]">
											<span className="text-[12px] text-t3 tabular-nums">
												{formatDuration(r.duration_ms)}
											</span>
										</td>
										<td className="px-5 h-14 border-b border-b1 align-middle transition-all duration-[180ms] group-hover:bg-[var(--hover)]">
											<span className="text-[12px] text-t3">
												{r.executed_by?.name ?? "--"}
											</span>
										</td>
									</tr>
								);
							})}
						</tbody>
					</table>
				)}
			</div>
		</>
	);
}
