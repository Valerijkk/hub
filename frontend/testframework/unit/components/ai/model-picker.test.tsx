import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useAIStore } from "@/stores/ai-store";
import { ModelPicker } from "@/components/ai/model-picker";

vi.mock("@/stores/ai-store", () => ({
	useAIStore: vi.fn(),
}));

describe("ModelPicker", () => {
	const loadModels = vi.fn();

	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows hint when no models configured", () => {
		vi.mocked(useAIStore).mockImplementation((selector) =>
			selector({
				models: [],
				selectedModel: null,
				saveSelectedModel: vi.fn(),
				loadModels,
			} as unknown as ReturnType<typeof useAIStore.getState>),
		);
		render(<ModelPicker />);
		expect(screen.getByText(/Нет моделей/)).toBeInTheDocument();
		expect(loadModels).toHaveBeenCalled();
	});

	it("renders model options when registry has entries", () => {
		vi.mocked(useAIStore).mockImplementation((selector) =>
			selector({
				models: [
					{
						id: "llama3",
						provider_id: "p1",
						provider_name: "Ollama",
						ru_resident: false,
					},
					{
						id: "GigaChat",
						provider_id: "p2",
						provider_name: "GigaChat",
						ru_resident: true,
					},
				],
				selectedModel: "llama3",
				saveSelectedModel: vi.fn(),
				loadModels,
			} as unknown as ReturnType<typeof useAIStore.getState>),
		);
		render(<ModelPicker />);
		expect(screen.getByRole("combobox")).toBeInTheDocument();
	});
});
