import { useCallback, useEffect, useState } from "react";
import {
	type BackupRecord,
	deleteBackup,
	listBackups,
	restoreBackup,
	triggerBackup,
} from "@/api/backup";

/** Форматирует размер файла в человекочитаемый вид. */
function formatSize(bytes: number): string {
	if (bytes === 0) return "0 B";
	const units = ["B", "KB", "MB", "GB"];
	const i = Math.floor(Math.log(bytes) / Math.log(1024));
	return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
}

/** Страница управления бэкапами в админ-панели. */
export function BackupsPage() {
	const [backups, setBackups] = useState<BackupRecord[]>([]);
	const [loading, setLoading] = useState(true);
	const [creating, setCreating] = useState(false);
	const [restoreConfirm, setRestoreConfirm] = useState<string | null>(null);
	const [restoreInput, setRestoreInput] = useState("");

	const fetchBackups = useCallback(async () => {
		try {
			setLoading(true);
			const data = await listBackups();
			setBackups(data ?? []);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		fetchBackups();
	}, [fetchBackups]);

	const handleCreate = async () => {
		setCreating(true);
		try {
			await triggerBackup();
			fetchBackups();
		} finally {
			setCreating(false);
		}
	};

	const handleRestore = async (id: string) => {
		if (restoreInput !== "RESTORE") return;
		await restoreBackup(id);
		setRestoreConfirm(null);
		setRestoreInput("");
		fetchBackups();
	};

	const handleDelete = async (id: string) => {
		if (!confirm("Удалить этот бэкап?")) return;
		await deleteBackup(id);
		fetchBackups();
	};

	const statusBadge = (status: BackupRecord["status"]) => {
		const colors = {
			in_progress: "bg-orange/[0.08] text-orange border border-orange/[0.15]",
			completed: "bg-green/[0.08] text-green border border-green/[0.15]",
			failed: "bg-red/[0.08] text-red border border-red/[0.15]",
		};
		const labels = {
			in_progress: "В процессе",
			completed: "Завершён",
			failed: "Ошибка",
		};
		return (
			<span className={`inline-flex px-[10px] py-[2px] rounded-full text-[11px] font-[650] ${colors[status]}`}>
				{labels[status]}
			</span>
		);
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em]">Резервные копии</h2>
				<button
					onClick={handleCreate}
					disabled={creating}
					className="inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96] disabled:opacity-50"
				>
					{creating ? "Создание..." : "Создать бэкап"}
				</button>
			</div>

			{restoreConfirm && (
				<div className="rounded-[var(--r3)] border border-red/[0.3] bg-red/[0.06] p-4">
					<p className="mb-2 text-[14px] font-semibold text-red">
						Это заменит ВСЕ текущие данные. Введите RESTORE для подтверждения.
					</p>
					<div className="flex gap-2">
						<input
							type="text"
							value={restoreInput}
							onChange={(e) => setRestoreInput(e.target.value)}
							placeholder="Введите RESTORE"
							className="h-[42px] px-[14px] bg-[var(--input)] border border-red/[0.3] rounded-[var(--r2)] text-[14px] text-t1 outline-none placeholder:text-t3"
						/>
						<button
							onClick={() => handleRestore(restoreConfirm)}
							disabled={restoreInput !== "RESTORE"}
							className="rounded-[var(--r2)] bg-red px-4 py-2 text-[13px] font-bold text-white disabled:opacity-50 hover:bg-[var(--redS)]"
						>
							Восстановить
						</button>
						<button
							onClick={() => {
								setRestoreConfirm(null);
								setRestoreInput("");
							}}
							className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-t3 text-[13px]">Загрузка...</p>
			) : backups.length === 0 ? (
				<p className="text-t3 text-[13px]">Нет резервных копий</p>
			) : (
				<table className="w-full" style={{ borderCollapse: "separate", borderSpacing: 0 }}>
					<thead>
						<tr>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Дата</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Размер</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Статус</th>
							<th className="text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Действия</th>
						</tr>
					</thead>
					<tbody>
						{backups.map((b) => (
							<tr key={b.id} className="hover:bg-[var(--hover)]">
								<td className="px-4 py-[14px] border-b border-b1 text-[12px] text-t3 whitespace-nowrap">
									{new Date(b.created_at).toLocaleString("ru-RU")}
								</td>
								<td className="px-4 py-[14px] border-b border-b1 text-[13px]">{formatSize(b.file_size)}</td>
								<td className="px-4 py-[14px] border-b border-b1">{statusBadge(b.status)}</td>
								<td className="px-4 py-[14px] border-b border-b1">
									<div className="flex gap-2">
										{b.status === "completed" && (
											<button
												onClick={() => setRestoreConfirm(b.id)}
												className="text-[12px] font-semibold text-blue hover:underline"
											>
												Восстановить
											</button>
										)}
										<button
											onClick={() => handleDelete(b.id)}
											className="text-[12px] font-bold text-red hover:underline"
										>
											Удалить
										</button>
									</div>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
		</div>
	);
}
