import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { updateProfile } from "@/api/auth";
import {
	getPreferences,
	type PreferenceDTO,
	updatePreference,
} from "@/api/notifications";
import { TelegramLinkCard } from "@/components/notifications/telegram-link-card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toastApiError } from "@/lib/format-api-error";

/** Карточка настройки тихих часов с сохранением в профиль. */
function QuietHoursCard() {
	const [start, setStart] = useState("");
	const [end, setEnd] = useState("");
	const [saving, setSaving] = useState(false);

	const handleSave = useCallback(async () => {
		setSaving(true);
		try {
			await updateProfile({
				quiet_hours_start: start || undefined,
				quiet_hours_end: end || undefined,
			});
			toast.success("Тихие часы сохранены");
		} catch (err) {
			toastApiError(err, "Не удалось сохранить тихие часы");
		} finally {
			setSaving(false);
		}
	}, [start, end]);

	return (
		<div className="settings-quiet px-[18px] py-[18px] bg-gradient-to-br from-[var(--card)] to-[var(--raised)] border border-b1 rounded-[var(--r3)] mt-5 relative overflow-hidden">
			<h4 className="text-[14px] font-bold mb-[6px]">Тихие часы</h4>
			<p className="text-[12px] text-t3 mb-3">
				Несрочные уведомления подавляются в указанный период
			</p>
			<div className="flex items-center gap-3">
				<Input
					id="quiet-start"
					type="time"
					className="h-[38px] px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 [color-scheme:dark] w-auto"
					value={start}
					onChange={(e) => setStart(e.target.value)}
					placeholder="23:00"
				/>
				<span className="text-t4">-</span>
				<Input
					id="quiet-end"
					type="time"
					className="h-[38px] px-3 bg-[var(--input)] border border-b1 rounded-[var(--r2)] text-[14px] text-t1 [color-scheme:dark] w-auto"
					value={end}
					onChange={(e) => setEnd(e.target.value)}
					placeholder="07:00"
				/>
				<Button
					size="sm"
					onClick={() => void handleSave()}
					disabled={saving || (!start && !end)}
					className="text-[12px]"
				>
					{saving ? "..." : "OK"}
				</Button>
			</div>
		</div>
	);
}

/** Типы уведомлений с русскими названиями. */
const NOTIFICATION_TYPES = [
	{ key: "assigned", label: "Назначение задачи" },
	{ key: "mentioned", label: "Упоминание" },
	{ key: "status_changed", label: "Смена статуса" },
	{ key: "comment_added", label: "Новый комментарий" },
	{ key: "message_received", label: "Личное сообщение" },
	{ key: "call_started", label: "Начало звонка" },
	{ key: "run_completed", label: "Завершение прогона" },
	{ key: "run_failed", label: "Ошибки в прогоне" },
	{ key: "deadline_approaching", label: "Приближение дедлайна" },
	{ key: "blocker_added", label: "Блокировка задачи" },
] as const;

/** Каналы доставки. */
const CHANNELS = [
	{ key: "in_app", label: "В приложении" },
	{ key: "telegram", label: "Telegram" },
	{ key: "push", label: "Push в браузере" },
] as const;

/**
 * NotificationSettings — компонент настроек уведомлений.
 * Таблица типов уведомлений x каналов доставки с чекбоксами.
 * Секция тихих часов с выбором времени начала и конца.
 */
export function NotificationSettings() {
	const [preferences, setPreferences] = useState<PreferenceDTO[]>([]);
	const [loading, setLoading] = useState(true);

	const loadPreferences = useCallback(async () => {
		try {
			const prefs = await getPreferences();
			setPreferences(prefs ?? []);
		} catch (err) {
			toastApiError(err, "Не удалось загрузить настройки уведомлений");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void loadPreferences();
	}, [loadPreferences]);

	/** Проверяет, включена ли настройка для данного типа и канала. */
	const isEnabled = useCallback(
		(notifType: string, channel: string) => {
			const pref = preferences.find(
				(p) =>
					p.notification_type === notifType &&
					p.channel === channel &&
					!p.project_id,
			);
			// По умолчанию включено.
			return pref ? pref.enabled : true;
		},
		[preferences],
	);

	/** Переключает настройку. */
	const togglePreference = useCallback(
		async (notifType: string, channel: string) => {
			const current = isEnabled(notifType, channel);

			// Оптимистичное обновление — мгновенный отклик UI до ответа сервера
			setPreferences((prev) => {
				const idx = prev.findIndex(
					(p) =>
						p.notification_type === notifType &&
						p.channel === channel &&
						!p.project_id,
				);
				const existing = idx >= 0 ? prev[idx] : undefined;
				if (existing) {
					const next = [...prev];
					next[idx] = { ...existing, enabled: !current };
					return next;
				}
				return [
					...prev,
					{
						notification_type: notifType,
						channel,
						enabled: !current,
					} as PreferenceDTO,
				];
			});

			try {
				const updated = await updatePreference({
					notification_type: notifType,
					channel,
					enabled: !current,
				});
				setPreferences((prev) => {
					const idx = prev.findIndex(
						(p) =>
							p.notification_type === notifType &&
							p.channel === channel &&
							!p.project_id,
					);
					if (idx >= 0) {
						const next = [...prev];
						next[idx] = updated;
						return next;
					}
					return [...prev, updated];
				});
			} catch (err) {
				// Откат оптимистичного обновления при ошибке API
				setPreferences((prev) => {
					const idx = prev.findIndex(
						(p) =>
							p.notification_type === notifType &&
							p.channel === channel &&
							!p.project_id,
					);
					const existing = idx >= 0 ? prev[idx] : undefined;
					if (existing) {
						const next = [...prev];
						next[idx] = { ...existing, enabled: current };
						return next;
					}
					return prev;
				});
				toastApiError(err, "Не удалось обновить настройку");
			}
		},
		[isEnabled],
	);

	if (loading) {
		return (
			<div className="flex items-center justify-center py-8 text-sm text-t2">
				Загрузка настроек...
			</div>
		);
	}

	return (
		<div className="space-y-0">
			<TelegramLinkCard />

			<div className="overflow-x-auto">
				<table className="w-full border-collapse">
					<thead>
						<tr>
							<th className="px-[14px] py-[10px] text-[11px] font-bold text-t4 text-left uppercase tracking-[0.05em] border-b border-b2" />
							{CHANNELS.map((ch) => (
								<th
									key={ch.key}
									className="px-[14px] py-[10px] text-[11px] font-bold text-t4 text-center uppercase tracking-[0.05em] border-b border-b2"
								>
									{ch.label}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{NOTIFICATION_TYPES.map((nt) => (
							<tr key={nt.key}>
								<td className="px-[14px] py-[10px] border-b border-b1 text-left text-t2 text-[13px]">
									{nt.label}
								</td>
								{CHANNELS.map((ch) => (
									<td
										key={ch.key}
										className="px-[14px] py-[10px] border-b border-b1 text-center text-[13px]"
									>
										<input
											type="checkbox"
											checked={isEnabled(nt.key, ch.key)}
											onChange={() => void togglePreference(nt.key, ch.key)}
											className="w-[18px] h-[18px] accent-[var(--blue)] cursor-pointer"
											aria-label={`${nt.label} - ${ch.label}`}
										/>
									</td>
								))}
							</tr>
						))}
					</tbody>
				</table>
			</div>

			<QuietHoursCard />
		</div>
	);
}
