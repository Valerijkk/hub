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

/** Props для диалога переназначения задач. */
interface ReassignDialogProps {
	user: ManagedUser;
	onConfirm: (targetId: string) => void;
	onClose: () => void;
	disabled: boolean;
}

/**
 * ReassignDialog -- диалог подтверждения переназначения задач пользователя.
 * Содержит поле ввода ID целевого пользователя и кнопки подтверждения/отмены.
 */
function ReassignDialog({ user, onConfirm, onClose, disabled }: ReassignDialogProps) {
	const [targetId, setTargetId] = useState("");

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
			<div className="w-96 rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
				<h3 className="mb-4 text-[16px] font-bold">Переназначить задачи</h3>
				<p className="mb-3 text-[13px] text-t2">
					Все задачи пользователя{" "}
					<strong>
						{user.name || user.email}
					</strong>{" "}
					будут переназначены.
				</p>
				<input
					type="text"
					value={targetId}
					onChange={(e) => setTargetId(e.target.value)}
					placeholder="ID целевого пользователя"
					className="mb-4 w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none placeholder:text-t3"
				/>
				<div className="flex justify-end gap-2">
					<button
						type="button"
						onClick={onClose}
						className="rounded-[var(--r2)] border border-b1 px-3 py-1.5 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
					>
						Отмена
					</button>
					<button
						type="button"
						onClick={() => onConfirm(targetId)}
						disabled={!targetId || disabled}
						className="rounded-[var(--r2)] bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] px-3 py-1.5 text-[13px] font-bold text-white disabled:opacity-50"
					>
						Переназначить
					</button>
				</div>
			</div>
		</div>
	);
}

/** Props для диалога GDPR-удаления пользователя. */
interface GdprDeleteDialogProps {
	user: ManagedUser;
	onConfirm: () => void;
	onClose: () => void;
	disabled: boolean;
}

/**
 * GdprDeleteDialog -- диалог подтверждения безвозвратного удаления данных пользователя (GDPR).
 * Требует ввод слова «УДАЛИТЬ» для подтверждения действия.
 */
function GdprDeleteDialog({ user, onConfirm, onClose, disabled }: GdprDeleteDialogProps) {
	const [confirmText, setConfirmText] = useState("");

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
			<div className="w-96 rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
				<h3 className="mb-2 text-[16px] font-bold text-red">
					Удаление данных (GDPR)
				</h3>
				<p className="mb-2 text-[13px] text-t2">
					Все данные пользователя{" "}
					<strong>{user.name || user.email}</strong>{" "}
					будут безвозвратно удалены.
				</p>
				<p className="mb-3 text-[13px] font-semibold text-red">
					Это действие необратимо. Введите УДАЛИТЬ для подтверждения.
				</p>
				<input
					type="text"
					value={confirmText}
					onChange={(e) => setConfirmText(e.target.value)}
					placeholder="Введите УДАЛИТЬ"
					className="mb-4 w-full h-[42px] px-[14px] bg-[var(--input)] border border-red/[0.3] rounded-[var(--r2)] text-[14px] text-t1 outline-none placeholder:text-t3"
				/>
				<div className="flex justify-end gap-2">
					<button
						type="button"
						onClick={onClose}
						className="rounded-[var(--r2)] border border-b1 px-3 py-1.5 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
					>
						Отмена
					</button>
					<button
						type="button"
						onClick={onConfirm}
						disabled={confirmText !== "УДАЛИТЬ" || disabled}
						className="rounded-[var(--r2)] bg-red px-3 py-1.5 text-[13px] font-bold text-white hover:bg-[var(--redS)] disabled:opacity-50"
					>
						Удалить навсегда
					</button>
				</div>
			</div>
		</div>
	);
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
	const [gdprDialogUser, setGdprDialogUser] = useState<ManagedUser | null>(
		null,
	);

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
	const handleReassign = async (targetId: string) => {
		if (!reassignDialogUser || !targetId) return;

		setActionUserId(reassignDialogUser.id);
		try {
			const result = await reassignTasks(
				reassignDialogUser.id,
				targetId,
			);
			alert(`Переназначено задач: ${result.reassigned_count}`);
			setReassignDialogUser(null);
		} catch (err) {
			console.error("Ошибка переназначения:", err);
		} finally {
			setActionUserId(null);
		}
	};

	/** Удаление данных (GDPR). */
	const handleGdprDelete = async () => {
		if (!gdprDialogUser) return;

		setActionUserId(gdprDialogUser.id);
		try {
			await gdprDeleteUser(gdprDialogUser.id);
			setUsers((prev) => prev.filter((u) => u.id !== gdprDialogUser.id));
			setGdprDialogUser(null);
		} catch (err) {
			console.error("Ошибка GDPR удаления:", err);
		} finally {
			setActionUserId(null);
		}
	};

	return (
		<div className="space-y-4">
			{/* Invite section */}
			<div className="flex gap-[10px] mb-5 p-4 bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
				<input
					type="text"
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					placeholder="Email нового пользователя"
					className="flex-1 h-10 px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)] placeholder:text-t3"
				/>
				<button
					type="button"
					className="inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
				>
					Пригласить
				</button>
			</div>

			{/* User table */}
				<table className="w-full" style={{ borderCollapse: "separate", borderSpacing: 0 }}>
					<thead>
						<tr>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Пользователь</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Роль</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Статус</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">Последняя активность</th>
							<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2 w-[100px]">Действия</th>
						</tr>
					</thead>
					<tbody>
						{loading ? (
							<tr>
								<td
									colSpan={5}
									className="px-4 py-8 text-center text-t3 text-[13px]"
								>
									Загрузка...
								</td>
							</tr>
						) : filtered.length === 0 ? (
							<tr>
								<td
									colSpan={5}
									className="px-4 py-8 text-center text-t3 text-[13px]"
								>
									Пользователи не найдены
								</td>
							</tr>
						) : (
							filtered.map((user) => (
								<tr key={user.id} className={user.status === "deactivated" ? "opacity-50" : ""}>
									<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] hover:bg-[var(--hover)]">
										<div className="flex items-center gap-[11px]">
											<div className="w-8 h-8 rounded-full flex items-center justify-center text-[11px] font-bold text-white shrink-0 bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)]">
												{(user.name?.[0] ?? user.email[0] ?? "?").toUpperCase()}
											</div>
											<div>
												<div className="text-[14px] font-semibold">{user.name || user.display_name || "-"}</div>
												<div className="text-[12px] text-t3">{user.email}</div>
											</div>
										</div>
									</td>
									<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] hover:bg-[var(--hover)]">
										<span
											className={`inline-flex px-[10px] py-[3px] rounded-full text-[11px] font-[650] ${
												user.system_role === "admin"
													? "bg-blue/[0.08] text-blue border border-blue/[0.15]"
													: user.system_role === "moderator"
														? "bg-purple/[0.08] text-purple border border-purple/[0.15]"
														: user.status === "deactivated"
															? "bg-[var(--gray)]/[0.08] text-[var(--gray)] border border-[var(--gray)]/[0.15]"
															: "bg-green/[0.08] text-green border border-green/[0.15]"
											}`}
										>
											{user.system_role}
										</span>
									</td>
									<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] hover:bg-[var(--hover)]">
										{user.status === "deactivated"
											? <span className="text-t4">{"\u25CB"} Деактивирован</span>
											: <span className="text-green">{"\u25CF"} Онлайн</span>
										}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] text-t2 hover:bg-[var(--hover)]">
										{user.created_at ? new Date(user.created_at).toLocaleDateString("ru-RU") : "-"}
									</td>
									<td className="px-4 py-[14px] border-b border-b1 align-middle hover:bg-[var(--hover)]">
										<div className="flex gap-1">
											{user.status !== "deactivated" && (
												<button
													type="button"
													onClick={() => handleDeactivate(user)}
													disabled={actionUserId === user.id}
													className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-t1 disabled:opacity-50"
													title="Деактивировать"
												>
													{"\u2718"}
												</button>
											)}
											<button
												type="button"
												onClick={() => setReassignDialogUser(user)}
												disabled={actionUserId === user.id}
												className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-t1 disabled:opacity-50"
												title="Переназначить"
											>
												{"\u270E"}
											</button>
											<button
												type="button"
												onClick={() => setGdprDialogUser(user)}
												disabled={actionUserId === user.id}
												className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-red disabled:opacity-50"
												title="GDPR удаление"
											>
												{"\uD83D\uDDD1"}
											</button>
										</div>
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>

			{/* Диалог переназначения */}
			{reassignDialogUser && (
				<ReassignDialog
					user={reassignDialogUser}
					onConfirm={handleReassign}
					onClose={() => setReassignDialogUser(null)}
					disabled={actionUserId === reassignDialogUser.id}
				/>
			)}

			{/* Диалог GDPR удаления */}
			{gdprDialogUser && (
				<GdprDeleteDialog
					user={gdprDialogUser}
					onConfirm={handleGdprDelete}
					onClose={() => setGdprDialogUser(null)}
					disabled={actionUserId === gdprDialogUser.id}
				/>
			)}
		</div>
	);
}
