import { useCallback, useEffect, useState } from "react";
import {
	createWebhook,
	deleteWebhook,
	listDeliveries,
	listWebhooks,
	updateWebhook,
	type WebhookDelivery,
	type WebhookEndpoint,
} from "@/api/webhooks";

const AVAILABLE_EVENTS = [
	"issue.created",
	"issue.updated",
	"issue.status_changed",
	"test_run.completed",
	"sprint.completed",
];

/** Страница управления вебхуками проекта. */
export function WebhooksPage({ projectId }: { projectId: string }) {
	const [endpoints, setEndpoints] = useState<WebhookEndpoint[]>([]);
	const [loading, setLoading] = useState(true);
	const [showCreate, setShowCreate] = useState(false);
	const [url, setUrl] = useState("");
	const [selectedEvents, setSelectedEvents] = useState<string[]>([]);
	const [deliveries, setDeliveries] = useState<WebhookDelivery[]>([]);
	const [showDeliveriesFor, setShowDeliveriesFor] = useState<string | null>(
		null,
	);

	const fetchEndpoints = useCallback(async () => {
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
		if (!url.trim() || selectedEvents.length === 0) return;
		await createWebhook(projectId, { url: url.trim(), events: selectedEvents });
		setUrl("");
		setSelectedEvents([]);
		setShowCreate(false);
		fetchEndpoints();
	};

	const handleToggleActive = async (ep: WebhookEndpoint) => {
		await updateWebhook(ep.id, { is_active: !ep.is_active });
		fetchEndpoints();
	};

	const handleDelete = async (id: string) => {
		if (!confirm("Удалить этот вебхук?")) return;
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
		<div className="space-y-6 p-6">
			<div className="flex items-center justify-between">
				<h2 className="text-xl font-semibold">Вебхуки</h2>
				<button
					onClick={() => setShowCreate(true)}
					className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700"
				>
					Добавить вебхук
				</button>
			</div>

			{showCreate && (
				<div className="rounded-md border p-4 dark:border-gray-700">
					<h3 className="mb-3 font-medium">Новый вебхук</h3>
					<input
						type="url"
						value={url}
						onChange={(e) => setUrl(e.target.value)}
						placeholder="https://example.com/webhook"
						className="mb-3 w-full rounded-md border p-2 dark:border-gray-600 dark:bg-gray-800"
					/>
					<div className="mb-3 space-y-1">
						<p className="text-sm font-medium">События:</p>
						{AVAILABLE_EVENTS.map((event) => (
							<label key={event} className="flex items-center gap-2 text-sm">
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
							onClick={handleCreate}
							className="rounded-md bg-blue-600 px-4 py-2 text-sm text-white"
						>
							Создать
						</button>
						<button
							onClick={() => setShowCreate(false)}
							className="rounded-md border px-4 py-2 text-sm"
						>
							Отмена
						</button>
					</div>
				</div>
			)}

			{loading ? (
				<p className="text-gray-500">Загрузка...</p>
			) : endpoints.length === 0 ? (
				<p className="text-gray-500">Нет вебхуков</p>
			) : (
				<div className="space-y-4">
					{endpoints.map((ep) => (
						<div
							key={ep.id}
							className="rounded-md border p-4 dark:border-gray-700"
						>
							<div className="flex items-center justify-between">
								<div>
									<p className="font-medium">{ep.url}</p>
									<p className="text-sm text-gray-500">
										{ep.events.join(", ")}
									</p>
								</div>
								<div className="flex items-center gap-3">
									<span
										className={`rounded-full px-2 py-1 text-xs ${ep.is_active ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"}`}
									>
										{ep.is_active ? "Активен" : "Отключен"}
									</span>
									<button
										onClick={() => handleToggleActive(ep)}
										className="text-sm text-blue-600 hover:underline"
									>
										{ep.is_active ? "Отключить" : "Включить"}
									</button>
									<button
										onClick={() => handleShowDeliveries(ep.id)}
										className="text-sm text-blue-600 hover:underline"
									>
										Доставки
									</button>
									<button
										onClick={() => handleDelete(ep.id)}
										className="text-sm text-red-600 hover:underline"
									>
										Удалить
									</button>
								</div>
							</div>

							{showDeliveriesFor === ep.id && (
								<div className="mt-3 border-t pt-3 dark:border-gray-700">
									<h4 className="mb-2 text-sm font-medium">История доставок</h4>
									{deliveries.length === 0 ? (
										<p className="text-sm text-gray-500">Нет доставок</p>
									) : (
										<table className="w-full text-sm">
											<thead>
												<tr className="text-left text-gray-500">
													<th className="p-1">Событие</th>
													<th className="p-1">Статус</th>
													<th className="p-1">Дата</th>
												</tr>
											</thead>
											<tbody>
												{deliveries.map((d) => (
													<tr key={d.id}>
														<td className="p-1">{d.event_type}</td>
														<td className="p-1">
															<span
																className={
																	d.status_code >= 200 && d.status_code < 300
																		? "text-green-600"
																		: "text-red-600"
																}
															>
																{d.status_code}
															</span>
														</td>
														<td className="p-1">
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
		</div>
	);
}
