import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { getTraceabilityMatrix } from "@/api/tms";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Конфигурация статусов
// ---------------------------------------------------------------------------

const STATUS_CONFIG: Record<
	string,
	{ bg: string; text: string; cell: string; label: string }
> = {
	passed: {
		bg: "bg-green-500/10",
		text: "text-green-700 dark:text-green-400",
		cell: "bg-green-100 dark:bg-green-900/30",
		label: "Passed",
	},
	failed: {
		bg: "bg-red-500/10",
		text: "text-red-600 dark:text-red-400",
		cell: "bg-red-100 dark:bg-red-900/30",
		label: "Failed",
	},
	broken: {
		bg: "bg-orange-500/10",
		text: "text-orange-600 dark:text-orange-400",
		cell: "bg-orange-100 dark:bg-orange-900/30",
		label: "Broken",
	},
	skipped: {
		bg: "bg-yellow-500/10",
		text: "text-yellow-700 dark:text-yellow-400",
		cell: "bg-yellow-100 dark:bg-yellow-900/30",
		label: "Skipped",
	},
	blocked: {
		bg: "bg-gray-500/10",
		text: "text-gray-600 dark:text-gray-400",
		cell: "bg-gray-200 dark:bg-gray-700/40",
		label: "Blocked",
	},
	unknown: {
		bg: "bg-gray-500/10",
		text: "text-gray-500",
		cell: "bg-gray-100 dark:bg-gray-800/30",
		label: "Unknown",
	},
	no_results: {
		bg: "bg-gray-500/10",
		text: "text-gray-400",
		cell: "bg-gray-50 dark:bg-gray-800/20",
		label: "Нет результатов",
	},
};

function getStatusConfig(status: string) {
	return STATUS_CONFIG[status] ?? STATUS_CONFIG.unknown!;
}

// ---------------------------------------------------------------------------
// Компонент ячейки матрицы
// ---------------------------------------------------------------------------

function MatrixCell({ status }: { status: string }) {
	const cfg = getStatusConfig(status);

	return (
		<td
			className={cn(
				"border px-2 py-1.5 text-center transition-colors",
				cfg.cell,
			)}
			title={cfg.label}
		>
			<span className={cn("text-[10px] font-bold uppercase", cfg.text)}>
				{status === "no_results" ? "--" : status.charAt(0).toUpperCase()}
			</span>
		</td>
	);
}

// ---------------------------------------------------------------------------
// Основной компонент
// ---------------------------------------------------------------------------

/** Пропсы компонента TraceabilityMatrix. */
interface TraceabilityMatrixProps {
	/** ID проекта. */
	projectId: string;
}

/**
 * Матрица трассировки: задачи (строки) x тест-кейсы (столбцы).
 * Ячейки показывают последний статус выполнения теста.
 * Поддерживает горизонтальную/вертикальную прокрутку для больших матриц.
 */
export function TraceabilityMatrix({ projectId }: TraceabilityMatrixProps) {
	const [search, setSearch] = useState("");

	const { data, isLoading } = useQuery({
		queryKey: ["tms-traceability", projectId],
		queryFn: () => getTraceabilityMatrix(projectId),
	});

	/** Уникальные тест-кейсы для заголовков столбцов. */
	const allCaseKeys = useMemo(() => {
		if (!data?.rows) return [];
		const keys = new Set<string>();
		for (const row of data.rows) {
			for (const tc of row.test_cases) {
				keys.add(tc.case_key);
			}
		}
		return Array.from(keys).sort();
	}, [data]);

	/** Карта тест-кейс-ключ -> название (для тултипов). */
	const caseTitleMap = useMemo(() => {
		if (!data?.rows) return new Map<string, string>();
		const map = new Map<string, string>();
		for (const row of data.rows) {
			for (const tc of row.test_cases) {
				if (!map.has(tc.case_key)) {
					map.set(tc.case_key, tc.case_title);
				}
			}
		}
		return map;
	}, [data]);

	/** Отфильтрованные строки. */
	const filteredRows = useMemo(() => {
		if (!data?.rows) return [];
		if (!search) return data.rows;
		const lower = search.toLowerCase();
		return data.rows.filter(
			(row) =>
				row.issue_key.toLowerCase().includes(lower) ||
				row.issue_title.toLowerCase().includes(lower),
		);
	}, [data, search]);

	/** Подсчёт статистики. */
	const stats = useMemo(() => {
		if (!data?.rows) return { total: 0, covered: 0, passed: 0, failed: 0 };
		let total = 0;
		let passed = 0;
		let failed = 0;
		for (const row of data.rows) {
			for (const tc of row.test_cases) {
				total++;
				if (tc.last_status === "passed") passed++;
				if (tc.last_status === "failed" || tc.last_status === "broken")
					failed++;
			}
		}
		return {
			total,
			covered: data.rows.length,
			passed,
			failed,
		};
	}, [data]);

	// Загрузка
	if (isLoading) {
		return (
			<div className="py-8 text-center text-t2">
				Загрузка матрицы трассировки...
			</div>
		);
	}

	// Пусто
	if (!data || data.rows.length === 0) {
		return (
			<Card>
				<CardContent className="py-8 text-center text-t2">
					<div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-raised/60 to-raised/30 border border-dashed border-t2/20">
						<svg
							className="h-8 w-8 text-t2/30"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
							strokeWidth={1.5}
						>
							<path
								strokeLinecap="round"
								strokeLinejoin="round"
								d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z"
							/>
						</svg>
					</div>
					<p className="font-medium mb-1">Нет данных для матрицы трассировки</p>
					<p className="text-sm">
						Привяжите тест-кейсы к задачам для отображения матрицы
					</p>
				</CardContent>
			</Card>
		);
	}

	return (
		<div className="space-y-4">
			{/* Статистика */}
			<div className="flex flex-wrap items-center gap-4">
				<div className="flex items-center gap-2">
					<span className="text-xs text-t2">Задач:</span>
					<span className="text-sm font-semibold tabular-nums">
						{stats.covered}
					</span>
				</div>
				<div className="flex items-center gap-2">
					<span className="text-xs text-t2">Связей:</span>
					<span className="text-sm font-semibold tabular-nums">
						{stats.total}
					</span>
				</div>
				<div className="flex items-center gap-2">
					<span className="h-2 w-2 rounded-full bg-green-500" />
					<span className="text-xs text-t2">
						Passed: {stats.passed}
					</span>
				</div>
				<div className="flex items-center gap-2">
					<span className="h-2 w-2 rounded-full bg-red-500" />
					<span className="text-xs text-t2">
						Failed: {stats.failed}
					</span>
				</div>

				{/* Легенда */}
				<div className="ml-auto flex items-center gap-2">
					{["passed", "failed", "broken", "skipped", "no_results"].map(
						(status) => {
							const cfg = getStatusConfig(status);
							return (
								<div key={status} className="flex items-center gap-1">
									<span
										className={cn(
											"inline-flex h-5 w-5 items-center justify-center rounded text-[9px] font-bold",
											cfg.cell,
											cfg.text,
										)}
									>
										{status === "no_results"
											? "--"
											: status.charAt(0).toUpperCase()}
									</span>
									<span className="text-[10px] text-t2">
										{cfg.label}
									</span>
								</div>
							);
						},
					)}
				</div>
			</div>

			{/* Поиск */}
			<div className="relative">
				<svg
					className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-t2/60"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
					strokeWidth={2}
				>
					<path
						strokeLinecap="round"
						strokeLinejoin="round"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
					/>
				</svg>
				<Input
					placeholder="Поиск по задачам..."
					className="w-64 pl-8 h-8 text-sm"
					value={search}
					onChange={(e) => setSearch(e.target.value)}
				/>
			</div>

			{/* Матрица — прокручиваемый контейнер */}
			<div className="rounded-xl border shadow-card overflow-auto max-h-[calc(100vh-300px)]">
				<table className="w-max min-w-full border-collapse">
					<thead className="sticky top-0 z-10">
						<tr className="bg-raised/60 backdrop-blur-sm">
							<th className="sticky left-0 z-20 bg-raised/90 border-b border-r px-4 py-2.5 text-left text-[11px] font-bold uppercase tracking-wider text-t2 min-w-[200px]">
								Задача
							</th>
							{allCaseKeys.map((key) => (
								<th
									key={key}
									className="border-b px-2 py-2.5 text-center text-[10px] font-bold uppercase tracking-wider text-t2 whitespace-nowrap"
									title={caseTitleMap.get(key) ?? key}
								>
									{key}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{filteredRows.map((row) => {
							// Карта: case_key -> статус для быстрого доступа
							const statusMap = new Map<string, string>();
							for (const tc of row.test_cases) {
								statusMap.set(tc.case_key, tc.last_status);
							}

							return (
								<tr
									key={row.issue_key}
									className="border-b hover:bg-raised/30 transition-colors"
								>
									<td className="sticky left-0 z-10 bg-card/95 border-r px-4 py-2 min-w-[200px]">
										<div className="flex flex-col">
											<span className="font-mono text-xs font-bold text-primary">
												{row.issue_key}
											</span>
											<span className="text-xs text-t2 line-clamp-1">
												{row.issue_title}
											</span>
										</div>
									</td>
									{allCaseKeys.map((key) => {
										const status = statusMap.get(key);
										if (!status) {
											return (
												<td
													key={key}
													className="border px-2 py-1.5 text-center"
												>
													<span className="text-[10px] text-t2/20">
														--
													</span>
												</td>
											);
										}
										return <MatrixCell key={key} status={status} />;
									})}
								</tr>
							);
						})}
					</tbody>
				</table>
			</div>

			{/* Количество строк */}
			<div className="text-xs text-t2">
				Показано {filteredRows.length} из {data.rows.length} задач
			</div>
		</div>
	);
}
