import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as aiApi from "@/api/ai";
import { UserAISettings } from "@/modules/settings/ai-privacy";

vi.mock("@/components/ai/byok-settings", () => ({
	BYOKSettings: () => <div>BYOK mock</div>,
}));

describe("UserAISettings", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("loads and toggles external AI consent", async () => {
		vi.spyOn(aiApi, "getAIPreferences").mockResolvedValue({
			allow_external_ai: true,
		});
		const patchSpy = vi.spyOn(aiApi, "updateAIPreferences").mockResolvedValue({
			allow_external_ai: false,
		});

		render(<UserAISettings />);

		await waitFor(() => {
			expect(screen.getByRole("checkbox")).toBeChecked();
		});

		fireEvent.click(screen.getByRole("checkbox"));
		fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

		expect(patchSpy).toHaveBeenCalledWith({ allow_external_ai: false });
	});
});
