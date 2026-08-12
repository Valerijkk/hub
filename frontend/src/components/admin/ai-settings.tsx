import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	AI_PROVIDER_PRESETS,
	type AIProvider,
	type AIUsageHistory,
	type AIUsageSummary,
	createAIProvider,
	getAIAdminSettings,
	getAIUsageHistory,
	getAIUsageSummary,
	listAIProviders,
	testAIProvider,
	updateAIAdminSettings,
	updateAIProvider,
} from "@/api/admin-ai";
import { getAIBotInfo } from "@/api/ai";
import { createOrGetDM } from "@/api/messenger";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toastApiError } from "@/lib/format-api-error";

/** Admin UI для AI Provider Hub (Phase 17 W1, AI-01). */
export function AISettingsPanel() {
	const [providers, setProviders] = useState<AIProvider[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [name, setName] = useState("");
	const [baseURL, setBaseURL] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [models, setModels] = useState("");
	const [providerType, setProviderType] = useState("openai_compatible");
	const [ruResident, setRuResident] = useState(false);
	const [saving, setSaving] = useState(false);
	const [usageSummary, setUsageSummary] = useState<AIUsageSummary | null>(null);
	const [usageHistory, setUsageHistory] = useState<AIUsageHistory | null>(null);
	const [systemPrompt, setSystemPrompt] = useState("");
	const [promptSaving, setPromptSaving] = useState(false);
	const [offlineOnly, setOfflineOnly] = useState(false);

	const refresh = useCallback(async () => {
		setLoading(true);
		setLoadError(false);
		try {
			const [data, summary, history, settings] = await Promise.all([
				listAIProviders(),
				getAIUsageSummary().catch(() => null),
				getAIUsageHistory(7).catch(() => null),
				getAIAdminSettings().catch(() => null),
			]);
			setProviders(data);
			setUsageSummary(summary);
			setUsageHistory(history);
			if (settings) {
				setSystemPrompt(settings.system_prompt);
				setOfflineOnly(settings.offline_only);
			}
		} catch (err) {
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить AI-провайдеров");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const applyPreset = (idx: number) => {
		const p = AI_PROVIDER_PRESETS[idx];
		if (!p) return;
		setName(p.name);
		setProviderType(p.provider_type);
		setBaseURL(p.base_url);
		setModels(p.models.join(", "));
		setRuResident(p.ru_resident ?? false);
	};

	const handleCreate = async () => {
		setSaving(true);
		try {
			await createAIProvider({
				name: name.trim(),
				provider_type: providerType,
				base_url: baseURL.trim(),
				api_key: apiKey.trim() || undefined,
				models: models
					.split(",")
					.map((m) => m.trim())
					.filter(Boolean),
				ru_resident: ruResident,
			});
			toast.success("Провайдер добавлен");
			setName("");
			setApiKey("");
			void refresh();
		} catch (err) {
			toastApiError(err, "Не удалось создать провайдера");
		} finally {
			setSaving(false);
		}
	};

	if (loading) {
		return <p className="text-[13px] text-t3 py-8 text-center">Загрузка...</p>;
	}

	// W9-A1-012: fail ≠ пустой реестр провайдеров
	if (loadError) {
		return (
			<QueryErrorAlert
				message="Не удалось загрузить AI-провайдеров"
				retry={() => void refresh()}
			/>
		);
	}

	return (
		<div className="max-w-3xl space-y-6">
			<div>
				<h3 className="text-[16px] font-bold mb-1">AI Provider Hub</h3>
				<p className="text-[13px] text-t3">
					Реестр LLM-провайдеров: Ollama, DeepSeek, OpenRouter, GigaChat (v3.5).
					GigaChat ключ:{" "}
					<code className="text-xs">client_id:client_secret</code>. В каналах:{" "}
					<code className="text-xs">@hub-ai</code> или DM с ботом.
				</p>
				{usageSummary && (
					<p className="text-[12px] text-t2 mt-1">
						Сегодня: {usageSummary.total_requests} запросов ·{" "}
						{usageSummary.active_users} пользователей · ~
						{usageSummary.tokens_in + usageSummary.tokens_out} токенов
					</p>
				)}
				{usageHistory && usageHistory.days.length > 0 && (
					<p className="text-[11px] text-t3 mt-1">
						7 дней:{" "}
						{usageHistory.days
							.map((d) => `${d.date.slice(5)}:${d.requests}`)
							.join(" · ")}
					</p>
				)}
				{offlineOnly && (
					<p className="text-[11px] text-amber-600 mt-1">
						AI_OFFLINE_ONLY: только Ollama/localhost
					</p>
				)}
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="mt-2"
					onClick={async () => {
						try {
							const bot = await getAIBotInfo();
							const ch = await createOrGetDM(bot.user_id);
							window.location.href = `/messenger/${ch.id}`;
						} catch (err) {
							toastApiError(err, "Не удалось открыть DM с @hub-ai");
						}
					}}
				>
					Открыть DM с @hub-ai
				</Button>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-3">
				<h4 className="text-[14px] font-semibold">
					System prompt @hub-ai (M3-014)
				</h4>
				<textarea
					value={systemPrompt}
					onChange={(e) => setSystemPrompt(e.target.value)}
					rows={4}
					className="w-full rounded-[var(--r2)] border border-b1 bg-[var(--input)] px-3 py-2 text-[13px] resize-y min-h-[80px]"
					placeholder="Ты AI-ассистент команды..."
				/>
				<Button
					type="button"
					size="sm"
					disabled={promptSaving || !systemPrompt.trim()}
					onClick={async () => {
						setPromptSaving(true);
						try {
							await updateAIAdminSettings({
								system_prompt: systemPrompt.trim(),
							});
							toast.success("System prompt сохранён");
						} catch (err) {
							toastApiError(err, "Не удалось сохранить prompt");
						} finally {
							setPromptSaving(false);
						}
					}}
				>
					{promptSaving ? "Сохранение..." : "Сохранить prompt"}
				</Button>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-4">
				<h4 className="text-[14px] font-semibold">Добавить провайдера</h4>
				<div className="flex flex-wrap gap-2">
					{AI_PROVIDER_PRESETS.map((p, i) => (
						<Button
							key={p.name}
							type="button"
							variant="outline"
							size="sm"
							onClick={() => applyPreset(i)}
						>
							{p.name}
						</Button>
					))}
				</div>
				<div className="grid gap-3 sm:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="ai-name">Название</Label>
						<Input
							id="ai-name"
							value={name}
							onChange={(e) => setName(e.target.value)}
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="ai-type">Тип</Label>
						<select
							id="ai-type"
							value={providerType}
							onChange={(e) => setProviderType(e.target.value)}
							className="w-full h-9 rounded-[var(--r2)] border border-b1 bg-[var(--input)] px-3 text-[13px]"
						>
							<option value="openai_compatible">OpenAI-compatible</option>
							<option value="ollama">Ollama</option>
							<option value="gigachat">GigaChat</option>
							<option value="anthropic">Anthropic</option>
						</select>
					</div>
					<div className="space-y-2 sm:col-span-2">
						<Label htmlFor="ai-url">Base URL</Label>
						<Input
							id="ai-url"
							value={baseURL}
							onChange={(e) => setBaseURL(e.target.value)}
							placeholder="https://api.deepseek.com/v1"
						/>
					</div>
					<div className="space-y-2 sm:col-span-2">
						<Label htmlFor="ai-key">
							API Key (не показывается после сохранения)
						</Label>
						<Input
							id="ai-key"
							type="password"
							value={apiKey}
							onChange={(e) => setApiKey(e.target.value)}
							autoComplete="off"
						/>
					</div>
					<div className="space-y-2 sm:col-span-2">
						<Label htmlFor="ai-models">Модели (через запятую)</Label>
						<Input
							id="ai-models"
							value={models}
							onChange={(e) => setModels(e.target.value)}
							placeholder="deepseek-chat, llama3"
						/>
					</div>
				</div>
				<label className="flex items-center gap-2 text-[13px] cursor-pointer">
					<input
						type="checkbox"
						checked={ruResident}
						onChange={(e) => setRuResident(e.target.checked)}
						className="rounded border-b1"
					/>
					RU-resident (данные в РФ)
				</label>
				<Button
					disabled={saving || !name.trim() || !baseURL.trim()}
					onClick={() => void handleCreate()}
				>
					{saving ? "Сохранение..." : "Добавить"}
				</Button>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 overflow-hidden">
				<table className="w-full text-[13px]">
					<thead className="bg-raised border-b border-b1">
						<tr>
							<th className="text-left px-4 py-2 font-medium">Провайдер</th>
							<th className="text-left px-4 py-2 font-medium">Модели</th>
							<th className="text-left px-4 py-2 font-medium">Статус</th>
							<th className="text-right px-4 py-2 font-medium">Действия</th>
						</tr>
					</thead>
					<tbody>
						{providers.length === 0 ? (
							<tr>
								<td colSpan={4} className="px-4 py-6 text-center text-t3">
									Нет провайдеров — добавьте Ollama или DeepSeek
								</td>
							</tr>
						) : (
							providers.map((p) => (
								<tr key={p.id} className="border-b border-b1 last:border-0">
									<td className="px-4 py-3">
										<div className="font-medium">{p.name}</div>
										<div className="text-[11px] text-t3 truncate max-w-[200px]">
											{p.base_url}
										</div>
										{p.ru_resident && (
											<span className="text-[10px] text-green">RU</span>
										)}
									</td>
									<td className="px-4 py-3 text-t2">
										{p.models.join(", ") || "—"}
									</td>
									<td className="px-4 py-3">
										<span
											className={`inline-block h-2 w-2 rounded-full mr-1 ${p.enabled ? "bg-green" : "bg-t3"}`}
										/>
										{p.enabled ? "Вкл" : "Выкл"}
										{p.has_api_key && (
											<span className="text-[11px] text-t3 ml-1">· key</span>
										)}
									</td>
									<td className="px-4 py-3 text-right space-x-2">
										<Button
											type="button"
											variant="outline"
											size="sm"
											onClick={async () => {
												try {
													const r = await testAIProvider(p.id);
													if (r.ok) {
														toast.success(
															`${p.name}: OK (${r.latency_ms}ms, ${r.model})`,
														);
													} else {
														toast.error(r.error ?? "Ping failed");
													}
												} catch (err) {
													toastApiError(err);
												}
											}}
										>
											Test
										</Button>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											onClick={async () => {
												try {
													await updateAIProvider(p.id, {
														enabled: !p.enabled,
													});
													void refresh();
												} catch (err) {
													toastApiError(err);
												}
											}}
										>
											{p.enabled ? "Выкл" : "Вкл"}
										</Button>
									</td>
								</tr>
							))
						)}
					</tbody>
				</table>
			</div>
		</div>
	);
}
