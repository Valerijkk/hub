import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import {
	type AIProvider,
	createMyAIProvider,
	listMyAIProviders,
} from "@/api/ai-my";
import { QueryErrorAlert } from "@/components/query-state-view";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toastApiError } from "@/lib/format-api-error";

/** BYOK — AI-провайдеры пользователя (Phase 17 W4). */
export function BYOKSettings({ onChanged }: { onChanged?: () => void }) {
	const [providers, setProviders] = useState<AIProvider[]>([]);
	const [loading, setLoading] = useState(true);
	const [loadError, setLoadError] = useState(false);
	const [name, setName] = useState("");
	const [baseURL, setBaseURL] = useState("");
	const [apiKey, setApiKey] = useState("");
	const [models, setModels] = useState("");

	const refresh = useCallback(async () => {
		setLoading(true);
		setLoadError(false);
		try {
			setProviders(await listMyAIProviders());
		} catch (err) {
			setLoadError(true);
			toastApiError(err, "Не удалось загрузить BYOK провайдеры");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		void refresh();
	}, [refresh]);

	const handleAdd = async () => {
		try {
			await createMyAIProvider({
				name: name.trim(),
				provider_type: "openai_compatible",
				base_url: baseURL.trim(),
				api_key: apiKey.trim() || undefined,
				models: models
					.split(",")
					.map((m) => m.trim())
					.filter(Boolean),
			});
			toast.success("BYOK провайдер добавлен");
			setName("");
			setApiKey("");
			void refresh();
			onChanged?.();
		} catch (err) {
			toastApiError(err);
		}
	};

	return (
		<div className="space-y-2 text-[12px]">
			<p className="font-semibold text-[13px]">BYOK — ваш API ключ</p>
			{loading && (
				<p className="text-t3" aria-busy="true">
					Загрузка…
				</p>
			)}
			{!loading && loadError && (
				<QueryErrorAlert
					compact
					message="Не удалось загрузить BYOK провайдеры"
					retry={() => void refresh()}
				/>
			)}
			{!loading && !loadError && providers.length > 0 && (
				<ul className="text-t3 space-y-1">
					{providers.map((p) => (
						<li key={p.id}>
							{p.name} · {p.models.join(", ") || "—"}
						</li>
					))}
				</ul>
			)}
			{!loading && !loadError && providers.length === 0 && (
				<p className="text-t3">Нет BYOK провайдеров</p>
			)}
			<Input
				placeholder="Название"
				value={name}
				onChange={(e) => setName(e.target.value)}
				className="h-8 text-xs"
			/>
			<Input
				placeholder="Base URL"
				value={baseURL}
				onChange={(e) => setBaseURL(e.target.value)}
				className="h-8 text-xs"
			/>
			<Input
				type="password"
				placeholder="API Key"
				value={apiKey}
				onChange={(e) => setApiKey(e.target.value)}
				className="h-8 text-xs"
			/>
			<Input
				placeholder="models через запятую"
				value={models}
				onChange={(e) => setModels(e.target.value)}
				className="h-8 text-xs"
			/>
			<Button
				type="button"
				size="sm"
				disabled={!name.trim() || !baseURL.trim()}
				onClick={() => void handleAdd()}
			>
				Добавить BYOK
			</Button>
		</div>
	);
}
