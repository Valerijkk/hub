import { create } from "zustand";
import {
	type AIModelInfo,
	getAIBotInfo,
	getAIModels,
	getAIPreferences,
	updateAIPreferences,
} from "@/api/ai";

interface AIStore {
	botUserId: string | null;
	botDisplayName: string;
	models: AIModelInfo[];
	selectedModel: string | null;
	panelOpen: boolean;
	prefillInput: string | null;
	loaded: boolean;
	loadBotInfo: () => Promise<void>;
	loadModels: () => Promise<void>;
	setSelectedModel: (model: string) => void;
	saveSelectedModel: (model: string) => Promise<void>;
	setPanelOpen: (open: boolean) => void;
	togglePanel: () => void;
	openWithPrefill: (text: string) => void;
	consumePrefill: () => string | null;
	isBotMessage: (senderId: string) => boolean;
}

export const useAIStore = create<AIStore>((set, get) => ({
	botUserId: null,
	botDisplayName: "@hub-ai",
	models: [],
	selectedModel: null,
	panelOpen: false,
	prefillInput: null,
	loaded: false,
	loadBotInfo: async () => {
		if (get().loaded) return;
		try {
			const info = await getAIBotInfo();
			set({
				botUserId: info.user_id,
				botDisplayName: info.display_name,
				loaded: true,
			});
		} catch {
			set({ loaded: true });
		}
	},
	loadModels: async () => {
		try {
			const [resp, pref] = await Promise.all([
				getAIModels(),
				getAIPreferences().catch(() => ({ default_model: "" })),
			]);
			const preferred =
				pref.default_model &&
				resp.models.some((m) => m.id === pref.default_model)
					? pref.default_model
					: null;
			const first = resp.models[0]?.id ?? null;
			set((s) => ({
				models: resp.models,
				selectedModel: s.selectedModel ?? preferred ?? first,
			}));
		} catch {
			/* Админ мог ещё не настроить провайдеров — UI остаётся с пустым списком моделей. */
		}
	},
	setSelectedModel: (model) => set({ selectedModel: model }),
	saveSelectedModel: async (model) => {
		set({ selectedModel: model });
		try {
			await updateAIPreferences({ default_model: model });
		} catch {
			/* Синхронизация предпочтений best-effort: локальный выбор важнее серверного ответа. */
		}
	},
	setPanelOpen: (open) => set({ panelOpen: open }),
	togglePanel: () => set((s) => ({ panelOpen: !s.panelOpen })),
	openWithPrefill: (text) => set({ panelOpen: true, prefillInput: text }),
	consumePrefill: () => {
		const text = get().prefillInput;
		if (text) {
			set({ prefillInput: null });
		}
		return text;
	},
	isBotMessage: (senderId: string) => {
		const id = get().botUserId;
		return id !== null && id === senderId;
	},
}));
