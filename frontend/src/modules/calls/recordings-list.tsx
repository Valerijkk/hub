import { useCallback, useEffect, useState } from "react";

import type { RecordingResponse } from "@/api/calls";
import * as callsApi from "@/api/calls";

// ---------------------------------------------------------------------------
// RecordingsList — раздел «Записи» проекта (CALL-05)
// ---------------------------------------------------------------------------

/** Пропсы компонента RecordingsList. */
interface RecordingsListProps {
	/** ID проекта для загрузки записей. */
	projectId: string;
}

/**
 * RecordingsList отображает таблицу записей звонков проекта
 * с фильтрацией по дате и кнопкой скачивания через presigned URL.
 *
 * Доступен через sidebar проекта (иконка Video).
 * Записи отсортированы по дате (новые первые).
 */
export function RecordingsList({ projectId }: RecordingsListProps) {
	const [recordings, setRecordings] = useState<RecordingResponse[]>([]);
	const [total, setTotal] = useState(0);
	const [offset, setOffset] = useState(0);
	const [isLoading, setIsLoading] = useState(false);
	const [dateFilter, setDateFilter] = useState("");

	const limit = 20;

	/** Загрузить записи проекта. */
	const loadRecordings = useCallback(async () => {
		setIsLoading(true);
		try {
			const data = await callsApi.listProjectRecordings(
				projectId,
				offset,
				limit,
			);
			setRecordings(data.recordings ?? []);
			setTotal(data.total);
		} catch {
			// Ошибка загрузки
		} finally {
			setIsLoading(false);
		}
	}, [projectId, offset]);

	useEffect(() => {
		loadRecordings();
	}, [loadRecordings]);

	/** Скачать запись по presigned URL. */
	const handleDownload = useCallback(async (recordingId: string) => {
		try {
			const data = await callsApi.getRecordingDownloadURL(recordingId);
			window.open(data.url, "_blank");
		} catch {
			// Ошибка скачивания
		}
	}, []);

	/** Форматирует длительность в минуты:секунды. */
	const formatDuration = (seconds: number) => {
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return `${m}:${s.toString().padStart(2, "0")}`;
	};

	/** Форматирует размер файла. */
	const formatSize = (bytes: number) => {
		if (bytes === 0) return "0 B";
		const units = ["B", "KB", "MB", "GB"];
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / 1024 ** i).toFixed(1)} ${units[i]}`;
	};

	/** Фильтрация по дате. */
	const filteredRecordings = dateFilter
		? recordings.filter((r) => r.started_at.startsWith(dateFilter))
		: recordings;

	const totalPages = Math.ceil(total / limit);
	const currentPage = Math.floor(offset / limit) + 1;

	return (
		<div className="flex flex-col gap-4 p-4">
			<div className="flex items-center justify-between">
				<h2 className="text-lg font-semibold">Записи звонков</h2>
				<input
					type="date"
					value={dateFilter}
					onChange={(e) => setDateFilter(e.target.value)}
					className="rounded-md border border-border bg-base px-3 py-1.5 text-sm"
					aria-label="Фильтр по дате"
				/>
			</div>

			{isLoading ? (
				<div className="flex justify-center py-8 text-t2">Загрузка...</div>
			) : filteredRecordings.length === 0 ? (
				<div className="flex justify-center py-8 text-t2">Записей пока нет</div>
			) : (
				<>
					<div className="overflow-x-auto rounded-lg border border-border">
						<table className="w-full text-sm">
							<thead>
								<tr className="border-b border-border bg-raised/50">
									<th className="px-4 py-2 text-left font-medium">Дата</th>
									<th className="px-4 py-2 text-left font-medium">
										Длительность
									</th>
									<th className="px-4 py-2 text-left font-medium">Статус</th>
									<th className="px-4 py-2 text-left font-medium">Размер</th>
									<th className="px-4 py-2 text-right font-medium" />
								</tr>
							</thead>
							<tbody>
								{filteredRecordings.map((rec) => (
									<tr
										key={rec.id}
										className="border-b border-border last:border-0"
									>
										<td className="px-4 py-2">
											{new Date(rec.started_at).toLocaleString("ru-RU")}
										</td>
										<td className="px-4 py-2">
											{formatDuration(rec.duration_seconds)}
										</td>
										<td className="px-4 py-2">
											<RecordingStatusBadge status={rec.status} />
										</td>
										<td className="px-4 py-2">{formatSize(rec.file_size)}</td>
										<td className="px-4 py-2 text-right">
											{rec.status === "ready" && (
												<button
													type="button"
													onClick={() => handleDownload(rec.id)}
													className="rounded-md bg-primary px-3 py-1 text-xs text-primary-foreground transition-colors hover:bg-primary/90"
												>
													Скачать
												</button>
											)}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{/* Пагинация */}
					{totalPages > 1 && (
						<div className="flex items-center justify-center gap-2">
							<button
								type="button"
								onClick={() => setOffset(Math.max(0, offset - limit))}
								disabled={offset === 0}
								className="rounded-md border border-border px-3 py-1 text-sm disabled:opacity-50"
							>
								Назад
							</button>
							<span className="text-sm text-t2">
								{currentPage} / {totalPages}
							</span>
							<button
								type="button"
								onClick={() => setOffset(offset + limit)}
								disabled={offset + limit >= total}
								className="rounded-md border border-border px-3 py-1 text-sm disabled:opacity-50"
							>
								Далее
							</button>
						</div>
					)}
				</>
			)}
		</div>
	);
}

/** Бейдж статуса записи. */
function RecordingStatusBadge({
	status,
}: {
	status: RecordingResponse["status"];
}) {
	const styles: Record<string, string> = {
		recording: "bg-red-500/10 text-red-500",
		processing: "bg-yellow-500/10 text-yellow-500",
		ready: "bg-green-500/10 text-green-500",
		failed: "bg-destructive/10 text-destructive",
	};

	const labels: Record<string, string> = {
		recording: "Запись",
		processing: "Обработка",
		ready: "Готово",
		failed: "Ошибка",
	};

	return (
		<span
			className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${styles[status] ?? ""}`}
		>
			{labels[status] ?? status}
		</span>
	);
}
