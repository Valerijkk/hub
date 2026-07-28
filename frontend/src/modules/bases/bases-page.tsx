import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import {
	type ColumnType,
	createBase,
	createColumn,
	createRecord,
	deleteBase,
	deleteRecord,
	getBase,
	getTableView,
	listBases,
	updateBase,
	updateRecord,
} from "@/api/bases";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";

const COLUMN_TYPES: Array<{ value: ColumnType; label: string }> = [
	{ value: "text", label: "Текст" },
	{ value: "number", label: "Число" },
	{ value: "date", label: "Дата" },
	{ value: "checkbox", label: "Флаг" },
	{ value: "url", label: "URL" },
];

/** V4-04: клон Lark Base — no-code таблицы и grid-редактор. */
export function BasesPage() {
	const queryClient = useQueryClient();
	const [selectedBaseId, setSelectedBaseId] = useState<string | null>(null);
	const [selectedTableId, setSelectedTableId] = useState<string | null>(null);
	const [newBaseName, setNewBaseName] = useState("");
	const [renameValue, setRenameValue] = useState("");
	const [deleteBaseOpen, setDeleteBaseOpen] = useState(false);
	const [newColName, setNewColName] = useState("");
	const [newColType, setNewColType] = useState<ColumnType>("text");
	const [newRowCells, setNewRowCells] = useState<Record<string, string>>({});

	const {
		data: bases = [],
		isLoading,
		isError: basesError,
		refetch: refetchBases,
	} = useQuery({
		queryKey: ["bases"],
		queryFn: listBases,
	});

	const { data: baseDetail } = useQuery({
		queryKey: ["bases", selectedBaseId],
		queryFn: () => {
			if (!selectedBaseId) {
				throw new Error("base not selected");
			}
			return getBase(selectedBaseId);
		},
		enabled: !!selectedBaseId,
	});

	const { data: tableView } = useQuery({
		queryKey: ["bases", selectedBaseId, "table", selectedTableId],
		queryFn: () => {
			if (!selectedBaseId || !selectedTableId) {
				throw new Error("table not selected");
			}
			return getTableView(selectedBaseId, selectedTableId);
		},
		enabled: !!selectedBaseId && !!selectedTableId,
	});

	const createBaseMutation = useMutation({
		mutationFn: createBase,
		onSuccess: (data) => {
			setNewBaseName("");
			setSelectedBaseId(data.id);
			setRenameValue(data.name);
			if (data.tables[0]) {
				setSelectedTableId(data.tables[0].id);
			}
			void queryClient.invalidateQueries({ queryKey: ["bases"] });
			toast.success("База создана");
		},
		onError: () => toast.error("Не удалось создать базу"),
	});

	const updateBaseMutation = useMutation({
		mutationFn: ({
			id,
			name,
			version,
		}: {
			id: string;
			name: string;
			version: number;
		}) => updateBase(id, { name, version }),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["bases"] });
			if (selectedBaseId) {
				void queryClient.invalidateQueries({
					queryKey: ["bases", selectedBaseId],
				});
			}
			toast.success("База обновлена");
		},
		onError: () => toast.error("Не удалось обновить базу"),
	});

	const deleteBaseMutation = useMutation({
		mutationFn: (id: string) => deleteBase(id),
		onSuccess: () => {
			setSelectedBaseId(null);
			setSelectedTableId(null);
			setRenameValue("");
			void queryClient.invalidateQueries({ queryKey: ["bases"] });
			toast.success("База удалена");
		},
		onError: () => toast.error("Не удалось удалить базу"),
	});

	const createColumnMutation = useMutation({
		mutationFn: () => {
			if (!selectedBaseId || !selectedTableId) {
				return Promise.reject(new Error("table not selected"));
			}
			return createColumn(selectedBaseId, selectedTableId, {
				name: newColName.trim(),
				col_type: newColType,
			});
		},
		onSuccess: () => {
			setNewColName("");
			void queryClient.invalidateQueries({
				queryKey: ["bases", selectedBaseId, "table", selectedTableId],
			});
			toast.success("Колонка добавлена");
		},
		onError: () => toast.error("Не удалось добавить колонку"),
	});

	const createRecordMutation = useMutation({
		mutationFn: () => {
			if (!selectedBaseId || !selectedTableId) {
				return Promise.reject(new Error("table not selected"));
			}
			const cells: Record<string, unknown> = {};
			for (const [k, v] of Object.entries(newRowCells)) {
				if (v.trim()) {
					cells[k] = v.trim();
				}
			}
			return createRecord(selectedBaseId, selectedTableId, cells);
		},
		onSuccess: () => {
			setNewRowCells({});
			void queryClient.invalidateQueries({
				queryKey: ["bases", selectedBaseId, "table", selectedTableId],
			});
			toast.success("Запись добавлена");
		},
		onError: () => toast.error("Не удалось добавить запись"),
	});

	const updateCell = async (
		recordId: string,
		version: number,
		cells: Record<string, unknown>,
	) => {
		if (!selectedBaseId || !selectedTableId) return;
		try {
			await updateRecord(selectedBaseId, selectedTableId, recordId, {
				cells,
				version,
			});
			void queryClient.invalidateQueries({
				queryKey: ["bases", selectedBaseId, "table", selectedTableId],
			});
		} catch {
			toast.error("Не удалось сохранить");
		}
	};

	const handleDeleteRecord = async (recordId: string) => {
		if (!selectedBaseId || !selectedTableId) return;
		try {
			await deleteRecord(selectedBaseId, selectedTableId, recordId);
			void queryClient.invalidateQueries({
				queryKey: ["bases", selectedBaseId, "table", selectedTableId],
			});
		} catch {
			toast.error("Не удалось удалить");
		}
	};

	const handleSelectBase = (id: string, name: string) => {
		setSelectedBaseId(id);
		setSelectedTableId(null);
		setRenameValue(name);
	};

	return (
		<div className="mx-auto max-w-[1400px] space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">Hub Base</h1>
				<p className="text-sm text-t2">
					No-code таблицы · Lark Base clone (V4-04)
				</p>
			</header>

			<div className="grid gap-6 lg:grid-cols-[240px_1fr]">
				<aside className="space-y-4 rounded-xl border border-b1 p-4">
					<h2 className="text-sm font-semibold">Базы</h2>
					<form
						className="flex gap-2"
						onSubmit={(e) => {
							e.preventDefault();
							if (!newBaseName.trim()) return;
							createBaseMutation.mutate({ name: newBaseName.trim() });
						}}
					>
						<Input
							value={newBaseName}
							onChange={(e) => setNewBaseName(e.target.value)}
							placeholder="Новая база"
							className="h-8 text-[13px]"
						/>
						<Button
							type="submit"
							size="sm"
							disabled={createBaseMutation.isPending}
						>
							+
						</Button>
					</form>
					{isLoading ? (
						<p className="text-[13px] text-t3">Загрузка...</p>
					) : basesError ? (
						// W10-A1-006 / P2-FE-01: fail listBases ≠ «нет баз»
						<QueryErrorAlert
							compact
							message="Не удалось загрузить базы"
							retry={() => void refetchBases()}
						/>
					) : (
						<ul className="space-y-1">
							{bases.map((b) => (
								<li key={b.id}>
									<button
										type="button"
										onClick={() => handleSelectBase(b.id, b.name)}
										className={`w-full rounded-md px-2 py-1.5 text-left text-[13px] transition-colors ${
											selectedBaseId === b.id
												? "bg-red/10 text-t1 font-medium"
												: "text-t2 hover:bg-raised"
										}`}
									>
										{b.name}
									</button>
								</li>
							))}
						</ul>
					)}
				</aside>

				<main className="min-w-0 space-y-4">
					{!selectedBaseId && (
						<p className="text-[13px] text-t3 py-12 text-center">
							Создайте или выберите базу слева
						</p>
					)}

					{selectedBaseId && baseDetail && (
						<>
							<form
								className="flex flex-wrap items-end gap-2 rounded-lg border border-b1 bg-[var(--card)] p-3"
								aria-label="Управление базой"
								onSubmit={(e) => {
									e.preventDefault();
									if (!renameValue.trim()) return;
									updateBaseMutation.mutate({
										id: baseDetail.id,
										name: renameValue.trim(),
										version: baseDetail.version,
									});
								}}
							>
								<div className="min-w-[180px] flex-1 space-y-1">
									<label
										htmlFor="base-rename"
										className="text-[11px] font-medium text-t3"
									>
										Название базы
									</label>
									<Input
										id="base-rename"
										value={renameValue}
										onChange={(e) => setRenameValue(e.target.value)}
										className="h-8 text-[13px]"
									/>
								</div>
								<Button
									type="submit"
									size="sm"
									disabled={updateBaseMutation.isPending}
								>
									Сохранить
								</Button>
								<Button
									type="button"
									size="sm"
									variant="outline"
									className="text-[var(--red)]"
									aria-label={`Удалить базу ${baseDetail.name}`}
									onClick={() => setDeleteBaseOpen(true)}
									disabled={deleteBaseMutation.isPending}
								>
									Удалить
								</Button>
							</form>

							<div className="flex flex-wrap gap-2 border-b border-b1 pb-3">
								{baseDetail.tables.map((t) => (
									<button
										key={t.id}
										type="button"
										onClick={() => setSelectedTableId(t.id)}
										className={`rounded-md px-3 py-1.5 text-[13px] ${
											selectedTableId === t.id
												? "bg-card border border-b2 font-medium"
												: "text-t2 hover:text-t1"
										}`}
									>
										{t.name}
									</button>
								))}
							</div>

							{selectedTableId && tableView && (
								<div className="space-y-4">
									<form
										className="flex flex-wrap items-end gap-2 rounded-lg border border-b1 p-3"
										onSubmit={(e) => {
											e.preventDefault();
											if (!newColName.trim()) return;
											createColumnMutation.mutate();
										}}
									>
										<Input
											value={newColName}
											onChange={(e) => setNewColName(e.target.value)}
											placeholder="Новая колонка"
											className="h-8 w-40 text-[13px]"
										/>
										<select
											value={newColType}
											onChange={(e) =>
												setNewColType(e.target.value as ColumnType)
											}
											className="h-8 rounded-md border border-b1 bg-card px-2 text-[13px]"
										>
											{COLUMN_TYPES.map((opt) => (
												<option key={opt.value} value={opt.value}>
													{opt.label}
												</option>
											))}
										</select>
										<Button type="submit" size="sm" variant="outline">
											Колонка
										</Button>
									</form>

									<div className="overflow-x-auto rounded-xl border border-b1">
										<table className="w-full min-w-[480px] text-[13px]">
											<thead>
												<tr className="border-b border-b1 bg-raised/50">
													{tableView.columns.map((col) => (
														<th
															key={col.id}
															className="px-3 py-2 text-left font-semibold text-t2"
														>
															{col.name}
														</th>
													))}
													<th className="w-16 px-2" />
												</tr>
											</thead>
											<tbody>
												{tableView.records.map((rec) => (
													<tr key={rec.id} className="border-b border-b1/60">
														{tableView.columns.map((col) => (
															<td key={col.id} className="px-1 py-1">
																<input
																	type="text"
																	defaultValue={String(rec.cells[col.id] ?? "")}
																	onBlur={(e) => {
																		const next = {
																			...rec.cells,
																			[col.id]: e.target.value,
																		};
																		void updateCell(rec.id, rec.version, next);
																	}}
																	className="w-full min-w-[100px] rounded border border-transparent bg-transparent px-2 py-1 focus:border-b2 focus:bg-card"
																/>
															</td>
														))}
														<td className="px-2">
															<button
																type="button"
																onClick={() => void handleDeleteRecord(rec.id)}
																className="text-[11px] text-red hover:underline"
															>
																×
															</button>
														</td>
													</tr>
												))}
												<tr className="bg-raised/30">
													{tableView.columns.map((col) => (
														<td key={col.id} className="px-1 py-1">
															<input
																type="text"
																value={newRowCells[col.id] ?? ""}
																onChange={(e) =>
																	setNewRowCells((prev) => ({
																		...prev,
																		[col.id]: e.target.value,
																	}))
																}
																placeholder="…"
																className="w-full min-w-[100px] rounded border border-b1 bg-card px-2 py-1"
															/>
														</td>
													))}
													<td className="px-2">
														<Button
															type="button"
															size="sm"
															variant="outline"
															onClick={() => createRecordMutation.mutate()}
															disabled={createRecordMutation.isPending}
														>
															+
														</Button>
													</td>
												</tr>
											</tbody>
										</table>
									</div>
								</div>
							)}
						</>
					)}
				</main>
			</div>

			<ConfirmDialog
				open={deleteBaseOpen}
				onOpenChange={setDeleteBaseOpen}
				title="Удалить базу?"
				description={
					baseDetail
						? `База «${baseDetail.name}» и все таблицы будут удалены.`
						: undefined
				}
				confirmLabel="Удалить"
				disabled={deleteBaseMutation.isPending}
				onConfirm={() => {
					if (!selectedBaseId) return;
					deleteBaseMutation.mutate(selectedBaseId);
				}}
			/>
		</div>
	);
}
