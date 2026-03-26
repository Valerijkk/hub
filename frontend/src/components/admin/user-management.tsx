import { useEffect, useState } from "react";
import { deactivateUser, gdprDeleteUser, reassignTasks } from "@/api/admin";
import { api } from "@/api/client";

/** Пользователь для управления. */
interface ManagedUser {
	id: string;
	email: string;
	name: string;
	display_name: string;
	system_role: string;
	status?: string;
	created_at?: string;
}

/** Простой ответ со списком пользователей. */
interface UsersResponse {
	items?: ManagedUser[];
	users?: ManagedUser[];
}

/**
 * UserManagement -- страница администратора для управления пользователями.
 * Позволяет деактивировать, переназначить задачи и удалить данные (GDPR).
 */
export function UserManagement() {
	const [users, setUsers] = useState<ManagedUser[]>([]);
	const [loading, setLoading] = useState(false);
	const [search, setSearch] = useState("");
	const [actionUserId, setActionUserId] = useState<string | null>(null);
	const [reassignDialogUser, setReassignDialogUser] =
		useState<ManagedUser | null>(null);
	const [reassignTargetId, setReassignTargetId] = useState("");
	const [gdprDialogUser, setGdprDialogUser] = useState<ManagedUser | null>(
		null,
	);
	const [gdprConfirmText, setGdprConfirmText] = useState("");

	useEffect(() => {
		setLoading(true);
		api
			.get<UsersResponse | ManagedUser[]>("/admin/users")
			.then((data) => {
				if (Array.isArray(data)) {
					setUsers(data);
				} else if (data.items) {
					setUsers(data.items);
				} else if (data.users) {
					setUsers(data.users);
				}
			})
			.catch(() => {
				// API may not exist yet, use empty list
				setUsers([]);
			})
			.finally(() => setLoading(false));
	}, []);

	/** Фильтрованные пользователи. */
	const filtered = users.filter(
		(u) =>
			u.email.toLowerCase().includes(search.toLowerCase()) ||
			(u.name ?? "").toLowerCase().includes(search.toLowerCase()),
	);

	/** Деактивировать пользователя. */
	const handleDeactivate = async (user: ManagedUser) => {
		if (
			!confirm(
				`Деактивировать пользователя "${user.name || user.email}"? Все его сессии будут отозваны.`,
			)
		)
			return;

		setActionUserId(user.id);
		try {
			await deactivateUser(user.id);
			setUsers((prev) =>
				prev.map((u) =>
					u.id === user.id ? { ...u, status: "deactivated" } : u,
				),
			);
		} catch (err) {
			console.error("Ошибка деактивации:", err);
		} finally {
			setActionUserId(null);
		}
	};

	/** Переназначить задачи. */
	const handleReassign = async () => {
		if (!reassignDialogUser || !reassignTargetId) return;

		setActionUserId(reassignDialogUser.id);
		try {
			const result = await reassignTasks(
				reassignDialogUser.id,
				reassignTargetId,
			);
			alert(`Переназначено задач: ${result.reassigned_count}`);
			setReassignDialogUser(null);
			setReassignTargetId("");
		} catch (err) {
			console.error("Ошибка переназначения:", err);
		} finally {
			setActionUserId(null);
		}
	};

	/** Удаление данных (GDPR). */
	const handleGdprDelete = async () => {
		if (!gdprDialogUser || gdprConfirmText !== "УДАЛИТЬ") return;

		setActionUserId(gdprDialogUser.id);
		try {
			await gdprDeleteUser(gdprDialogUser.id);
			setUsers((prev) => prev.filter((u) => u.id !== gdprDialogUser.id));
			setGdprDialogUser(null);
			setGdprConfirmText("");
		} catch (err) {
			console.error("Ошибка GDPR удаления:", err);
		} finally {
			setActionUserId(null);
		}
	};

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Управление пользователями</h2>
				<input
					type="text"
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					placeholder="Поиск по имени или email..."
					className="w-64 rounded-lg bg-card border-white/10 px-3 py-1.5 text-sm"
				/>
			</div>

			{/* Таблица пользователей */}
			<div className="overflow-x-auto rounded-xl border border-white/[0.06]">
				<table className="w-full text-sm">
					<thead className="bg-muted/50">
						<tr>
							<th className="px-3 py-2.5 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Имя</th>
							<th className="px-3 py-2.5 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Email</th>
							<th className="px-3 py-2.5 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Роль</th>
							<th className="px-3 py-2.5 text-left text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Статус</th>
							<th className="px-3 py-2.5 text-right text-[11px] font-medium uppercase tracking-wider text-muted-foreground">Действия</th>
						</tr>
					</thead>
					<tbody>
						{loading ? (
							<tr>
								<td
									colSpan={5}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Загрузка...
								</td>
							</tr>
						) : filtered.length === 0 ? (
							<tr>
								<td
									colSpan={5}
									className="px-3 py-8 text-center text-muted-foreground"
								>
									Пользователи не найдены
								</td>
							</tr>
						) : (
							filtered.map((user) => (
								<tr key={user.id} className="border-t border-white/[0.06] hover:bg-muted/30">
									<td className="px-3 py-2.5">
										<div className="flex items-center gap-2.5">
											<div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-[11px] font-semibold text-white">
												{(user.name?.[0] ?? user.email[0] ?? "?").toUpperCase()}
											</div>
											<span className="font-medium">{user.name || user.display_name || "-"}</span>
										</div>
									</td>
									<td className="px-3 py-2.5 text-muted-foreground">{user.email}</td>
									<td className="px-3 py-2.5">
										<span
											className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
												user.system_role === "admin"
													? "bg-purple-500/15 text-purple-400"
													: user.system_role === "moderator"
														? "bg-blue-500/15 text-blue-400"
														: "bg-emerald-500/15 text-emerald-400"
											}`}
										>
											{user.system_role}
										</span>
									</td>
									<td className="px-3 py-2.5">
										<span
											className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium ${
												user.status === "deactivated"
													? "bg-red-500/15 text-red-400"
													: "bg-emerald-500/15 text-emerald-400"
											}`}
										>
											<span className={`h-1.5 w-1.5 rounded-full ${user.status === "deactivated" ? "bg-red-400" : "bg-emerald-400"}`} />
											{user.status === "deactivated"
												? "Деактивирован"
												: "Активен"}
										</span>
									</td>
									<td className="px-3 py-2 text-right">
										<div className="flex justify-end gap-1">
											{user.status !== "deactivated" && (
												<button
													onClick={() => handleDeactivate(user)}
													disabled={actionUserId === user.id}
													className="rounded border px-2 py-1 text-xs hover:bg-muted disabled:opacity-50"
												>
													Деактивировать
												</button>
											)}
											<button
												onClick={() => setReassignDialogUser(user)}
												disabled={actionUserId === user.id}
												className="rounded border px-2 py-1 text-xs hover:bg-muted disabled:opacity-50"
											>
												Переназначить
											</button>
											<button
												onClick={() => setGdprDialogUser(user)}
												disabled={actionUserId === user.id}
												className="rounded border border-red-300 px-2 py-1 text-xs text-red-600 hover:bg-red-50 disabled:opacity-50"
											>
												GDPR
											</button>
										</div>
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>

			{/* Диалог переназначения */}
			{reassignDialogUser && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
					<div className="w-96 rounded-lg bg-background p-6 shadow-lg">
						<h3 className="mb-4 text-lg font-semibold">Переназначить задачи</h3>
						<p className="mb-3 text-sm text-muted-foreground">
							Все задачи пользователя{" "}
							<strong>
								{reassignDialogUser.name || reassignDialogUser.email}
							</strong>{" "}
							будут переназначены.
						</p>
						<input
							type="text"
							value={reassignTargetId}
							onChange={(e) => setReassignTargetId(e.target.value)}
							placeholder="ID целевого пользователя"
							className="mb-4 w-full rounded border px-3 py-2 text-sm"
						/>
						<div className="flex justify-end gap-2">
							<button
								onClick={() => {
									setReassignDialogUser(null);
									setReassignTargetId("");
								}}
								className="rounded border px-3 py-1.5 text-sm"
							>
								Отмена
							</button>
							<button
								onClick={handleReassign}
								disabled={!reassignTargetId}
								className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
							>
								Переназначить
							</button>
						</div>
					</div>
				</div>
			)}

			{/* Диалог GDPR удаления */}
			{gdprDialogUser && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
					<div className="w-96 rounded-lg bg-background p-6 shadow-lg">
						<h3 className="mb-2 text-lg font-semibold text-red-600">
							Удаление данных (GDPR)
						</h3>
						<p className="mb-2 text-sm text-muted-foreground">
							Все данные пользователя{" "}
							<strong>{gdprDialogUser.name || gdprDialogUser.email}</strong>{" "}
							будут безвозвратно удалены.
						</p>
						<p className="mb-3 text-sm font-medium text-red-600">
							Это действие необратимо. Введите УДАЛИТЬ для подтверждения.
						</p>
						<input
							type="text"
							value={gdprConfirmText}
							onChange={(e) => setGdprConfirmText(e.target.value)}
							placeholder="Введите УДАЛИТЬ"
							className="mb-4 w-full rounded border border-red-300 px-3 py-2 text-sm"
						/>
						<div className="flex justify-end gap-2">
							<button
								onClick={() => {
									setGdprDialogUser(null);
									setGdprConfirmText("");
								}}
								className="rounded border px-3 py-1.5 text-sm"
							>
								Отмена
							</button>
							<button
								onClick={handleGdprDelete}
								disabled={gdprConfirmText !== "УДАЛИТЬ"}
								className="rounded bg-red-600 px-3 py-1.5 text-sm text-white hover:bg-red-700 disabled:opacity-50"
							>
								Удалить навсегда
							</button>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}
