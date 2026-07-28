import { useQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useState } from "react";
import { getProject } from "@/api/projects";
import {
	createWebhook,
	deleteWebhook,
	listDeliveries,
	listWebhooks,
	updateWebhook,
	type WebhookDelivery,
	type WebhookEndpoint,
} from "@/api/webhooks";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { canMutateProject } from "@/lib/permissions";

const AVAILABLE_EVENTS = [
	"issue.created",
	"issue.updated",
	"issue.status_changed",
	"test_run.completed",
	"sprint.completed",
];

/** Страница управления вебхуками проекта. */
export function WebhooksPage({
	projectId,
	canMutate: canMutateProp,
}: {
	projectId: string;
	/** Явный override; иначе берём из `project.my_role`. */
	canMutate?: boolean;
}) {
	const [endpoints, setEndpoints] = useState<WebhookEndpoint[]>([]);
	const [loading, setLoading] = useState(true);
	const [showCreate, setShowCreate] = useState(false);
	const [url, setUrl] = useState("");
	const [selectedEvents, setSelectedEvents] = useState<string[]>([]);
	const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);
	const [showDeliveriesFor, setShowDeliveriesFor] = useState<string | null>(
		null,
	);
	const [deleteDialogId, setDeleteDialogId] = useState<string | null>(null);

	const hasProjectId = Boolean(projectId?.trim());

	const { data: project } = useQuery({
		queryKey: ["project", projectId],
		queryFn: () => getProject(projectId),
		enabled: hasProjectId && canMutateProp === undefined,
	});

	const canMutate = canMutateProp ?? canMutateProject(project?.my_role);

	const fetchEndpoints = useCallback(async () => {
		if (!projectId?.trim()) {
			setEndpoints([]);
			setLoading(false);
			return;
		}
		try {
			setLoading(true);
			const data = await listWebhooks(projectId);
			setEndpoints(data ?? []);
		} finally {
			setLoading(false);
		}
	}, [projectId]);

	useEffect(() => {
		fetchEndpoints();
	}, [fetchEndpoints]);

	const handleCreate = async () => {
		if (!projectId?.trim() || !url.trim() || selectedEvents.length === 0)
			return;
		await createWebhook(projectId, { url: url.trim(), events: selectedEvents });
		setUrl("");
		setSelectedEvents([]);
		setShowCreate(false);
		fetchEndpoints();
	};

	if (!hasProjectId) {
		return (
			<div
				role="alert"
				className="rounded-[var(--r3)] border border-b1 bg-card p-4 text-[13px] text-t2"
				data-testid="webhooks-missing-project"
			>
				Вебхуки доступны только в настройках проекта. Выберите проект и откройте
				раздел «Интеграции».
			</div>
		);
	}

	const handleToggleActive = async (ep: WebhookEndpoint) => {
		await updateWebhook(ep.id, { is_active: !ep.is_active });
		fetchEndpoints();
	};

	const handleDelete = async (id: string) => {
		await deleteWebhook(id);
		fetchEndpoints();
	};

	const handleShowDeliveries = async (id: string) => {
		if (showDeliveriesFor === id) {
			setShowDeliveriesFor(null);
			return;
		}
		const data = await listDeliveries(id);
		setDeliveries(data ?? []);
		setShowDeliveriesFor(id);
	};

	const toggleEvent = (event: string) => {
		setSelectedEvents((prev) =>
			prev.includes(event) ? prev.filter((e) => e !== event) : [...prev, event],
		);
	};

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em]">
					Вебхуки
				</h2>
				{canMutate ? (
					<button
						type="button"
						onClick={() => setShowCreate(true)}
						data-testid="webhook-create-btn"
						className="settings-btn-save inline-flex items-center gap-[7px] px-5 py-[9px] rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)] transition-[color,background-color,border-color,box-shadow,transform,filter,opacity] duration-220 hover:shadow-[0_0_28px_rgba(91,156,245,0.3)] hover:-translate-y-px active:scale-[0.96]"
					>
						Добавить вебхук
					</button>
				) : null}
			</div>

			{showCreate && canMutate && (
				<div className="rounded-[var(--r3)] border border-b1 p-4 bg-[var(--card)]">
					<h3 className="mb-3 text-[14px] font-semibold">Новый вебхук</h3>
					<input
						type="url"
						value={url}
						onChange={(e) => setUrl(e.target.value)}
						placeholder="https://example.com/webhook"
						className="mb-3 w-full h-[42px] px-[14px] bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 outline-none focus:border-[rgba(91,156,245,0.3)] placeholder:text-t3"
					/>
					<div className="mb-3 space-y-1">
						<p className="text-[13px] font-semibold text-t2 mb-1">События:</p>
						{AVAILABLE_EVENTS.map((event) => (
							<label
								key={event}
								className="flex items-center gap-2 text-[13px] text-t2"
							>
								<input
									type="checkbox"
									checked={selectedEvents.includes(event)}
									onChange={() => toggleEvent(event)}
								/>
								{event}
							</label>
						))}
					</div>
					<div className="flex gap-2">
						<button
							type="button"
							onClick={handleCreate}
							className="settings-btn-save inline-flex items-center px-4 py-2 rounded-[var(--r2)] text-[13px] font-bold bg-gradient-to-br from-[var(--blueS)] to-[#2563EB] text-white shadow-[0_0_16px_var(--gB)]"
						>
							Создать
						</button>
						<button
							type="button"
							onClick={() => setShowCreate(false)}
							className="rounded-[var(--r2)] border border-b1 px-4 py-2 text-[13px] text-t2 hover:text-t1 hover:border-b2 transition-colors duration-180"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-t3 text-[13px]">Загрузка...</p>
			) : endpoints.length === 0 ? (
				<p className="text-t3 text-[13px]">Нет вебхуков</p>
			) : (
				<div className="space-y-4">
					{endpoints.map((ep) => (
						<div
							key={ep.id}
							className="rounded-[var(--r3)] border border-b1 p-4 bg-gradient-to-br from-[var(--card)] to-[var(--raised)]"
						>
							<div className="flex items-center justify-between">
								<div>
									<p className="text-[14px] font-semibold">{ep.url}</p>
									<p className="text-[12px] text-t3">{ep.events.join(", ")}</p>
								</div>
								<div className="flex items-center gap-3">
									<span
										className={`rounded-full px-[10px] py-[2px] text-[11px] font-bold ${ep.is_active ? "bg-green/[0.08] text-green border border-green/[0.15]" : "bg-[var(--gray)]/[0.08] text-t4 border border-b1"}`}
									>
										{ep.is_active ? "Активен" : "Отключен"}
									</span>
									{canMutate ? (
										<button
											type="button"
											onClick={() => handleToggleActive(ep)}
											className="text-[12px] font-semibold text-blue hover:underline"
										>
											{ep.is_active ? "Отключить" : "Включить"}
										</button>
									) : null}
									<button
										type="button"
										onClick={() => handleShowDeliveries(ep.id)}
										className="text-[12px] font-semibold text-blue hover:underline"
									>
										Доставки
									</button>
									{canMutate ? (
										<button
											type="button"
											onClick={() => setDeleteDialogId(ep.id)}
											data-testid="webhook-delete-btn"
											className="text-[12px] font-bold text-red hover:underline"
										>
											Удалить
										</button>
									) : null}
								</div>
							</div>

							{showDeliveriesFor === ep.id && (
								<div className="mt-3 border-t border-b1 pt-3">
									<h4 className="mb-2 text-[13px] font-semibold">
										История доставок
									</h4>
									{deliveries.length === 0 ? (
										<p className="text-[13px] text-t3">Нет доставок</p>
									) : (
										<table className="w-full border-collapse">
											<thead>
												<tr className="text-left">
													<th className="text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-3 py-2 border-b border-b2">
														Событие
													</th>
													<th className="text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-3 py-2 border-b border-b2">
														Статус
													</th>
													<th className="text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-3 py-2 border-b border-b2">
														Ошибка
													</th>
													<th className="text-[10px] font-bold text-t4 uppercase tracking-[0.08em] px-3 py-2 border-b border-b2">
														Дата
													</th>
												</tr>
											</thead>
											<tbody>
												{deliveries.map((d) => (
													<tr key={d.id} className="hover:bg-[var(--hover)]">
														<td className="px-3 py-[10px] border-b border-b1 text-[13px]">
															{d.event_type}
														</td>
														<td className="px-3 py-[10px] border-b border-b1">
															<span
																className={`text-[12px] font-bold ${
																	d.status_code >= 200 && d.status_code < 300
																		? "text-green"
																		: "text-red"
																}`}
															>
																{d.status_code}
															</span>
														</td>
														<td
															className="px-3 py-[10px] border-b border-b1 text-[12px] text-t3 max-w-[240px] truncate"
															title={d.error ?? d.response_body ?? ""}
														>
															{d.error ?? d.response_body ?? "—"}
														</td>
														<td className="px-3 py-[10px] border-b border-b1 text-[12px] text-t3">
															{new Date(d.delivered_at).toLocaleString("ru-RU")}
														</td>
													</tr>
												))}
											</tbody>
										</table>
									)}
								</div>
							)}
						</div>
					))}
				</div>
			)}

			<ConfirmDialog
				open={deleteDialogId !== null}
				onOpenChange={(open) => {
					if (!open) setDeleteDialogId(null);
				}}
				title="Удалить вебхук?"
				description="Endpoint больше не будет получать события."
				confirmLabel="Удалить"
				destructive
				onConfirm={async () => {
					if (deleteDialogId) {
						await handleDelete(deleteDialogId);
					}
				}}
			/>
		</div>
	);
}
