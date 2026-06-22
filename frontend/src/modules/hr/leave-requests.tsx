import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import {
	approveLeave,
	cancelLeave,
	createLeave,
	LEAVE_STATUS_LABELS,
	LEAVE_TYPE_LABELS,
	type LeaveRequest,
	type LeaveType,
	listMyLeave,
	listPendingLeave,
	rejectLeave,
} from "@/api/hr";
import { Button } from "@/components/ui/button";
import { isSystemAdmin } from "@/lib/is-system-admin";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

const STATUS_COLOR: Record<string, string> = {
	pending: "text-amber-400",
	approved: "text-emerald-400",
	rejected: "text-rose-400",
	cancelled: "text-t3",
};

/** MOD-06: заявки на отпуск + очередь approver (admin). */
export function LeaveRequestsPage() {
	const user = useAuthStore((s) => s.user);
	const isAdmin = isSystemAdmin(user?.system_role);
	const queryClient = useQueryClient();

	const [leaveType, setLeaveType] = useState<LeaveType>("vacation");
	const [startDate, setStartDate] = useState("");
	const [endDate, setEndDate] = useState("");
	const [reason, setReason] = useState("");

	const { data: myData, isLoading } = useQuery({
		queryKey: ["hr", "leave", "my"],
		queryFn: listMyLeave,
	});

	const { data: pendingData } = useQuery({
		queryKey: ["hr", "leave", "pending"],
		queryFn: listPendingLeave,
		enabled: isAdmin,
	});

	const createMutation = useMutation({
		mutationFn: createLeave,
		onSuccess: () => {
			setStartDate("");
			setEndDate("");
			setReason("");
			void queryClient.invalidateQueries({ queryKey: ["hr", "leave"] });
			void queryClient.invalidateQueries({ queryKey: ["calendar"] });
			toast.success("Заявка отправлена на согласование");
		},
		onError: () => toast.error("Не удалось создать заявку"),
	});

	const approveMutation = useMutation({
		mutationFn: ({ id, version }: LeaveRequest) => approveLeave(id, version),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["hr", "leave"] });
			void queryClient.invalidateQueries({ queryKey: ["calendar"] });
			toast.success("Заявка одобрена");
		},
		onError: () => toast.error("Не удалось одобрить"),
	});

	const rejectMutation = useMutation({
		mutationFn: ({ id, version }: LeaveRequest) =>
			rejectLeave(id, version, "Отклонено"),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["hr", "leave"] });
			toast.success("Заявка отклонена");
		},
		onError: () => toast.error("Не удалось отклонить"),
	});

	const cancelMutation = useMutation({
		mutationFn: ({ id, version }: LeaveRequest) => cancelLeave(id, version),
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: ["hr", "leave"] });
			toast.success("Заявка отменена");
		},
		onError: () => toast.error("Не удалось отменить"),
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!startDate || !endDate) return;
		createMutation.mutate({
			leave_type: leaveType,
			start_date: startDate,
			end_date: endDate,
			reason: reason.trim() || undefined,
		});
	};

	const myItems = myData?.items ?? [];
	const pendingItems = pendingData?.items ?? [];

	return (
		<div className="mx-auto max-w-4xl space-y-6 px-4 py-6">
			<header className="space-y-1">
				<h1 className="text-2xl font-bold tracking-tight">HR — отпуска</h1>
				<p className="text-sm text-t2">
					Заявки на отсутствие · согласование admin/founder · блок в календаре
				</p>
			</header>

			<form
				onSubmit={handleSubmit}
				className="space-y-3 rounded-xl border border-b1 bg-[var(--hover)]/10 p-4"
			>
				<p className="text-sm font-medium">Новая заявка</p>
				<div className="grid gap-3 sm:grid-cols-2">
					<label className="space-y-1 text-sm">
						<span className="text-t2">Тип</span>
						<select
							value={leaveType}
							onChange={(e) => setLeaveType(e.target.value as LeaveType)}
							className="w-full rounded-lg border border-b1 px-3 py-2"
						>
							{(Object.keys(LEAVE_TYPE_LABELS) as LeaveType[]).map((t) => (
								<option key={t} value={t}>
									{LEAVE_TYPE_LABELS[t]}
								</option>
							))}
						</select>
					</label>
					<div />
					<label className="space-y-1 text-sm">
						<span className="text-t2">С</span>
						<input
							type="date"
							value={startDate}
							onChange={(e) => setStartDate(e.target.value)}
							className="w-full rounded-lg border border-b1 px-3 py-2"
							required
						/>
					</label>
					<label className="space-y-1 text-sm">
						<span className="text-t2">По</span>
						<input
							type="date"
							value={endDate}
							onChange={(e) => setEndDate(e.target.value)}
							className="w-full rounded-lg border border-b1 px-3 py-2"
							required
						/>
					</label>
				</div>
				<label className="block space-y-1 text-sm">
					<span className="text-t2">Комментарий</span>
					<textarea
						value={reason}
						onChange={(e) => setReason(e.target.value)}
						rows={2}
						className="w-full rounded-lg border border-b1 px-3 py-2"
						placeholder="Необязательно"
					/>
				</label>
				<Button type="submit" disabled={createMutation.isPending}>
					Отправить на согласование
				</Button>
			</form>

			{isAdmin && pendingItems.length > 0 && (
				<section className="space-y-3">
					<h2 className="text-sm font-semibold uppercase tracking-wide text-t2">
						Очередь согласования ({pendingItems.length})
					</h2>
					<ul className="space-y-2">
						{pendingItems.map((item) => (
							<li
								key={item.id}
								className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-b1 p-3"
							>
								<div>
									<p className="font-medium">
										{item.requester_name} · {LEAVE_TYPE_LABELS[item.leave_type]}
									</p>
									<p className="text-sm text-t2">
										{item.start_date} — {item.end_date}
									</p>
								</div>
								<div className="flex gap-2">
									<Button
										type="button"
										size="sm"
										onClick={() => approveMutation.mutate(item)}
									>
										Одобрить
									</Button>
									<Button
										type="button"
										size="sm"
										variant="outline"
										onClick={() => rejectMutation.mutate(item)}
									>
										Отклонить
									</Button>
								</div>
							</li>
						))}
					</ul>
				</section>
			)}

			<section className="space-y-3">
				<h2 className="text-sm font-semibold uppercase tracking-wide text-t2">
					Мои заявки
				</h2>
				{isLoading && (
					<p className="text-sm text-t3" role="status">
						Загрузка…
					</p>
				)}
				<ul className="space-y-2">
					{myItems.map((item) => (
						<li
							key={item.id}
							className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-b1 p-3"
						>
							<div>
								<p className="font-medium">
									{LEAVE_TYPE_LABELS[item.leave_type]} ·{" "}
									<span className={cn("text-sm", STATUS_COLOR[item.status])}>
										{LEAVE_STATUS_LABELS[item.status]}
									</span>
								</p>
								<p className="text-sm text-t2">
									{item.start_date} — {item.end_date}
								</p>
							</div>
							{item.status === "pending" && (
								<Button
									type="button"
									size="sm"
									variant="outline"
									onClick={() => cancelMutation.mutate(item)}
								>
									Отменить
								</Button>
							)}
						</li>
					))}
					{!isLoading && myItems.length === 0 && (
						<li className="text-sm text-t3">Заявок пока нет</li>
					)}
				</ul>
			</section>
		</div>
	);
}
