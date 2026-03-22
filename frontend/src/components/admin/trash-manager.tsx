import { useCallback, useEffect, useState } from "react";
import { getTrash, restoreItem, type TrashItem } from "@/api/admin";

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

	const fetchTrash = useCallback(async () => {
		setLoading(true);
		try {
			const resp = await getTrash(activeTab, undefined, page);
			setItems(resp.items);
			setTotal(resp.total);
		} catch (err) {
			console.error("Ошибка загрузки корзины:", err);
		} finally {
			setLoading(false);
		}
	}, [activeTab, page]);

	useEffect(() => {
		setPage(1);
	}, [activeTab]);

	useEffect(() => {
		fetchTrash();
	}, [fetchTrash]);

	/** Восстановить элемент из корзины. */
	const handleRestore = async (item: TrashItem) => {
		if (!confirm(`Восстановить "${item.name}"?`)) return;

		setRestoringId(item.id);
		try {
			await restoreItem(item.entity_type, item.id);
			await fetchTrash();
		} catch (err) {
			console.error("Ошибка восстановления:", err);
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
						key={tab.key}
						onClick={() => setActiveTab(tab.key)}
						className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
							activeTab === tab.key
								? "border-blue-600 text-blue-600"
								: "border-transparent text-muted-foreground hover:text-foreground"
						}`}
					>
						{tab.label}
					</button>
				))}
			</div>

			{/* Таблица */}
			<div className="overflow-x-auto rounded border">
				<table className="w-full text-sm">
					<thead className="bg-muted/50">
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
								<td
									colSpan={4}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Загрузка...
								</td>
							</tr>
						) : items.length === 0 ? (
							<tr>
								<td
									colSpan={4}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Корзина пуста
								</td>
							</tr>
						) : (
							items.map((item) => (
								<tr key={item.id} className="border-t hover:bg-muted/30">
									<td className="px-3 py-2 font-medium">{item.name}</td>
									<td className="px-3 py-2 text-muted-foreground">
										{new Date(item.deleted_at).toLocaleString("ru-RU")}
									</td>
									<td className="px-3 py-2">
										<span
											className={`font-mono text-sm ${
												item.days_remaining <= 7
													? "text-red-600"
													: "text-muted-foreground"
											}`}
										>
											{item.days_remaining}
										</span>
									</td>
									<td className="px-3 py-2 text-right">
										<button
											onClick={() => handleRestore(item)}
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

			{/* Пагинация */}
			{totalPages > 1 && (
				<div className="flex items-center justify-between">
					<span className="text-sm text-muted-foreground">Всего: {total}</span>
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
