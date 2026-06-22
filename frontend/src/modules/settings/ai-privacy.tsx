import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { getAIPreferences, updateAIPreferences } from "@/api/ai";
import { BYOKSettings } from "@/components/ai/byok-settings";
import { Button } from "@/components/ui/button";
import { toastApiError } from "@/lib/format-api-error";

/** Приватность AI пользователя + BYOK (M3-030, Phase 18). */
export function UserAISettings() {
	const [allowExternal, setAllowExternal] = useState(true);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState(false);

	const refresh = useCallback(async () => {
		setLoading(true);
		try {
			const pref = await getAIPreferences();
			setAllowExternal(pref.allow_external_ai ?? true);
		} catch (err) {
			toastApiError(err, "Не удалось загрузить настройки AI");
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
			await updateAIPreferences({ allow_external_ai: allowExternal });
			toast.success("Настройки AI сохранены");
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
		<div className="space-y-8 max-w-2xl">
			<div>
				<h2 className="text-[18px] font-extrabold tracking-[-0.02em] mb-1">
					AI и приватность
				</h2>
				<p className="text-[13px] text-t2 mb-5">
					Управление отправкой данных во внешние LLM API (DeepSeek, GigaChat,
					OpenRouter и др.)
				</p>
			</div>

			<div className="rounded-[var(--r3)] border border-b1 bg-gradient-to-b from-card to-raised p-5 space-y-4">
				<label className="flex items-start gap-3 cursor-pointer">
					<input
						type="checkbox"
						checked={allowExternal}
						onChange={(e) => setAllowExternal(e.target.checked)}
						className="mt-1 rounded border-b1"
					/>
					<span className="text-[13px]">
						<span className="font-semibold block mb-1">
							Разрешить облачные AI-провайдеры
						</span>
						<span className="text-t2">
							Если выключено, Hub использует только локальный Ollama/localhost.
							Промпты не отправляются во внешние API без вашего согласия
							(M3-030).
						</span>
					</span>
				</label>
				<Button
					type="button"
					size="sm"
					disabled={saving}
					onClick={() => void handleSave()}
				>
					{saving ? "Сохранение..." : "Сохранить"}
				</Button>
			</div>

			<div>
				<h3 className="text-[15px] font-bold mb-2">Свои ключи (BYOK)</h3>
				<BYOKSettings />
			</div>
		</div>
	);
}
