import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, expectTypeOf, it, vi } from "vitest";
import * as aiApi from "@/api/ai";
import type {
	AIUserPreference,
	UpdateAIUserPreferenceRequest,
} from "@/api/ai";
import { UserAISettings } from "@/modules/settings/ai-privacy";

vi.mock("@/components/ai/byok-settings", () => ({
	BYOKSettings: () => <div>BYOK mock</div>,
}));

/**
 * FE-SMA-002 / TASK-192: AI privacy save шлёт version; 409 → toast + refresh.
 * Mutation: убрать version из updateAIPreferences / handleSave → RED.
 */
describe("UserAISettings version OL (FE-SMA-002 / TASK-192)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("AIUserPreference / UpdateAIUserPreferenceRequest expose version", () => {
		expectTypeOf<AIUserPreference>().toHaveProperty("version");
		expectTypeOf<AIUserPreference["version"]>().toEqualTypeOf<number>();
		expectTypeOf<UpdateAIUserPreferenceRequest>().toHaveProperty("version");
		expectTypeOf<UpdateAIUserPreferenceRequest["version"]>().toEqualTypeOf<number>();
	});

	it("save sends allow_external_ai + version from load", async () => {
		vi.spyOn(aiApi, "getAIPreferences").mockResolvedValue({
			allow_external_ai: true,
			version: 2,
		});
		const patchSpy = vi.spyOn(aiApi, "updateAIPreferences").mockResolvedValue({
			allow_external_ai: false,
			version: 3,
		});

		render(<UserAISettings />);

		await waitFor(() => {
			expect(screen.getByRole("checkbox")).toBeChecked();
		});

		fireEvent.click(screen.getByRole("checkbox"));
		fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

		await waitFor(() => {
			expect(patchSpy).toHaveBeenCalledWith({
				allow_external_ai: false,
				version: 2,
			});
		});
	});

	it("409 on save refreshes preferences", async () => {
		const getSpy = vi
			.spyOn(aiApi, "getAIPreferences")
			.mockResolvedValueOnce({
				allow_external_ai: true,
				version: 1,
			})
			.mockResolvedValueOnce({
				allow_external_ai: false,
				version: 4,
			});
		vi.spyOn(aiApi, "updateAIPreferences").mockRejectedValue({
			code: 409,
			message: "version conflict",
			request_id: "req-ol",
		});

		render(<UserAISettings />);

		await waitFor(() => {
			expect(screen.getByRole("checkbox")).toBeChecked();
		});

		fireEvent.click(screen.getByRole("button", { name: "Сохранить" }));

		await waitFor(() => {
			expect(getSpy).toHaveBeenCalledTimes(2);
		});
		await waitFor(() => {
			expect(screen.getByRole("checkbox")).not.toBeChecked();
		});
	});
});
