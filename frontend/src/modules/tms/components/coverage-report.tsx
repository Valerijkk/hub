import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { getCoverageReport, type IssueCoverage } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

/** Цвета статусов результатов для покрытия. */
const COVERAGE_STATUS_COLORS: Record<string, string> = {
	passed: "bg-green-100 text-green-800",
	failed: "bg-red-100 text-red-800",
	broken: "bg-orange-100 text-orange-800",
	skipped: "bg-gray-200 text-gray-700",
	unknown: "bg-gray-100 text-gray-500",
	no_results: "bg-yellow-100 text-yellow-700",
};

/** Лейблы статусов. */
const COVERAGE_STATUS_LABELS: Record<string, string> = {
	passed: "Passed",
	failed: "Failed",
	broken: "Broken",
	skipped: "Skipped",
	unknown: "Unknown",
	no_results: "Нет результатов",
};

/**
 * Строка покрытия задачи с раскрывающимися тест-кейсами.
 */
function IssueCoverageRow({ issue }: { issue: IssueCoverage }) {
	const [expanded, setExpanded] = useState(false);

	const passedCount = issue.test_cases.filter(
		(tc) => tc.last_status === "passed",
	).length;
	const totalCount = issue.test_cases.length;
	const passRate =
		totalCount > 0 ? Math.round((passedCount / totalCount) * 100) : 0;

	return (
		<>
			<tr
				className="cursor-pointer border-b transition-colors hover:bg-raised/50"
				onClick={() => setExpanded(!expanded)}
			>
				<td className="px-4 py-3">
					<span className="font-medium text-primary">{issue.issue_key}</span>
				</td>
				<td className="px-4 py-3 text-sm">{issue.issue_title}</td>
				<td className="px-4 py-3 text-center text-sm">{totalCount}</td>
				<td className="px-4 py-3">
					<div className="flex items-center gap-2">
						<div className="h-2 w-20 overflow-hidden rounded-full bg-raised">
							<div
								className={cn(
									"h-full rounded-full transition-[width]",
									passRate === 100
										? "bg-green-500"
										: passRate >= 50
											? "bg-yellow-500"
											: "bg-red-500",
								)}
								style={{ width: `${passRate}%` }}
							/>
						</div>
						<span className="text-xs text-t2">{passRate}%</span>
					</div>
				</td>
				<td className="px-4 py-3 text-right text-xs text-t2">
					{expanded ? "Свернуть" : "Развернуть"}
				</td>
			</tr>
			{expanded &&
				issue.test_cases.map((tc) => (
					<tr key={tc.case_key} className="border-b bg-raised/20">
						<td className="px-8 py-2 text-xs text-t2">
							{tc.case_key}
						</td>
						<td className="px-4 py-2 text-xs">{tc.case_title}</td>
						<td className="px-4 py-2" />
						<td className="px-4 py-2" colSpan={2}>
							<Badge
								variant="secondary"
								className={cn(
									"text-xs",
									COVERAGE_STATUS_COLORS[tc.last_status] ?? "bg-gray-100",
								)}
							>
								{COVERAGE_STATUS_LABELS[tc.last_status] ?? tc.last_status}
							</Badge>
						</td>
					</tr>
				))}
		</>
	);
}

/**
 * Отчёт о покрытии задач тестами.
 * Таблица: Задача -> Кол-во тестов -> Процент покрытия.
 * Раскрываемые строки показывают отдельные тест-кейсы с последним статусом.
 */
/** Отчёт о покрытии задач тестами с раскрывающимися строками и прогресс-барами. */
export function CoverageReport({ projectId }: { projectId: string }) {
	const { data, isLoading } = useQuery({
		queryKey: ["tms-coverage", projectId],
		queryFn: () => getCoverageReport(projectId),
	});

	if (isLoading) {
		return (
			<div className="py-8 text-center text-t2">
				Загрузка отчёта...
			</div>
		);
	}

	if (!data || data.issues.length === 0) {
		return (
			<Card>
				<CardContent className="py-8 text-center text-t2">
					Нет задач с привязанными тест-кейсами.
					<br />
					Привяжите тест-кейсы к задачам для отображения покрытия.
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm">
					Покрытие задач тестами ({data.issues.length} задач)
				</CardTitle>
			</CardHeader>
			<CardContent className="p-0">
				<div className="rounded-md border">
					<table className="w-full">
						<thead>
							<tr className="border-b bg-raised/50">
								<th className="px-4 py-3 text-left text-xs font-medium text-t2">
									Задача
								</th>
								<th className="px-4 py-3 text-left text-xs font-medium text-t2">
									Название
								</th>
								<th className="px-4 py-3 text-center text-xs font-medium text-t2">
									Тестов
								</th>
								<th className="px-4 py-3 text-left text-xs font-medium text-t2">
									Покрытие
								</th>
								<th className="px-4 py-3" />
							</tr>
						</thead>
						<tbody>
							{data.issues.map((issue) => (
								<IssueCoverageRow key={issue.issue_id} issue={issue} />
							))}
						</tbody>
					</table>
				</div>
			</CardContent>
		</Card>
	);
}
