import { Fragment, useCallback, useEffect, useState } from "react";
import {
	type AuditLogEntry,
	type AuditLogFilters,
	banUser,
	getAuditLogs,
} from "@/api/admin";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { toastApiError } from "@/lib/format-api-error";

const ACTION_OPTIONS = [
	{ value: "", label: "Все действия" },
	{ value: "login", label: "Вход" },
	{ value: "logout", label: "Выход" },
	{ value: "create", label: "Создание" },
	{ value: "update", label: "Обновление" },
	{ value: "delete", label: "Удаление" },
	{ value: "ban", label: "Блокировка" },
	{ value: "request", label: "HTTP-запрос" },
	{ value: "request.error", label: "HTTP-ошибка" },
	{ value: "alert.firing", label: "Алерт (сраб.)" },
	{ value: "alert.resolved", label: "Алерт (снят)" },
];

const ENTITY_OPTIONS = [
	{ value: "", label: "Все типы" },
	{ value: "http", label: "HTTP" },
	{ value: "user", label: "Пользователь" },
	{ value: "issue", label: "Задача" },
	{ value: "project", label: "Проект" },
	{ value: "channel", label: "Канал" },
	{ value: "message", label: "Сообщение" },
	{ value: "file", label: "Файл" },
	{ value: "session", label: "Сессия" },
	{ value: "call", label: "Звонок" },
	{ value: "alert", label: "Алерт" },
];

function formatDuration(entry: AuditLogEntry): string {
	const ms = entry.details?.duration_ms;
	if (typeof ms === "number") return `${ms} ms`;
	return "-";
}

function formatHttpPath(entry: AuditLogEntry): string {
	if (entry.entity_type === "alert") {
		const labels = entry.details?.labels as Record<string, string> | undefined;
		const name =
			(typeof labels?.alertname === "string" && labels.alertname) ||
			entry.entity_id ||
			"alert";
		const severity =
			typeof labels?.severity === "string" ? ` · ${labels.severity}` : "";
		return `${name}${severity}`;
	}
	if (entry.entity_type !== "http") {
		return (
			entry.entity_type +
			(entry.entity_id ? ` · ${entry.entity_id.slice(0, 8)}` : "")
		);
	}
	const method = entry.details?.method;
	const path = entry.details?.path ?? entry.entity_id;
	if (typeof method === "string" && typeof path === "string") {
		return `${method} ${path}`;
	}
	return String(path ?? entry.entity_type);
}

/**
 * AuditLogViewer — детальный журнал действий системы для администратора.
 */
export function AuditLogViewer() {
	const [entries, setEntries] = useState<AuditLogEntry[]>([]);
	const [total, setTotal] = useState(0);
	const [page, setPage] = useState(1);
	const [loading, setLoading] = useState(false);
	const [expandedId, setExpandedId] = useState<string | null>(null);
	const [banTarget, setBanTarget] = useState<AuditLogEntry | null>(null);
	const [banReason, setBanReason] = useState("");
	const [banning, setBanning] = useState(false);

	const [filters, setFilters] = useState<AuditLogFilters>({
		page: 1,
		per_page: 25,
	});

	const fetchLogs = useCallback(async (signal?: { cancelled: boolean }) => {
		setLoading(true);
		try {
			const resp = await getAuditLogs({ ...filters, page });
			if (signal?.cancelled) return;
			setEntries(resp.items);
			setTotal(resp.total);
		} catch (err) {
			if (!signal?.cancelled) {
				toastApiError(err, "Не удалось загрузить журнал аудита");
			}
		} finally {
			if (!signal?.cancelled) setLoading(false);
		}
	}, [filters, page]);

	useEffect(() => {
		const signal = { cancelled: false };
		void fetchLogs(signal);
		const interval = setInterval(() => void fetchLogs(signal), 30000);
		return () => {
			signal.cancelled = true;
			clearInterval(interval);
		};
	}, [fetchLogs]);

	const exportCSV = () => {
		const headers = [
			"Время",
			"Действие",
			"Тип",
			"Сущность",
			"Пользователь",
			"IP",
			"Длительность",
			"Детали",
		];
		const rows = entries.map((e) => [
			new Date(e.created_at).toLocaleString("ru-RU"),
			e.action,
			e.entity_type,
			e.entity_id ?? "",
			e.user_id ?? "",
			e.ip_address ?? "",
			formatDuration(e),
			e.details ? JSON.stringify(e.details) : "",
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

	const handleBan = async () => {
		if (!banTarget?.user_id || !banReason.trim()) return;
		setBanning(true);
		try {
			await banUser(banTarget.user_id, banReason.trim());
			setBanTarget(null);
			setBanReason("");
			await fetchLogs();
		} catch (err) {
			toastApiError(err, "Не удалось заблокировать пользователя");
		} finally {
			setBanning(false);
		}
	};

	const totalPages = Math.ceil(total / (filters.per_page ?? 25));

	const actionBadgeClass = (action: string) => {
		switch (action) {
			case "ban":
			case "delete":
			case "request.error":
			case "alert.firing":
				return "bg-red/[0.08] text-red";
			case "deploy":
			case "create":
				return "bg-green/[0.08] text-green";
			case "update":
			case "request":
				return "bg-blue/[0.08] text-blue";
			case "comment":
				return "bg-purple/[0.08] text-purple";
			case "login":
			case "logout":
				return "bg-orange/[0.08] text-orange";
			case "alert.resolved":
				return "bg-green/[0.08] text-green";
			default:
				return "bg-blue/[0.08] text-blue";
		}
	};

	const inputClass =
		"h-9 px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 outline-none placeholder:text-t2";

	return (
		<div className="space-y-4">
			<div className="flex flex-wrap gap-[10px] mb-4">
				<label htmlFor="audit-date-from" className="sr-only">
					Дата от
				</label>
				<input
					id="audit-date-from"
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
					className={`w-[150px] ${inputClass} [color-scheme:dark]`}
				/>
				<label htmlFor="audit-date-to" className="sr-only">
					Дата до
				</label>
				<input
					id="audit-date-to"
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
					className={`w-[150px] ${inputClass} [color-scheme:dark]`}
				/>
				<label htmlFor="audit-filter-action" className="sr-only">
					Действие
				</label>
				<Select
					value={filters.action || "all"}
					onValueChange={(v) =>
						setFilters((f) => ({ ...f, action: v === "all" ? undefined : v }))
					}
				>
					<SelectTrigger
						id="audit-filter-action"
						className={`w-[150px] ${inputClass}`}
					>
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
				<label htmlFor="audit-filter-entity" className="sr-only">
					Тип сущности
				</label>
				<Select
					value={filters.entity_type || "all"}
					onValueChange={(v) =>
						setFilters((f) => ({
							...f,
							entity_type: v === "all" ? undefined : v,
						}))
					}
				>
					<SelectTrigger
						id="audit-filter-entity"
						className={`w-[150px] ${inputClass}`}
					>
						<SelectValue placeholder="Тип сущности" />
					</SelectTrigger>
					<SelectContent>
						{ENTITY_OPTIONS.map((o) => (
							<SelectItem key={o.value || "all"} value={o.value || "all"}>
								{o.label}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<label htmlFor="audit-filter-user" className="sr-only">
					User ID
				</label>
				<input
					id="audit-filter-user"
					type="text"
					placeholder="User ID"
					value={filters.user_id ?? ""}
					onChange={(e) =>
						setFilters((f) => ({
							...f,
							user_id: e.target.value || undefined,
						}))
					}
					className={`w-[140px] ${inputClass}`}
				/>
				<label htmlFor="audit-filter-ip" className="sr-only">
					IP-адрес
				</label>
				<input
					id="audit-filter-ip"
					type="text"
					placeholder="IP"
					value={filters.ip_address ?? ""}
					onChange={(e) =>
						setFilters((f) => ({
							...f,
							ip_address: e.target.value || undefined,
						}))
					}
					className={`w-[120px] ${inputClass}`}
				/>
				<label
					htmlFor="audit-filter-slow"
					className="flex items-center gap-2 h-9 px-2 text-[12px] text-[var(--t2)] cursor-pointer select-none"
				>
					<input
						id="audit-filter-slow"
						type="checkbox"
						checked={filters.slow ?? false}
						onChange={(e) =>
							setFilters((f) => ({
								...f,
								slow: e.target.checked || undefined,
							}))
						}
						className="rounded border-b1"
					/>
					Медленные (&gt;500ms)
				</label>
				<button
					type="button"
					onClick={() => {
						setPage(1);
						void fetchLogs();
					}}
					className="px-[14px] py-[6px] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] rounded-[var(--r2)] text-[12px] font-semibold text-white"
				>
					Применить
				</button>
				<div className="flex-1" />
				<button
					type="button"
					onClick={exportCSV}
					aria-label="Экспорт в CSV"
					className="px-[14px] py-[6px] bg-[var(--card)] border border-b1 rounded-[var(--r2)] text-[12px] font-semibold text-[var(--t2)] hover:text-t1 hover:border-b2 transition-colors duration-180"
				>
					CSV
				</button>
			</div>

			<table
				className="w-full"
				aria-label="Журнал аудита"
				style={{ borderCollapse: "separate", borderSpacing: 0 }}
			>
				<thead>
					<tr>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							Время
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							Действие
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							Сущность
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							Пользователь
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							Время отклика
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2"
						>
							IP
						</th>
						<th
							scope="col"
							className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-[var(--t2)] uppercase tracking-[0.08em] px-4 py-3 border-b border-b2 w-[80px]"
						>
							Детали
						</th>
					</tr>
				</thead>
				<tbody>
					{loading && entries.length === 0 ? (
						<tr>
							<td
								colSpan={7}
								className="px-4 py-8 text-center text-t2 text-[13px]"
							>
								Загрузка...
							</td>
						</tr>
					) : entries.length === 0 ? (
						<tr>
							<td
								colSpan={7}
								className="px-4 py-8 text-center text-t2 text-[13px]"
							>
								Записи не найдены
							</td>
						</tr>
					) : (
						entries.map((entry) => (
							<Fragment key={entry.id}>
								<tr className="hover:bg-[var(--hover)]">
									<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t2 whitespace-nowrap">
										{new Date(entry.created_at).toLocaleString("ru-RU")}
									</td>
									<td className="px-4 py-[14px] border-b border-b1">
										<span
											className={`inline-flex px-[10px] py-[2px] rounded-[var(--r1)] text-[11.5px] font-[650] ${actionBadgeClass(entry.action)}`}
										>
											{entry.action}
										</span>
									</td>
									<td className="px-4 py-[14px] border-b border-b1 text-[13px] font-mono text-t2 max-w-[280px] truncate">
										{formatHttpPath(entry)}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 text-[13px]">
										<div className="flex items-center gap-2">
											<span className="font-mono text-[12px]">
												{entry.user_id?.slice(0, 8) ?? "-"}
											</span>
											{entry.user_id && entry.action !== "ban" && (
												<button
													type="button"
													aria-label="Заблокировать пользователя"
													onClick={() => setBanTarget(entry)}
													className="text-[11px] text-[var(--redS)] hover:underline"
												>
													Бан
												</button>
											)}
										</div>
									</td>
									<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t2 font-mono">
										{formatDuration(entry)}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 font-mono text-[12px] text-t2">
										{entry.ip_address ?? "-"}
									</td>
									<td className="px-4 py-[14px] border-b border-b1">
										{entry.details && (
											<button
												type="button"
												aria-expanded={expandedId === entry.id}
												onClick={() =>
													setExpandedId((id) =>
														id === entry.id ? null : entry.id,
													)
												}
												className="text-[11px] text-[var(--blueS)] hover:underline"
											>
												{expandedId === entry.id ? "Скрыть" : "Показать"}
											</button>
										)}
									</td>
								</tr>
								{expandedId === entry.id && entry.details && (
									<tr>
										<td
											colSpan={7}
											className="px-4 py-3 border-b border-b1 bg-[var(--raised)]"
										>
											<pre className="text-[11px] text-t2 overflow-x-auto whitespace-pre-wrap">
												{JSON.stringify(entry.details, null, 2)}
											</pre>
										</td>
									</tr>
								)}
							</Fragment>
						))
					)}
				</tbody>
			</table>

			{totalPages > 1 && (
				<div className="flex items-center justify-between mt-4">
					<span className="text-[12px] text-t2">Всего: {total} записей</span>
					<div className="flex gap-1">
						<button
							type="button"
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
							type="button"
							disabled={page >= totalPages}
							onClick={() => setPage((p) => p + 1)}
							className="px-3 py-[5px] rounded-[var(--r2)] border border-b1 text-[12px] text-t2 hover:text-t1 hover:border-b2 disabled:opacity-50 transition-colors duration-180"
						>
							Вперёд
						</button>
					</div>
				</div>
			)}

			{banTarget && (
				<div
					className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
					role="dialog"
					aria-modal="true"
					aria-labelledby="ban-dialog-title"
				>
					<div className="w-[420px] rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
						<h3
							id="ban-dialog-title"
							className="mb-2 text-[16px] font-bold text-[var(--red)]"
						>
							Блокировка пользователя
						</h3>
						<p className="mb-3 text-[13px] text-t2">
							ID: <span className="font-mono">{banTarget.user_id}</span>
						</p>
						<label
							htmlFor="ban-reason"
							className="mb-2 block text-[12px] font-semibold text-[var(--t2)]"
						>
							Причина блокировки
						</label>
						<textarea
							id="ban-reason"
							value={banReason}
							onChange={(e) => setBanReason(e.target.value)}
							placeholder="Причина блокировки (спам, оскорбления, нарушение правил...)"
							rows={3}
							className="mb-4 w-full px-3 py-2 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 outline-none resize-none placeholder:text-[var(--t2)]"
						/>
						<div className="flex justify-end gap-2">
							<button
								type="button"
								onClick={() => {
									setBanTarget(null);
									setBanReason("");
								}}
								className="rounded-[var(--r2)] border border-b1 px-3 py-1.5 text-[13px] text-[var(--t2)]"
							>
								Отмена
							</button>
							<button
								type="button"
								onClick={() => void handleBan()}
								disabled={!banReason.trim() || banning}
								className="rounded-[var(--r2)] bg-red px-3 py-1.5 text-[13px] font-bold text-white disabled:opacity-50"
							>
								Заблокировать
							</button>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}
