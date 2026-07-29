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
	/** OL для PATCH /ai/preferences (FE-SMA-002). */
	preferenceVersion: number;
	panelOpen: boolean;
	prefillInput: string | null;
	/** true только после успешного loadBotInfo — fail не блокирует retry. */
	loaded: boolean;
	botLoadError: boolean;
	modelsError: boolean;
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
	preferenceVersion: 0,
	panelOpen: false,
	prefillInput: null,
	loaded: false,
	botLoadError: false,
	modelsError: false,
	loadBotInfo: async () => {
		if (get().loaded) return;
		try {
			const info = await getAIBotInfo();
			set({
				botUserId: info.user_id,
				botDisplayName: info.display_name,
				loaded: true,
				botLoadError: false,
			});
		} catch {
			// W11-A2-002: не ставить loaded:true — иначе retry навсегда заблокирован
			set({ botLoadError: true });
		}
	},
	loadModels: async () => {
		try {
			const [resp, pref] = await Promise.all([
				getAIModels(),
				getAIPreferences().catch(() => ({
					default_model: "",
					version: 0,
				})),
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
				preferenceVersion: pref.version ?? 0,
				modelsError: false,
			}));
		} catch {
			// W11-A2-003: fail ≠ «список моделей пуст»
			set({ modelsError: true });
		}
	},
	setSelectedModel: (model) => set({ selectedModel: model }),
	saveSelectedModel: async (model) => {
		const prev = get().selectedModel;
		const version = get().preferenceVersion;
		set({ selectedModel: model });
		try {
			const saved = await updateAIPreferences({
				default_model: model,
				version,
			});
			set({ preferenceVersion: saved.version ?? version });
		} catch {
			// rollback optimistic выбора при fail sync
			set({ selectedModel: prev });
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
