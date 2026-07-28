import { beforeEach, describe, expect, it, vi } from "vitest";
import * as aiApi from "@/api/ai";
import { useAIStore } from "@/stores/ai-store";

function resetAIStore(): void {
	useAIStore.setState({
		botUserId: null,
		botDisplayName: "@hub-ai",
		models: [],
		selectedModel: null,
		panelOpen: false,
		prefillInput: null,
		loaded: false,
		botLoadError: false,
		modelsError: false,
	});
}

describe("ai-store", () => {
	beforeEach(() => {
		resetAIStore();
		vi.restoreAllMocks();
	});

	it("togglePanel flips panelOpen", () => {
		expect(useAIStore.getState().panelOpen).toBe(false);
		useAIStore.getState().togglePanel();
		expect(useAIStore.getState().panelOpen).toBe(true);
	});

	it("openWithPrefill opens panel and sets text", () => {
		useAIStore.getState().openWithPrefill("Explain this bug");
		const s = useAIStore.getState();
		expect(s.panelOpen).toBe(true);
		expect(s.prefillInput).toBe("Explain this bug");
	});

	it("consumePrefill returns and clears prefill once", () => {
		useAIStore.getState().openWithPrefill("hello");
		expect(useAIStore.getState().consumePrefill()).toBe("hello");
		expect(useAIStore.getState().consumePrefill()).toBeNull();
	});

	it("isBotMessage matches bot user id", async () => {
		vi.spyOn(aiApi, "getAIBotInfo").mockResolvedValue({
			user_id: "bot-uuid",
			display_name: "@hub-ai",
			email: "hub-ai@system.local",
		});
		await useAIStore.getState().loadBotInfo();
		expect(useAIStore.getState().isBotMessage("bot-uuid")).toBe(true);
		expect(useAIStore.getState().isBotMessage("other")).toBe(false);
	});

	it("loadModels prefers saved default_model", async () => {
		vi.spyOn(aiApi, "getAIModels").mockResolvedValue({
			models: [
				{
					id: "llama3",
					provider_id: "p1",
					provider_name: "Ollama",
					ru_resident: false,
				},
				{
					id: "deepseek-chat",
					provider_id: "p2",
					provider_name: "DeepSeek",
					ru_resident: false,
				},
			],
		});
		vi.spyOn(aiApi, "getAIPreferences").mockResolvedValue({
			default_model: "deepseek-chat",
			allow_external_ai: true,
		});
		await useAIStore.getState().loadModels();
		expect(useAIStore.getState().selectedModel).toBe("deepseek-chat");
	});

	it("saveSelectedModel persists preference", async () => {
		const patchSpy = vi.spyOn(aiApi, "updateAIPreferences").mockResolvedValue({
			default_model: "llama3",
			allow_external_ai: true,
		});
		await useAIStore.getState().saveSelectedModel("llama3");
		expect(patchSpy).toHaveBeenCalledWith({ default_model: "llama3" });
		expect(useAIStore.getState().selectedModel).toBe("llama3");
	});

	it("loadBotInfo fail does not set loaded (W11-A2-002)", async () => {
		vi.spyOn(aiApi, "getAIBotInfo").mockRejectedValue(new Error("network"));
		await useAIStore.getState().loadBotInfo();
		const s = useAIStore.getState();
		expect(s.loaded).toBe(false);
		expect(s.botLoadError).toBe(true);
		// retry still possible
		vi.spyOn(aiApi, "getAIBotInfo").mockResolvedValue({
			user_id: "bot-uuid",
			display_name: "@hub-ai",
			email: "hub-ai@system.local",
		});
		await useAIStore.getState().loadBotInfo();
		expect(useAIStore.getState().loaded).toBe(true);
		expect(useAIStore.getState().botLoadError).toBe(false);
	});

	it("loadModels fail sets modelsError (W11-A2-003)", async () => {
		vi.spyOn(aiApi, "getAIModels").mockRejectedValue(new Error("network"));
		await useAIStore.getState().loadModels();
		expect(useAIStore.getState().modelsError).toBe(true);
		expect(useAIStore.getState().models).toEqual([]);
	});

	it("saveSelectedModel rolls back on fail (W11-A2-003)", async () => {
		useAIStore.setState({ selectedModel: "prev-model" });
		vi.spyOn(aiApi, "updateAIPreferences").mockRejectedValue(
			new Error("network"),
		);
		await useAIStore.getState().saveSelectedModel("new-model");
		expect(useAIStore.getState().selectedModel).toBe("prev-model");
	});
});
