import { useCallback, useEffect, useState } from "react";
import { getOperationMetrics, type OperationMetric } from "@/api/admin";
import { toastApiError } from "@/lib/format-api-error";

function formatMs(value: number): string {
	if (value < 1) return "<1";
	return value.toFixed(value >= 100 ? 0 : 1);
}

function formatLastSeen(iso?: string): string {
	if (!iso) return "—";
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return iso;
	return d.toLocaleString();
}

/**
 * OperationMetricsPanel — таблица in-memory метрик observability.Timed (OBS-01).
 */
export function OperationMetricsPanel() {
	const [metrics, setMetrics] = useState<OperationMetric[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState("");

	const load = useCallback(async () => {
		setLoading(true);
		setError("");
		try {
			const data = await getOperationMetrics();
			setMetrics(data);
		} catch (err) {
			toastApiError(err, "Не удалось загрузить метрики");
			setError("Не удалось загрузить метрики");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void load();
	}, [load]);

	if (loading) {
		return (
			<div className="py-8 text-center text-t3 text-[13px]">Загрузка...</div>
		);
	}

	if (error) {
		return (
			<div className="py-8 text-center">
				<p className="text-red text-[13px] mb-3">{error}</p>
				<button
					type="button"
					onClick={() => void load()}
					className="text-[13px] font-semibold text-t2 hover:text-t1"
				>
					Повторить
				</button>
			</div>
		);
	}

	return (
		<div>
			<div className="flex items-center justify-between mb-[18px]">
				<div>
					<h3 className="text-[16px] font-bold">Метрики операций</h3>
					<p className="text-[12px] text-t3 mt-1">
						Агрегаты observability.Timed — in-memory с момента старта процесса.
						Prometheus: <code className="text-[11px]">GET /metrics</code>
						{" · "}
						OTel:{" "}
						<code className="text-[11px]">OTEL_EXPORTER_OTLP_ENDPOINT</code>
					</p>
				</div>
				<button
					type="button"
					onClick={() => void load()}
					className="px-3 py-[6px] text-[12px] font-semibold rounded-[var(--r2)] border border-b1 text-t2 hover:text-t1 hover:border-b2 transition-colors"
				>
					Обновить
				</button>
			</div>

			{metrics.length === 0 ? (
				<p className="text-[13px] text-t3 py-6 text-center">
					Пока нет данных — метрики появятся после вызовов Timed()-операций
				</p>
			) : (
				<div className="overflow-x-auto border border-b1 rounded-[var(--r3)]">
					<table className="w-full text-[13px]">
						<thead>
							<tr className="border-b border-b1 bg-[var(--raised)] text-left text-[11px] uppercase tracking-[0.05em] text-t3">
								<th className="px-4 py-3 font-bold">Операция</th>
								<th className="px-4 py-3 font-bold text-right">Вызовов</th>
								<th className="px-4 py-3 font-bold text-right">Ошибок</th>
								<th className="px-4 py-3 font-bold text-right">Ср. мс</th>
								<th className="px-4 py-3 font-bold text-right">Макс. мс</th>
								<th className="px-4 py-3 font-bold">Последний</th>
							</tr>
						</thead>
						<tbody>
							{metrics.map((m) => (
								<tr
									key={m.operation}
									className="border-b border-b1 last:border-b-0 hover:bg-[var(--raised)]/50"
								>
									<td className="px-4 py-2.5 font-mono text-[12px] text-t1">
										{m.operation}
									</td>
									<td className="px-4 py-2.5 text-right tabular-nums">
										{m.count}
									</td>
									<td
										className={`px-4 py-2.5 text-right tabular-nums ${
											m.error_count > 0 ? "text-red font-semibold" : ""
										}`}
									>
										{m.error_count}
									</td>
									<td className="px-4 py-2.5 text-right tabular-nums">
										{formatMs(m.avg_ms)}
									</td>
									<td className="px-4 py-2.5 text-right tabular-nums">
										{m.max_ms}
									</td>
									<td className="px-4 py-2.5 text-t2 whitespace-nowrap">
										{formatLastSeen(m.last_seen_utc)}
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
