import { beforeEach, describe, expect, it, vi } from "vitest";
import { suggestComposerCompletion } from "@/api/ai";
import { api } from "@/api/client";

describe("ai API client", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("suggestComposerCompletion posts prefix to /ai/composer/suggest", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue({
			suggestion: " — давай созвонимся",
			model: "llama3",
		});

		const result = await suggestComposerCompletion({
			prefix: "Привет, команда",
			channel_id: "ch-1",
		});

		expect(postSpy).toHaveBeenCalledWith("/ai/composer/suggest", {
			prefix: "Привет, команда",
			channel_id: "ch-1",
		});
		expect(result.suggestion).toContain("созвонимся");
		expect(result.model).toBe("llama3");
	});
});
