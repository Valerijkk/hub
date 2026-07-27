import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
	banUser,
	deactivateUser,
	gdprDeleteUser,
	reassignTasks,
} from "@/api/admin";
import {
	getOrgInvite,
	getRegistrationMode,
	type RegistrationMode,
	setRegistrationMode,
} from "@/api/admin-org";
import { api } from "@/api/client";
import { PendingUsersPanel } from "@/components/admin/pending-users-panel";
import { QueryStateView } from "@/components/query-state-view";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { toastApiError } from "@/lib/format-api-error";

const ADMIN_USERS_QUERY_KEY = ["admin", "users"] as const;

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

/** Нормализует ответ GET /admin/users к массиву. */
async function fetchAdminUsers(): Promise<ManagedUser[]> {
	const data = await api.get<UsersResponse | ManagedUser[]>("/admin/users");
	if (Array.isArray(data)) return data;
	return data.items ?? data.users ?? [];
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
function ReassignDialog({
	user,
	onConfirm,
	onClose,
	disabled,
}: ReassignDialogProps) {
	const [targetId, setTargetId] = useState("");

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
			<div className="w-96 rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
				<h3 className="mb-4 text-[16px] font-bold">Переназначить задачи</h3>
				<p className="mb-3 text-[13px] text-t2">
					Все задачи пользователя <strong>{user.name || user.email}</strong>{" "}
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
function GdprDeleteDialog({
	user,
	onConfirm,
	onClose,
	disabled,
}: GdprDeleteDialogProps) {
	const [confirmText, setConfirmText] = useState("");

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
			<div className="w-96 rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
				<h3 className="mb-2 text-[16px] font-bold text-red">
					Удаление данных (GDPR)
				</h3>
				<p className="mb-2 text-[13px] text-t2">
					Все данные пользователя <strong>{user.name || user.email}</strong>{" "}
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
	const queryClient = useQueryClient();
	const usersQuery = useQuery({
		queryKey: ADMIN_USERS_QUERY_KEY,
		queryFn: fetchAdminUsers,
	});
	const [search, setSearch] = useState("");
	const [actionUserId, setActionUserId] = useState<string | null>(null);
	const [reassignDialogUser, setReassignDialogUser] =
		useState<ManagedUser | null>(null);
	const [gdprDialogUser, setGdprDialogUser] = useState<ManagedUser | null>(
		null,
	);
	const [deactivateDialogUser, setDeactivateDialogUser] =
		useState<ManagedUser | null>(null);
	const [banDialogUser, setBanDialogUser] = useState<ManagedUser | null>(null);
	const [banReason, setBanReason] = useState("");
	const [inviteLink, setInviteLink] = useState("");
	const [inviteLoading, setInviteLoading] = useState(false);
	const [regMode, setRegMode] = useState<RegistrationMode>("invite_only");
	const [regModeLoading, setRegModeLoading] = useState(true);
	const [regModeSaving, setRegModeSaving] = useState(false);

	useEffect(() => {
		let cancelled = false;
		setInviteLoading(true);
		getOrgInvite()
			.then((res) => {
				if (cancelled) return;
				const url =
					typeof window !== "undefined"
						? `${window.location.origin}/login?invite=${encodeURIComponent(res.invite_token)}`
						: res.invite_url;
				setInviteLink(url);
			})
			.catch(() => {
				if (!cancelled) setInviteLink("");
			})
			.finally(() => {
				if (!cancelled) setInviteLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, []);

	useEffect(() => {
		let cancelled = false;
		setRegModeLoading(true);
		getRegistrationMode()
			.then((res) => {
				if (!cancelled) setRegMode(res.mode);
			})
			.catch(() => {
				/* оставляем invite_only по умолчанию */
			})
			.finally(() => {
				if (!cancelled) setRegModeLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, []);

	const handleSaveRegistrationMode = async () => {
		setRegModeSaving(true);
		try {
			const res = await setRegistrationMode(regMode);
			setRegMode(res.mode);
			toast.success(
				res.mode === "open"
					? "Регистрация открыта"
					: "Регистрация только по инвайту",
			);
		} catch (err) {
			toastApiError(err, "Не удалось сохранить режим регистрации");
		} finally {
			setRegModeSaving(false);
		}
	};

	const patchUsers = (updater: (prev: ManagedUser[]) => ManagedUser[]) => {
		queryClient.setQueryData<ManagedUser[]>(ADMIN_USERS_QUERY_KEY, (prev) =>
			updater(prev ?? []),
		);
	};

	/** Деактивировать пользователя. */
	const handleDeactivate = async (user: ManagedUser) => {
		setActionUserId(user.id);
		try {
			await deactivateUser(user.id);
			patchUsers((prev) =>
				prev.map((u) =>
					u.id === user.id ? { ...u, status: "deactivated" } : u,
				),
			);
		} catch (err) {
			toastApiError(err, "Не удалось деактивировать пользователя");
		} finally {
			setActionUserId(null);
		}
	};

	/** Переназначить задачи. */
	const handleReassign = async (targetId: string) => {
		if (!reassignDialogUser || !targetId) return;

		setActionUserId(reassignDialogUser.id);
		try {
			const result = await reassignTasks(reassignDialogUser.id, targetId);
			alert(`Переназначено задач: ${result.reassigned_count}`);
			setReassignDialogUser(null);
		} catch (err) {
			toastApiError(err, "Не удалось переназначить задачи");
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
			patchUsers((prev) => prev.filter((u) => u.id !== gdprDialogUser.id));
			setGdprDialogUser(null);
		} catch (err) {
			toastApiError(err, "Не удалось выполнить GDPR удаление");
		} finally {
			setActionUserId(null);
		}
	};

	/** Заблокировать пользователя с причиной. */
	const handleBan = async () => {
		if (!banDialogUser || !banReason.trim()) return;
		setActionUserId(banDialogUser.id);
		try {
			await banUser(banDialogUser.id, banReason.trim());
			patchUsers((prev) =>
				prev.map((u) =>
					u.id === banDialogUser.id ? { ...u, status: "deactivated" } : u,
				),
			);
			setBanDialogUser(null);
			setBanReason("");
		} catch (err) {
			toastApiError(err, "Не удалось заблокировать пользователя");
		} finally {
			setActionUserId(null);
		}
	};

	return (
		<div className="space-y-4">
			<PendingUsersPanel />

			{/* Инвайт-ссылка организации */}
			<div className="flex flex-col gap-2 mb-5 p-4 bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
				<h3 className="text-[14px] font-bold">Ссылка-приглашение в Hub</h3>
				<p className="text-[12px] text-t3">
					Отправьте новым участникам — после регистрации они попадут в очередь
					одобрения
				</p>
				<div className="flex gap-2">
					<input
						type="text"
						readOnly
						aria-label="Ссылка-приглашение в Hub"
						value={
							inviteLoading
								? "Загрузка..."
								: inviteLink || "Организация не настроена"
						}
						className="flex-1 h-10 px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 outline-none"
					/>
					<button
						type="button"
						disabled={!inviteLink}
						onClick={() => {
							void navigator.clipboard.writeText(inviteLink);
							toast.success("Ссылка скопирована");
						}}
						className="inline-flex items-center px-4 py-2 rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white disabled:opacity-50"
					>
						Копировать
					</button>
				</div>
			</div>

			{/* Режим регистрации (A11-013) */}
			<div className="flex flex-col gap-2 mb-5 p-4 bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
				<h3 className="text-[14px] font-bold">Режим регистрации</h3>
				<p className="text-[12px] text-t3">
					open — любой может зарегистрироваться; invite_only — только по
					инвайт-ссылке
				</p>
				<div className="flex flex-wrap items-center gap-2">
					<label htmlFor="admin-reg-mode" className="sr-only">
						Режим регистрации
					</label>
					<select
						id="admin-reg-mode"
						aria-label="Режим регистрации"
						value={regMode}
						disabled={regModeLoading || regModeSaving}
						onChange={(e) => setRegMode(e.target.value as RegistrationMode)}
						className="h-10 min-w-[180px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 outline-none"
					>
						<option value="invite_only">Только по инвайту</option>
						<option value="open">Открытая регистрация</option>
					</select>
					<button
						type="button"
						disabled={regModeLoading || regModeSaving}
						onClick={() => void handleSaveRegistrationMode()}
						className="inline-flex items-center px-4 py-2 rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white disabled:opacity-50"
					>
						{regModeSaving ? "Сохранение…" : "Сохранить"}
					</button>
				</div>
			</div>

			{/* Поиск */}
			<div className="flex gap-[10px] mb-5 p-4 bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)]">
				<input
					type="text"
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					placeholder="Email или имя"
					aria-label="Поиск пользователей"
					className="flex-1 h-10 px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)] placeholder:text-t3"
				/>
				<button
					type="button"
					className="inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
				>
					Пригласить
				</button>
			</div>

			{/* Таблица пользователей */}
			<QueryStateView
				query={usersQuery}
				isEmpty={(users) => users.length === 0}
				emptyFallback={
					<p
						className="px-4 py-8 text-center text-t3 text-[13px]"
						role="status"
					>
						Пользователи не найдены
					</p>
				}
				loadingFallback={
					<div
						className="space-y-2 px-4 py-6"
						aria-busy="true"
						aria-live="polite"
					>
						{[1, 2, 3, 4].map((i) => (
							<div
								key={`admin-user-sk-${i}`}
								className="h-12 rounded-[var(--r2)] animate-shimmer"
							/>
						))}
					</div>
				}
			>
				{(users) => {
					const filtered = users.filter(
						(u) =>
							u.email.toLowerCase().includes(search.toLowerCase()) ||
							(u.name ?? "").toLowerCase().includes(search.toLowerCase()),
					);
					if (filtered.length === 0) {
						return (
							<p
								className="px-4 py-8 text-center text-t3 text-[13px]"
								role="status"
							>
								Пользователи не найдены
							</p>
						);
					}
					return (
						<table
							className="w-full"
							style={{ borderCollapse: "separate", borderSpacing: 0 }}
						>
							<thead>
								<tr>
									<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
										Пользователь
									</th>
									<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
										Роль
									</th>
									<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
										Статус
									</th>
									<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2">
										Последняя активность
									</th>
									<th className="sticky top-0 z-[5] bg-[var(--base)] text-left text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-4 py-3 border-b border-b2 w-[100px]">
										Действия
									</th>
								</tr>
							</thead>
							<tbody>
								{filtered.map((user) => (
									<tr
										key={user.id}
										className={
											user.status === "deactivated" ? "opacity-50" : ""
										}
									>
										<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] hover:bg-[var(--hover)]">
											<div className="flex items-center gap-[11px]">
												<div className="w-8 h-8 rounded-full flex items-center justify-center text-[11px] font-bold text-white shrink-0 bg-gradient-to-br from-[var(--blueS)] to-[var(--purpleS)]">
													{(
														user.name?.[0] ??
														user.email[0] ??
														"?"
													).toUpperCase()}
												</div>
												<div>
													<div className="text-[14px] font-semibold">
														{user.name || user.display_name || "-"}
													</div>
													<div className="text-[12px] text-t3">
														{user.email}
													</div>
												</div>
											</div>
										</td>
										<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] hover:bg-[var(--hover)]">
											<span
												className={`inline-flex px-[10px] py-[3px] rounded-full text-[11px] font-[650] ${
													user.system_role === "founder" ||
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
											{user.status === "pending" ? (
												<span className="text-amber-400">
													{"\u25CF"} Ожидает
												</span>
											) : user.status === "deactivated" ? (
												<span className="text-t4">
													{"\u25CB"} Деактивирован
												</span>
											) : (
												<span className="text-green">{"\u25CF"} Активен</span>
											)}
										</td>
										<td className="px-4 py-[14px] border-b border-b1 align-middle text-[13px] text-t2 hover:bg-[var(--hover)]">
											{user.created_at
												? new Date(user.created_at).toLocaleDateString("ru-RU")
												: "-"}
										</td>
										<td className="px-4 py-[14px] border-b border-b1 align-middle hover:bg-[var(--hover)]">
											<div className="flex gap-1">
												{user.status !== "deactivated" && (
													<>
														<button
															type="button"
															onClick={() => setBanDialogUser(user)}
															disabled={actionUserId === user.id}
															className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-red/[0.1] hover:text-red disabled:opacity-50"
															title="Заблокировать"
															aria-label="Заблокировать пользователя"
														>
															{"\u26D4"}
														</button>
														<button
															type="button"
															onClick={() => setDeactivateDialogUser(user)}
															disabled={actionUserId === user.id}
															className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-t1 disabled:opacity-50"
															title="Деактивировать"
															aria-label="Деактивировать пользователя"
														>
															{"\u2718"}
														</button>
													</>
												)}
												<button
													type="button"
													onClick={() => setReassignDialogUser(user)}
													disabled={actionUserId === user.id}
													className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-t1 disabled:opacity-50"
													title="Переназначить"
													aria-label="Переназначить задачи пользователя"
												>
													{"\u270E"}
												</button>
												<button
													type="button"
													onClick={() => setGdprDialogUser(user)}
													disabled={actionUserId === user.id}
													className="w-[30px] h-[30px] flex items-center justify-center rounded-[var(--r1)] text-t3 transition-colors duration-150 hover:bg-[var(--hover)] hover:text-red disabled:opacity-50"
													title="GDPR удаление"
													aria-label="GDPR удаление данных пользователя"
												>
													{"\uD83D\uDDD1"}
												</button>
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					);
				}}
			</QueryStateView>

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

			{/* Диалог блокировки */}
			{banDialogUser && (
				<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
					<div className="w-[420px] rounded-[var(--r3)] bg-[var(--card)] border border-b1 p-6 shadow-[var(--sh3)]">
						<h3 className="mb-2 text-[16px] font-bold text-red">
							Блокировка пользователя
						</h3>
						<p className="mb-3 text-[13px] text-t2">
							<strong>{banDialogUser.name || banDialogUser.email}</strong> будет
							заблокирован. Все сессии будут отозваны.
						</p>
						<textarea
							value={banReason}
							onChange={(e) => setBanReason(e.target.value)}
							placeholder="Причина блокировки"
							aria-label="Причина блокировки"
							rows={3}
							className="mb-4 w-full px-3 py-2 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[13px] text-t1 outline-none resize-none"
						/>
						<div className="flex justify-end gap-2">
							<button
								type="button"
								onClick={() => {
									setBanDialogUser(null);
									setBanReason("");
								}}
								className="rounded-[var(--r2)] border border-b1 px-3 py-1.5 text-[13px] text-t2"
							>
								Отмена
							</button>
							<button
								type="button"
								onClick={() => void handleBan()}
								disabled={
									!banReason.trim() || actionUserId === banDialogUser.id
								}
								className="rounded-[var(--r2)] bg-red px-3 py-1.5 text-[13px] font-bold text-white disabled:opacity-50"
							>
								Заблокировать
							</button>
						</div>
					</div>
				</div>
			)}

			{/* Диалог деактивации */}
			<ConfirmDialog
				open={deactivateDialogUser !== null}
				onOpenChange={(open) => {
					if (!open) setDeactivateDialogUser(null);
				}}
				title="Деактивировать пользователя?"
				description={
					deactivateDialogUser
						? `Пользователь "${deactivateDialogUser.name || deactivateDialogUser.email}" будет деактивирован. Все его сессии будут отозваны.`
						: undefined
				}
				confirmLabel="Деактивировать"
				destructive
				disabled={
					deactivateDialogUser !== null &&
					actionUserId === deactivateDialogUser.id
				}
				onConfirm={async () => {
					if (deactivateDialogUser) {
						await handleDeactivate(deactivateDialogUser);
					}
				}}
			/>
		</div>
	);
}
