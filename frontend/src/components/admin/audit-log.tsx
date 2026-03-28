import { useCallback, useEffect, useState } from "react";
import {
	type AuditLogEntry,
	type AuditLogFilters,
	getAuditLogs,
} from "@/api/admin";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

/** Типы действий для фильтра. */
const ACTION_OPTIONS = [
	{ value: "", label: "Все действия" },
	{ value: "login", label: "Вход" },
	{ value: "logout", label: "Выход" },
	{ value: "create", label: "Создание" },
	{ value: "update", label: "Обновление" },
	{ value: "delete", label: "Удаление" },
];

/* Entity type options available: issue, project, test_case, channel, message, file, session, call */

/**
 * AuditLogViewer -- компонент просмотра журнала аудита.
 * Отображает пагинированную таблицу с фильтрами по действию, типу, дате.
 * Автоматически обновляется каждые 30 секунд.
 */
export function AuditLogViewer() {
	const [entries, setEntries] = useState<AuditLogEntry[]>([]);
	const [total, setTotal] = useState(0);
	const [page, setPage] = useState(1);
	const [loading, setLoading] = useState(false);

	const [filters, setFilters] = useState<AuditLogFilters>({
		page: 1,
		per_page: 25,
	});

	const fetchLogs = useCallback(async () => {
		setLoading(true);
		try {
			const resp = await getAuditLogs({ ...filters, page });
			setEntries(resp.items);
			setTotal(resp.total);
		} catch (err) {
			console.error("Ошибка загрузки журнала аудита:", err);
		} finally {
			setLoading(false);
		}
	}, [filters, page]);

	useEffect(() => {
		fetchLogs();
		const interval = setInterval(fetchLogs, 30000);
		return () => clearInterval(interval);
	}, [fetchLogs]);

	/** Экспорт журнала в CSV. */
	const exportCSV = () => {
		const headers = [
			"Время",
			"Действие",
			"Тип",
			"ID сущности",
			"ID пользователя",
			"IP",
		];
		const rows = entries.map((e) => [
			new Date(e.created_at).toLocaleString("ru-RU"),
			e.action,
			e.entity_type,
			e.entity_id ?? "",
			e.user_id ?? "",
			e.ip_address ?? "",
		]);
		const csv = [headers, ...rows].map((r) => r.join(",")).join("\n");
		const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `audit-log-${new Date().toISOString().slice(0, 10)}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	};

	const totalPages = Math.ceil(total / (filters.per_page ?? 25));

	/** Badge color for audit action type */
	const actionBadgeClass = (action: string) => {
		switch (action) {
			case "deploy": return "bg-green/[0.08] text-green";
			case "update": return "bg-blue/[0.08] text-blue";
			case "comment": return "bg-purple/[0.08] text-purple";
			case "login": case "logout": return "bg-orange/[0.08] text-orange";
			case "create": return "bg-green/[0.08] text-green";
			case "delete": return "bg-red/[0.08] text-red";
			default: return "bg-blue/[0.08] text-blue";
		}
	};

	return (
		<div className="space-y-4">
			{/* Filters row */}
			<div className="flex gap-[10px] mb-4">
				<input
					type="date"
					value={filters.date_from?.slice(0, 10) ?? ""}
					onChange={(e) =>
						setFilters((f) => ({
							...f,
							date_from: e.target.value
								? `${e.target.value}T00:00:00Z`
								: undefined,
						}))
					}
					className="w-[150px] h-9 px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 [color-scheme:dark]"
				/>
				<input
					type="date"
					value={filters.date_to?.slice(0, 10) ?? ""}
					onChange={(e) =>
						setFilters((f) => ({
							...f,
							date_to: e.target.value
								? `${e.target.value}T23:59:59Z`
								: undefined,
						}))
					}
					className="w-[150px] h-9 px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 [color-scheme:dark]"
				/>
				<Select
					value={filters.action || "all"}
					onValueChange={(v) =>
						setFilters((f) => ({ ...f, action: v === "all" ? undefined : v }))
					}
				>
					<SelectTrigger className="w-[150px] h-9 px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{ACTION_OPTIONS.map((o) => (
							<SelectItem key={o.value || "all"} value={o.value || "all"}>
								{o.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<div className="flex-1" />
				<button
					onClick={exportCSV}
					className="px-[14px] py-[6px] bg-[var(--card)] border border-b1 rounded-[var(--r2)] text-[12px] font-semibold text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
				>
					CSV
				</button>
			</div>

			{/* Table */}
				<table className="w-full" style={{ borderCollapse: "separate", borderSpacing: 0 }}>
					<thead>
						<tr>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Время</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Действие</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Сущность</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Пользователь</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">IP</th>
						</tr>
					</thead>
					<tbody>
						{loading && entries.length === 0 ? (
							<tr>
								<td colSpan={5} className="px-4 py-8 text-center text-t3 text-[13px]">
									Загрузка...
								</td>
							</tr>
						) : entries.length === 0 ? (
							<tr>
								<td colSpan={5} className="px-4 py-8 text-center text-t3 text-[13px]">
									Записи не найдены
								</td>
							</tr>
						) : (
							entries.map((entry) => (
								<tr key={entry.id} className="hover:bg-[var(--hover)]">
									<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t3 whitespace-nowrap">
										{new Date(entry.created_at).toLocaleString("ru-RU")}
									</td>
									<td className="px-4 py-[14px] border-b border-b1">
										<span className={`inline-flex px-[10px] py-[2px] rounded-[var(--r1)] text-[11.5px] font-[650] ${actionBadgeClass(entry.action)}`}>
											{entry.action}
										</span>
									</td>
									<td className="px-4 py-[14px] border-b border-b1 text-[13px]">
										{entry.entity_type}{entry.entity_id ? `, ${entry.entity_id.slice(0, 8)}` : ""}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 text-[13px]">
										{entry.user_id?.slice(0, 8) ?? "-"}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 font-mono text-[12px] text-t4">
										{entry.ip_address ?? "-"}
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>

			{/* Пагинация */}
			{totalPages > 1 && (
				<div className="flex items-center justify-between mt-4">
					<span className="text-[12px] text-t3">
						Всего: {total} записей
					</span>
					<div className="flex gap-1">
						<button
							disabled={page <= 1}
							onClick={() => setPage((p) => Math.max(1, p - 1))}
							className="px-3 py-[5px] rounded-[var(--r2)] border border-b1 text-[12px] text-t2 hover:text-t1 hover:border-b2 disabled:opacity-50 transition-colors duration-180"
						>
							Назад
						</button>
						<span className="px-2 py-[5px] text-[12px] text-t2">
							{page} / {totalPages}
						</span>
						<button
							disabled={page >= totalPages}
							onClick={() => setPage((p) => p + 1)}
							className="px-3 py-[5px] rounded-[var(--r2)] border border-b1 text-[12px] text-t2 hover:text-t1 hover:border-b2 disabled:opacity-50 transition-colors duration-180"
						>
							Вперёд
						</button>
					</div>
				</div>
			)}
		</div>
	);
}
