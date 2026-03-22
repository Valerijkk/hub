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
			in_progress: "bg-yellow-100 text-yellow-700",
			completed: "bg-green-100 text-green-700",
			failed: "bg-red-100 text-red-700",
		};
		const labels = {
			in_progress: "В процессе",
			completed: "Завершён",
			failed: "Ошибка",
		};
		return (
			<span className={`rounded-full px-2 py-1 text-xs ${colors[status]}`}>
				{labels[status]}
			</span>
		);
	};

	return (
		<div className="space-y-6 p-6">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Резервные копии</h2>
				<button
					onClick={handleCreate}
					disabled={creating}
					className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
				>
					{creating ? "Создание..." : "Создать бэкап"}
				</button>
			</div>

			{restoreConfirm && (
				<div className="rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-950">
					<p className="mb-2 font-medium text-red-800 dark:text-red-200">
						Это заменит ВСЕ текущие данные. Введите RESTORE для подтверждения.
					</p>
					<div className="flex gap-2">
						<input
							type="text"
							value={restoreInput}
							onChange={(e) => setRestoreInput(e.target.value)}
							placeholder="Введите RESTORE"
							className="rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
						/>
						<button
							onClick={() => handleRestore(restoreConfirm)}
							disabled={restoreInput !== "RESTORE"}
							className="rounded-md bg-red-600 px-4 py-2 text-sm text-white disabled:opacity-50"
						>
							Восстановить
						</button>
						<button
							onClick={() => {
								setRestoreConfirm(null);
								setRestoreInput("");
							}}
							className="rounded-md border px-4 py-2 text-sm"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-gray-500">Загрузка...</p>
			) : backups.length === 0 ? (
				<p className="text-gray-500">Нет резервных копий</p>
			) : (
				<table className="w-full text-sm">
					<thead>
						<tr className="border-b text-left dark:border-gray-700">
							<th className="p-2">Дата</th>
							<th className="p-2">Размер</th>
							<th className="p-2">Статус</th>
							<th className="p-2">Действия</th>
						</tr>
					</thead>
					<tbody>
						{backups.map((b) => (
							<tr key={b.id} className="border-b dark:border-gray-700">
								<td className="p-2">
									{new Date(b.created_at).toLocaleString("ru-RU")}
								</td>
								<td className="p-2">{formatSize(b.file_size)}</td>
								<td className="p-2">{statusBadge(b.status)}</td>
								<td className="p-2">
									<div className="flex gap-2">
										{b.status === "completed" && (
											<button
												onClick={() => setRestoreConfirm(b.id)}
												className="text-blue-600 hover:underline"
											>
												Восстановить
											</button>
										)}
										<button
											onClick={() => handleDelete(b.id)}
											className="text-red-600 hover:underline"
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
