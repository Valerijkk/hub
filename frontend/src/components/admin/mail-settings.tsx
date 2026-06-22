import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	getMailSettings,
	type MailSettings,
	sendMailBroadcast,
	updateMailSettings,
} from "@/api/admin-mail";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastApiError } from "@/lib/format-api-error";

/** Admin UI для hub-mail (Phase 15.2, MAIL-01). */
export function MailSettingsPanel() {
	const [settings, setSettings] = useState<MailSettings | null>(null);
	const [hubMailURL, setHubMailURL] = useState("");
	const [fromAlerts, setFromAlerts] = useState("");
	const [alertsEnabled, setAlertsEnabled] = useState(true);
	const [dailyLimit, setDailyLimit] = useState("20");
	const [broadcastSubject, setBroadcastSubject] = useState("");
	const [broadcastBody, setBroadcastBody] = useState("");
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);

	const refresh = useCallback(async () => {
		setLoading(true);
		try {
			const data = await getMailSettings();
			setSettings(data);
			setHubMailURL(data.hub_mail_url);
			setFromAlerts(data.from_alerts);
			setAlertsEnabled(data.alerts_enabled);
			setDailyLimit(String(data.daily_limit));
		} catch (err) {
			toastApiError(err, "Не удалось загрузить настройки почты");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const handleSave = async () => {
		setSaving(true);
		try {
			const limit = Number.parseInt(dailyLimit, 10);
			const updated = await updateMailSettings({
				hub_mail_url: hubMailURL.trim(),
				from_alerts: fromAlerts.trim(),
				alerts_enabled: alertsEnabled,
				daily_limit: Number.isNaN(limit) ? undefined : limit,
			});
			setSettings(updated);
			toast.success("Настройки почты сохранены");
		} catch (err) {
			toastApiError(err, "Не удалось сохранить");
		} finally {
			setSaving(false);
		}
	};

	if (loading) {
		return <p className="text-[13px] text-t3 py-8 text-center">Загрузка...</p>;
	}

	return (
		<div className="max-w-xl space-y-6">
			<div>
				<h3 className="text-[16px] font-bold mb-1">
					Почтовый сервис (hub-mail)
				</h3>
				<p className="text-[13px] text-t3">
					Transactional alerts: run_failed, new_login и др. (ADR-NOTIF-01)
				</p>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-4">
				<div className="flex items-center gap-2 text-[13px]">
					<span
						className={`inline-block h-2 w-2 rounded-full ${settings?.resend_ready ? "bg-green" : "bg-amber"}`}
					/>
					Resend:{" "}
					{settings?.resend_ready ? "настроен" : "не настроен (dev mode)"}
				</div>

				<div className="space-y-2">
					<Label htmlFor="hub-mail-url">URL hub-mail</Label>
					<Input
						id="hub-mail-url"
						value={hubMailURL}
						onChange={(e) => setHubMailURL(e.target.value)}
						placeholder="http://127.0.0.1:3001"
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor="from-alerts">From (alerts)</Label>
					<Input
						id="from-alerts"
						value={fromAlerts}
						onChange={(e) => setFromAlerts(e.target.value)}
						placeholder="Hub Alerts <alerts@company.com>"
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor="daily-limit">Лимит писем / пользователь / день</Label>
					<Input
						id="daily-limit"
						type="number"
						min={1}
						max={100}
						value={dailyLimit}
						onChange={(e) => setDailyLimit(e.target.value)}
					/>
				</div>

				<label className="flex items-center gap-2 text-[13px] cursor-pointer">
					<input
						type="checkbox"
						checked={alertsEnabled}
						onChange={(e) => setAlertsEnabled(e.target.checked)}
						className="rounded border-b1"
					/>
					Email alerts включены
				</label>

				<Button disabled={saving} onClick={() => void handleSave()}>
					{saving ? "Сохранение..." : "Сохранить"}
				</Button>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-4">
				<h4 className="text-[14px] font-semibold">
					Marketing broadcast (lite)
				</h4>
				<p className="text-[12px] text-t3">
					Отправка всем пользователям с email через hub-mail / Resend
				</p>
				<div className="space-y-2">
					<Label htmlFor="broadcast-subject">Тема</Label>
					<Input
						id="broadcast-subject"
						value={broadcastSubject}
						onChange={(e) => setBroadcastSubject(e.target.value)}
					/>
				</div>
				<div className="space-y-2">
					<Label htmlFor="broadcast-body">HTML / текст</Label>
					<textarea
						id="broadcast-body"
						value={broadcastBody}
						onChange={(e) => setBroadcastBody(e.target.value)}
						rows={4}
						className="w-full rounded-[var(--r2)] border border-b1 bg-[var(--input)] px-3 py-2 text-[13px]"
					/>
				</div>
				<div className="flex gap-2">
					<Button
						variant="outline"
						disabled={!broadcastSubject.trim() || !broadcastBody.trim()}
						onClick={async () => {
							try {
								const r = await sendMailBroadcast({
									subject: broadcastSubject.trim(),
									html_body: broadcastBody.trim(),
									dry_run: true,
								});
								toast.success(`Dry-run: ${r.recipients} получателей`);
							} catch (err) {
								toastApiError(err);
							}
						}}
					>
						Dry-run
					</Button>
					<Button
						disabled={!broadcastSubject.trim() || !broadcastBody.trim()}
						onClick={async () => {
							try {
								const r = await sendMailBroadcast({
									subject: broadcastSubject.trim(),
									html_body: broadcastBody.trim(),
								});
								toast.success(`Отправлено: ${r.sent}/${r.recipients}`);
							} catch (err) {
								toastApiError(err);
							}
						}}
					>
						Отправить
					</Button>
				</div>
			</div>
		</div>
	);
}
