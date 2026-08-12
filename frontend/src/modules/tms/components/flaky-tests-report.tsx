import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { getFlakyTests } from "@/api/tms";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

function statusHistoryEntries(
	itemId: string,
	history: string[],
): { key: string; status: string }[] {
	const counts = new Map<string, number>();
	return history.map((status) => {
		const n = (counts.get(status) ?? 0) + 1;
		counts.set(status, n);
		return { key: `${itemId}-${status}-${n}`, status };
	});
}

const STATUS_COLORS: Record<string, string> = {
	passed: "bg-green-500/10 text-green-700 dark:text-green-400",
	failed: "bg-red-500/10 text-red-600 dark:text-red-400",
	broken: "bg-orange-500/10 text-orange-600 dark:text-orange-400",
	skipped: "bg-yellow-500/10 text-yellow-700 dark:text-yellow-400",
};

interface FlakyTestsReportProps {
	projectId: string;
}

/** PROD-02: flaky-тесты по последним N прогонам. */
export function FlakyTestsReport({ projectId }: FlakyTestsReportProps) {
	const { data, isLoading, isError } = useQuery({
		queryKey: ["tms-flaky", projectId],
		queryFn: () => getFlakyTests(projectId),
	});

	if (isLoading) {
		return <p className="text-sm text-t3">Загрузка отчёта…</p>;
	}

	if (isError) {
		return (
			<p className="text-sm text-red-500">Не удалось загрузить flaky-отчёт</p>
		);
	}

	const items = data?.items ?? [];
	const runWindow = data?.run_window ?? 5;

	return (
		<div className="space-y-4">
			<p className="text-sm text-t2">
				Нестабильные тесты за последние {runWindow} прогонов: статус менялся
				между запусками или отмечен как flaky.
			</p>

			{items.length === 0 ? (
				<Card>
					<CardContent className="py-8 text-center text-sm text-t3">
						Flaky-тесты не обнаружены
					</CardContent>
				</Card>
			) : (
				<div className="overflow-x-auto rounded-lg border border-b1">
					<table className="w-full border-collapse text-sm">
						<thead>
							<tr className="border-b border-b1 bg-[var(--raised)] text-left">
								<th className="px-4 py-3 font-medium text-t2">Ключ</th>
								<th className="px-4 py-3 font-medium text-t2">Название</th>
								<th className="px-4 py-3 font-medium text-t2">Flake rate</th>
								<th className="px-4 py-3 font-medium text-t2">История</th>
							</tr>
						</thead>
						<tbody>
							{items.map((item) => (
								<tr
									key={item.id}
									className="border-b border-b1 hover:bg-[var(--hover)]"
								>
									<td className="px-4 py-3">
										<Link
											to="/projects/$projectId/tms/cases/$caseKey"
											params={{ projectId, caseKey: item.case_key }}
											className="font-mono text-primary hover:underline"
										>
											{item.case_key}
										</Link>
									</td>
									<td className="px-4 py-3 text-t1">{item.title}</td>
									<td className="px-4 py-3">
										<span className="font-medium">
											{Math.round(item.flake_rate * 100)}%
										</span>
										{item.marked_flaky && (
											<Badge variant="outline" className="ml-2 text-xs">
												manual
											</Badge>
										)}
									</td>
									<td className="px-4 py-3">
										<div className="flex flex-wrap gap-1">
											{statusHistoryEntries(item.id, item.status_history).map(
												({ key, status }) => (
													<span
														key={key}
														className={cn(
															"rounded px-1.5 py-0.5 text-xs font-medium",
															STATUS_COLORS[status] ?? "bg-gray-500/10 text-t3",
														)}
													>
														{status}
													</span>
												),
											)}
										</div>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
		</div>
	);
}
