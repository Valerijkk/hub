import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	approvePendingUser,
	listPendingUsers,
	type PendingUser,
	rejectPendingUser,
} from "@/api/admin-org";
import { QueryErrorAlert } from "@/components/query-state-view";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { toastApiError } from "@/lib/format-api-error";

/**
 * Очередь пользователей, ожидающих одобрения admin (ORG-06).
 */
export function PendingUsersPanel() {
	const [pending, setPending] = useState<PendingUser[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [actionId, setActionId] = useState<string | null>(null);
	const [rejectTarget, setRejectTarget] = useState<PendingUser | null>(null);

	const load = useCallback(async () => {
		setLoading(true);
		setLoadError(false);
		try {
			const items = await listPendingUsers();
			setPending(items ?? []);
		} catch (err) {
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить очередь");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void load();
	}, [load]);

	const handleApprove = async (user: PendingUser) => {
		setActionId(user.id);
		try {
			await approvePendingUser(user.id);
			setPending((prev) => prev.filter((u) => u.id !== user.id));
			toast.success(`${user.email} одобрен`);
		} catch (err) {
			toastApiError(err, "Не удалось одобрить");
		} finally {
			setActionId(null);
		}
	};

	const handleReject = async () => {
		if (!rejectTarget) return;
		setActionId(rejectTarget.id);
		try {
			await rejectPendingUser(rejectTarget.id);
			setPending((prev) => prev.filter((u) => u.id !== rejectTarget.id));
			toast.success("Регистрация отклонена");
			setRejectTarget(null);
		} catch (err) {
			toastApiError(err, "Не удалось отклонить");
		} finally {
			setActionId(null);
		}
	};

	return (
		<section className="mb-6 p-4 bg-gradient-to-br from-amber-500/[0.06] to-[var(--card)] border border-amber-500/20 rounded-[var(--r3)]">
			<div className="flex items-center justify-between gap-3 mb-3">
				<div>
					<h3 className="text-[15px] font-bold">Ожидают одобрения</h3>
					<p className="text-[12px] text-t3 mt-0.5">
						Invite-only регистрация · автоудаление через 7 дней
					</p>
				</div>
				<button
					type="button"
					onClick={() => void load()}
					className="text-[12px] text-t2 hover:text-t1 px-2 py-1 rounded-[var(--r1)] border border-b1"
				>
					Обновить
				</button>
			</div>

			{loading ? (
				<p className="text-[13px] text-t3 py-2">Загрузка...</p>
			) : loadError ? (
				<QueryErrorAlert
					message="Не удалось загрузить очередь"
					retry={() => void load()}
					compact
				/>
			) : pending.length === 0 ? (
				<p className="text-[13px] text-t3 py-2">Нет заявок на регистрацию</p>
			) : (
				<ul className="space-y-2">
					{pending.map((user) => (
						<li
							key={user.id}
							className="flex flex-wrap items-center justify-between gap-2 p-3 rounded-[var(--r2)] bg-[var(--base)]/60 border border-b1"
						>
							<div>
								<div className="text-[14px] font-semibold">{user.email}</div>
								<div className="text-[11px] text-t3">
									{user.pending_since
										? `С ${new Date(user.pending_since).toLocaleString("ru-RU")}`
										: "—"}
								</div>
							</div>
							<div className="flex gap-2">
								<button
									type="button"
									disabled={actionId === user.id}
									onClick={() => void handleApprove(user)}
									className="px-3 py-1.5 rounded-[var(--r2)] text-[12px] font-bold bg-green/15 text-green border border-green/25 hover:bg-green/25 disabled:opacity-50"
								>
									Одобрить
								</button>
								<button
									type="button"
									disabled={actionId === user.id}
									onClick={() => setRejectTarget(user)}
									className="px-3 py-1.5 rounded-[var(--r2)] text-[12px] font-bold bg-red/10 text-red border border-red/20 hover:bg-red/20 disabled:opacity-50"
								>
									Отклонить
								</button>
							</div>
						</li>
					))}
				</ul>
			)}

			<ConfirmDialog
				open={rejectTarget !== null}
				onOpenChange={(open) => {
					if (!open) setRejectTarget(null);
				}}
				title="Отклонить регистрацию?"
				description={
					rejectTarget
						? `Пользователь ${rejectTarget.email} будет удалён из очереди.`
						: undefined
				}
				confirmLabel="Отклонить"
				destructive
				disabled={rejectTarget !== null && actionId === rejectTarget.id}
				onConfirm={handleReject}
			/>
		</section>
	);
}
