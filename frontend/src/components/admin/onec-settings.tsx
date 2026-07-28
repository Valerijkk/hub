import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	getOnecSettings,
	type OnecEmployeeIDField,
	type OnecSettings,
	testOnecWebhook,
	updateOnecSettings,
} from "@/api/admin-onec";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastApiError } from "@/lib/format-api-error";

const ID_FIELD_OPTIONS: Array<{ value: OnecEmployeeIDField; label: string }> = [
	{ value: "email", label: "Email" },
	{ value: "hub_user_id", label: "Hub User ID (UUID)" },
	{ value: "phone", label: "Телефон" },
];

/** Admin UI для webhook-коннектора 1C leave (MOD-07, Phase 48). */
export function OnecSettingsPanel() {
	const [settings, setSettings] = useState<OnecSettings | null>(null);
	const [enabled, setEnabled] = useState(false);
	const [webhookURL, setWebhookURL] = useState("");
	const [bearerToken, setBearerToken] = useState("");
	const [employeeIDField, setEmployeeIDField] =
		useState<OnecEmployeeIDField>("email");
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [saving, setSaving] = useState(false);
	const [testing, setTesting] = useState(false);

	const refresh = useCallback(async () => {
		setLoading(true);
		setLoadError(false);
		try {
			const data = await getOnecSettings();
			setSettings(data);
			setEnabled(data.enabled);
			setWebhookURL(data.webhook_url);
			setEmployeeIDField(data.employee_id_field);
			setBearerToken("");
		} catch (err) {
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить настройки 1C");
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
			const payload: Parameters<typeof updateOnecSettings>[0] = {
				enabled,
				webhook_url: webhookURL.trim(),
				employee_id_field: employeeIDField,
			};
			if (bearerToken.trim()) {
				payload.bearer_token = bearerToken.trim();
			}
			const updated = await updateOnecSettings(payload);
			setSettings(updated);
			setBearerToken("");
			toast.success("Настройки 1C сохранены");
		} catch (err) {
			toastApiError(err, "Не удалось сохранить");
		} finally {
			setSaving(false);
		}
	};

	const handleTest = async () => {
		setTesting(true);
		try {
			const result = await testOnecWebhook();
			if (result.ok) {
				toast.success(result.message ?? "Webhook OK");
			} else {
				toast.error(result.message ?? "Webhook test failed");
			}
		} catch (err) {
			toastApiError(err, "Тест webhook не удался");
		} finally {
			setTesting(false);
		}
	};

	if (loading) {
		return <p className="text-[13px] text-t3 py-8 text-center">Загрузка...</p>;
	}

	// W9-A1-013: fail ≠ пустая панель 1C
	if (loadError || !settings) {
		return (
			<QueryErrorAlert
				message="Не удалось загрузить настройки 1C"
				retry={() => void refresh()}
			/>
		);
	}

	return (
		<div className="max-w-xl space-y-6">
			<div>
				<h3 className="text-[16px] font-bold mb-1">Коннектор 1C (MOD-07)</h3>
				<p className="text-[13px] text-t3">
					При одобрении отпуска Hub отправляет POST webhook в 1C/ZUP с данными
					сотрудника и периода.
				</p>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-4">
				<label className="flex items-center gap-2 text-[13px] cursor-pointer">
					<input
						type="checkbox"
						checked={enabled}
						onChange={(e) => setEnabled(e.target.checked)}
						className="rounded border-b2"
					/>
					<span>Включить экспорт в 1C</span>
				</label>

				<div className="space-y-2">
					<Label htmlFor="onec-webhook">Webhook URL</Label>
					<Input
						id="onec-webhook"
						value={webhookURL}
						onChange={(e) => setWebhookURL(e.target.value)}
						placeholder="https://1c.example.com/hub/leave"
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor="onec-token">Bearer token</Label>
					<Input
						id="onec-token"
						type="password"
						value={bearerToken}
						onChange={(e) => setBearerToken(e.target.value)}
						placeholder={
							settings?.bearer_token_set
								? "•••••••• (оставьте пустым)"
								: "опционально"
						}
					/>
				</div>

				<div className="space-y-2">
					<Label htmlFor="onec-id-field">Поле external_id для 1C</Label>
					<select
						id="onec-id-field"
						value={employeeIDField}
						onChange={(e) =>
							setEmployeeIDField(e.target.value as OnecEmployeeIDField)
						}
						className="w-full h-9 rounded-[var(--r2)] border border-b1 bg-card px-3 text-[13px]"
					>
						{ID_FIELD_OPTIONS.map((opt) => (
							<option key={opt.value} value={opt.value}>
								{opt.label}
							</option>
						))}
					</select>
				</div>

				<div className="flex gap-2 pt-2">
					<Button type="button" onClick={handleSave} disabled={saving}>
						{saving ? "Сохранение..." : "Сохранить"}
					</Button>
					<Button
						type="button"
						variant="outline"
						onClick={handleTest}
						disabled={testing || !webhookURL.trim()}
					>
						{testing ? "Тест..." : "Тест webhook"}
					</Button>
				</div>
			</div>
		</div>
	);
}
