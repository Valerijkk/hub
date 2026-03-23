import { useCallback, useEffect, useState } from "react";
import {
	getPreferences,
	type PreferenceDTO,
	updatePreference,
} from "@/api/notifications";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

/** Типы уведомлений с русскими названиями. */
const NOTIFICATION_TYPES = [
	{ key: "assigned", label: "Назначение задачи" },
	{ key: "mentioned", label: "Упоминание" },
	{ key: "status_changed", label: "Смена статуса" },
	{ key: "comment_added", label: "Новый комментарий" },
	{ key: "message_received", label: "Личное сообщение" },
	{ key: "call_started", label: "Начало звонка" },
	{ key: "run_completed", label: "Завершение прогона" },
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

	useEffect(() => {
		void loadPreferences();
	}, []);

	async function loadPreferences() {
		try {
			const prefs = await getPreferences();
			setPreferences(prefs ?? []);
		} catch {
			// Ошибка загрузки — оставляем пустой массив.
		} finally {
			setLoading(false);
		}
	}

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
			} catch {
				// Ошибка обновления — игнорируем.
			}
		},
		[isEnabled],
	);

	if (loading) {
		return (
			<div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
				Загрузка настроек...
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle className="text-base">Настройки уведомлений</CardTitle>
				</CardHeader>
				<CardContent>
					<div className="overflow-x-auto">
						<table className="w-full text-sm">
							<thead>
								<tr className="border-b">
									<th className="pb-2 pr-4 text-left font-medium">Тип</th>
									{CHANNELS.map((ch) => (
										<th
											key={ch.key}
											className="pb-2 px-4 text-center font-medium"
										>
											{ch.label}
										</th>
									))}
								</tr>
							</thead>
							<tbody>
								{NOTIFICATION_TYPES.map((nt) => (
									<tr key={nt.key} className="border-b last:border-0">
										<td className="py-3 pr-4">{nt.label}</td>
										{CHANNELS.map((ch) => (
											<td key={ch.key} className="py-3 px-4 text-center">
												<input
													type="checkbox"
													checked={isEnabled(nt.key, ch.key)}
													onChange={() => void togglePreference(nt.key, ch.key)}
													className="h-4 w-4 rounded border-gray-300"
													aria-label={`${nt.label} - ${ch.label}`}
												/>
											</td>
										))}
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">Тихие часы</CardTitle>
				</CardHeader>
				<CardContent>
					<p className="mb-4 text-sm text-muted-foreground">
						Во время тихих часов несрочные уведомления подавляются. Срочные
						уведомления (блокировки) всегда доставляются.
					</p>
					<div className="flex items-center gap-4">
						<div className="space-y-1">
							<Label htmlFor="quiet-start">Начало</Label>
							<Input
								id="quiet-start"
								type="time"
								className="w-32"
								placeholder="23:00"
							/>
						</div>
						<span className="mt-5 text-muted-foreground">&mdash;</span>
						<div className="space-y-1">
							<Label htmlFor="quiet-end">Конец</Label>
							<Input
								id="quiet-end"
								type="time"
								className="w-32"
								placeholder="07:00"
							/>
						</div>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
