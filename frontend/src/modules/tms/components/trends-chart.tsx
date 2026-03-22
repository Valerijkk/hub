import { useQuery } from "@tanstack/react-query";
import { getTestTrends, type TrendPoint } from "@/api/tms";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/**
 * Простая полоска стека для отображения соотношения статусов в одном прогоне.
 */
function StatusBar({ point }: { point: TrendPoint }) {
	const { total, passed, failed, skipped, broken } = point;
	if (total === 0) return <div className="h-6 rounded bg-gray-100" />;

	const pPassed = (passed / total) * 100;
	const pFailed = (failed / total) * 100;
	const pBroken = (broken / total) * 100;
	const pSkipped = (skipped / total) * 100;

	return (
		<div className="flex h-6 w-full overflow-hidden rounded">
			{pPassed > 0 && (
				<div
					className="bg-green-500 transition-all"
					style={{ width: `${pPassed}%` }}
					title={`Passed: ${passed}`}
				/>
			)}
			{pFailed > 0 && (
				<div
					className="bg-red-500 transition-all"
					style={{ width: `${pFailed}%` }}
					title={`Failed: ${failed}`}
				/>
			)}
			{pBroken > 0 && (
				<div
					className="bg-orange-500 transition-all"
					style={{ width: `${pBroken}%` }}
					title={`Broken: ${broken}`}
				/>
			)}
			{pSkipped > 0 && (
				<div
					className="bg-yellow-400 transition-all"
					style={{ width: `${pSkipped}%` }}
					title={`Skipped: ${skipped}`}
				/>
			)}
		</div>
	);
}

/**
 * График трендов прогонов тестов.
 * Отображает стековые полоски с pass/fail/broken/skipped для каждого прогона.
 * X-ось: прогоны (от старых к новым).
 */
/** График трендов прогонов тестов с горизонтальными полосками pass/fail/broken/skipped. */
export function TrendsChart({
	projectId,
	limit = 20,
}: {
	projectId: string;
	limit?: number;
}) {
	const { data: trends, isLoading } = useQuery({
		queryKey: ["tms-trends", projectId, limit],
		queryFn: () => getTestTrends(projectId, limit),
	});

	if (isLoading) {
		return (
			<div className="py-8 text-center text-muted-foreground">
				Загрузка трендов...
			</div>
		);
	}

	if (!trends || trends.length === 0) {
		return (
			<Card>
				<CardContent className="py-8 text-center text-muted-foreground">
					Нет данных для трендов. Выполните прогоны тестов.
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-sm">
					Тренды прогонов (последние {trends.length})
				</CardTitle>
				<div className="flex gap-4 text-xs text-muted-foreground">
					<span className="flex items-center gap-1">
						<span className="inline-block h-2 w-3 rounded bg-green-500" />
						Passed
					</span>
					<span className="flex items-center gap-1">
						<span className="inline-block h-2 w-3 rounded bg-red-500" />
						Failed
					</span>
					<span className="flex items-center gap-1">
						<span className="inline-block h-2 w-3 rounded bg-orange-500" />
						Broken
					</span>
					<span className="flex items-center gap-1">
						<span className="inline-block h-2 w-3 rounded bg-yellow-400" />
						Skipped
					</span>
				</div>
			</CardHeader>
			<CardContent>
				<div className="space-y-2">
					{trends.map((point) => (
						<div key={point.run_id} className="flex items-center gap-3">
							<div
								className="w-32 truncate text-xs text-muted-foreground"
								title={point.title}
							>
								{point.title}
							</div>
							<div className="flex-1">
								<StatusBar point={point} />
							</div>
							<div className="w-12 text-right text-xs font-medium">
								{point.pass_rate.toFixed(0)}%
							</div>
						</div>
					))}
				</div>
			</CardContent>
		</Card>
	);
}
