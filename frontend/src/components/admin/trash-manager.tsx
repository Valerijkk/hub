import { useCallback, useEffect, useState } from "react";
import { getTrash, restoreItem, type TrashItem } from "@/api/admin";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { toastApiError } from "@/lib/format-api-error";

/** Вкладки корзины. */
const TRASH_TABS = [
	{ key: "issue", label: "Задачи" },
	{ key: "test_case", label: "Тест-кейсы" },
	{ key: "file", label: "Файлы" },
	{ key: "project", label: "Проекты" },
] as const;

/**
 * TrashManager -- страница администратора/владельца проекта для управления корзиной.
 * Показывает удалённые элементы с возможностью восстановления.
 * Элементы хранятся 30 дней, после чего удаляются окончательно.
 */
export function TrashManager() {
	const [activeTab, setActiveTab] = useState<string>("issue");
	const [items, setItems] = useState<TrashItem[]>([]);
	const [total, setTotal] = useState(0);
	const [page, setPage] = useState(1);
	const [loading, setLoading] = useState(false);
	const [restoringId, setRestoringId] = useState<string | null>(null);
	const [restoreDialogItem, setRestoreDialogItem] = useState<TrashItem | null>(
		null,
	);

	const fetchTrash = useCallback(async () => {
		setLoading(true);
		try {
			const resp = await getTrash(activeTab, undefined, page);
			setItems(resp.items);
			setTotal(resp.total);
		} catch (err) {
			toastApiError(err, "Не удалось загрузить корзину");
		} finally {
			setLoading(false);
		}
	}, [activeTab, page]);

	// Сброс пагинации при смене категории корзины — `activeTab`
	// намеренный триггер effect, а не значение из тела effect.
	// biome-ignore lint/correctness/useExhaustiveDependencies: effect intentionally re-runs only on activeTab change to reset pagination.
	useEffect(() => {
		setPage(1);
	}, [activeTab]);

	useEffect(() => {
		fetchTrash();
	}, [fetchTrash]);

	/** Восстановить элемент из корзины. */
	const handleRestore = async (item: TrashItem) => {
		setRestoringId(item.id);
		try {
			await restoreItem(item.entity_type, item.id);
			await fetchTrash();
		} catch (err) {
			toastApiError(err, "Не удалось восстановить элемент");
		} finally {
			setRestoringId(null);
		}
	};

	const totalPages = Math.ceil(total / 50);

	return (
		<div className="space-y-4">
			<h2 className="text-xl font-semibold">Корзина</h2>

			{/* Вкладки */}
			<div className="flex gap-1 border-b">
				{TRASH_TABS.map((tab) => (
					<button
						type="button"
						key={tab.key}
						onClick={() => setActiveTab(tab.key)}
						className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
							activeTab === tab.key
								? "border-blue-600 text-blue-600"
								: "border-transparent text-t2 hover:text-t1"
						}`}
					>
						{tab.label}
					</button>
				))}
			</div>

			{/* Таблица */}
			<div className="overflow-x-auto rounded border">
				<table className="w-full text-sm">
					<thead className="bg-raised/50">
						<tr>
							<th className="px-3 py-2 text-left font-medium">Название</th>
							<th className="px-3 py-2 text-left font-medium">Удалено</th>
							<th className="px-3 py-2 text-left font-medium">Осталось дней</th>
							<th className="px-3 py-2 text-right font-medium">Действия</th>
						</tr>
					</thead>
					<tbody>
						{loading ? (
							<tr>
								<td colSpan={4} className="px-3 py-8 text-center text-t2">
									Загрузка...
								</td>
							</tr>
						) : items.length === 0 ? (
							<tr>
								<td colSpan={4} className="px-3 py-8 text-center text-t2">
									Корзина пуста
								</td>
							</tr>
						) : (
							items.map((item) => (
								<tr key={item.id} className="border-t hover:bg-raised/30">
									<td className="px-3 py-2 font-medium">{item.name}</td>
									<td className="px-3 py-2 text-t2">
										{new Date(item.deleted_at).toLocaleString("ru-RU")}
									</td>
									<td className="px-3 py-2">
										<span
											className={`font-mono text-sm ${
												item.days_remaining <= 7 ? "text-red-600" : "text-t2"
											}`}
										>
											{item.days_remaining}
										</span>
									</td>
									<td className="px-3 py-2 text-right">
										<button
											type="button"
											onClick={() => setRestoreDialogItem(item)}
											disabled={restoringId === item.id}
											className="rounded bg-green-600 px-2.5 py-1 text-xs text-white hover:bg-green-700 disabled:opacity-50"
										>
											{restoringId === item.id
												? "Восстановление..."
												: "Восстановить"}
										</button>
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>

			<ConfirmDialog
				open={restoreDialogItem !== null}
				onOpenChange={(open) => {
					if (!open) setRestoreDialogItem(null);
				}}
				title="Восстановить элемент?"
				description={
					restoreDialogItem
						? `"${restoreDialogItem.name}" будет восстановлен из корзины.`
						: undefined
				}
				confirmLabel="Восстановить"
				destructive={false}
				disabled={
					restoreDialogItem !== null && restoringId === restoreDialogItem.id
				}
				onConfirm={async () => {
					if (restoreDialogItem) {
						await handleRestore(restoreDialogItem);
					}
				}}
			/>

			{/* Пагинация */}
			{totalPages > 1 && (
				<div className="flex items-center justify-between">
					<span className="text-sm text-t2">Всего: {total}</span>
					<div className="flex gap-1">
						<button
							type="button"
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
							type="button"
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
