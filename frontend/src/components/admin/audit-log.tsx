import { useCallback, useEffect, useState } from "react";
import {
	type AuditLogEntry,
	type AuditLogFilters,
	getAuditLogs,
} from "@/api/admin";

/** Типы действий для фильтра. */
const ACTION_OPTIONS = [
	{ value: "", label: "Все действия" },
	{ value: "login", label: "Вход" },
	{ value: "logout", label: "Выход" },
	{ value: "create", label: "Создание" },
	{ value: "update", label: "Обновление" },
	{ value: "delete", label: "Удаление" },
];

/** Типы сущностей для фильтра. */
const ENTITY_TYPE_OPTIONS = [
	{ value: "", label: "Все типы" },
	{ value: "issue", label: "Задача" },
	{ value: "project", label: "Проект" },
	{ value: "test_case", label: "Тест-кейс" },
	{ value: "channel", label: "Канал" },
	{ value: "message", label: "Сообщение" },
	{ value: "file", label: "Файл" },
	{ value: "session", label: "Сессия" },
	{ value: "call", label: "Звонок" },
];

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

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Журнал аудита</h2>
				<button
					onClick={exportCSV}
					className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700"
				>
					Экспорт CSV
				</button>
			</div>

			{/* Фильтры */}
			<div className="flex flex-wrap gap-3">
				<select
					value={filters.action ?? ""}
					onChange={(e) =>
						setFilters((f) => ({ ...f, action: e.target.value || undefined }))
					}
					className="rounded border px-2 py-1.5 text-sm"
				>
					{ACTION_OPTIONS.map((o) => (
						<option key={o.value} value={o.value}>
							{o.label}
						</option>
					))}
				</select>

				<select
					value={filters.entity_type ?? ""}
					onChange={(e) =>
						setFilters((f) => ({
							...f,
							entity_type: e.target.value || undefined,
						}))
					}
					className="rounded border px-2 py-1.5 text-sm"
				>
					{ENTITY_TYPE_OPTIONS.map((o) => (
						<option key={o.value} value={o.value}>
							{o.label}
						</option>
					))}
				</select>

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
					className="rounded border px-2 py-1.5 text-sm"
					placeholder="С даты"
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
					className="rounded border px-2 py-1.5 text-sm"
					placeholder="По дату"
				/>
			</div>

			{/* Таблица */}
			<div className="overflow-x-auto rounded border">
				<table className="w-full text-sm">
					<thead className="bg-muted/50">
						<tr>
							<th className="px-3 py-2 text-left font-medium">Время</th>
							<th className="px-3 py-2 text-left font-medium">Действие</th>
							<th className="px-3 py-2 text-left font-medium">Тип</th>
							<th className="px-3 py-2 text-left font-medium">ID сущности</th>
							<th className="px-3 py-2 text-left font-medium">Пользователь</th>
							<th className="px-3 py-2 text-left font-medium">IP</th>
						</tr>
					</thead>
					<tbody>
						{loading && entries.length === 0 ? (
							<tr>
								<td
									colSpan={6}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Загрузка...
								</td>
							</tr>
						) : entries.length === 0 ? (
							<tr>
								<td
									colSpan={6}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Записи не найдены
								</td>
							</tr>
						) : (
							entries.map((entry) => (
								<tr key={entry.id} className="border-t hover:bg-muted/30">
									<td className="px-3 py-2 whitespace-nowrap">
										{new Date(entry.created_at).toLocaleString("ru-RU")}
									</td>
									<td className="px-3 py-2">
										<span className="rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-800 dark:bg-blue-900 dark:text-blue-200">
											{entry.action}
										</span>
									</td>
									<td className="px-3 py-2">{entry.entity_type}</td>
									<td className="px-3 py-2 font-mono text-xs">
										{entry.entity_id?.slice(0, 8) ?? "-"}
									</td>
									<td className="px-3 py-2 font-mono text-xs">
										{entry.user_id?.slice(0, 8) ?? "-"}
									</td>
									<td className="px-3 py-2 text-xs">
										{entry.ip_address ?? "-"}
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>

			{/* Пагинация */}
			{totalPages > 1 && (
				<div className="flex items-center justify-between">
					<span className="text-sm text-muted-foreground">
						Всего: {total} записей
					</span>
					<div className="flex gap-1">
						<button
							disabled={page <= 1}
							onClick={() => setPage((p) => Math.max(1, p - 1))}
							className="rounded border px-2 py-1 text-sm disabled:opacity-50"
						>
							Назад
						</button>
						<span className="px-2 py-1 text-sm">
							{page} / {totalPages}
						</span>
						<button
							disabled={page >= totalPages}
							onClick={() => setPage((p) => p + 1)}
							className="rounded border px-2 py-1 text-sm disabled:opacity-50"
						>
							Вперёд
						</button>
					</div>
				</div>
			)}
		</div>
	);
}
