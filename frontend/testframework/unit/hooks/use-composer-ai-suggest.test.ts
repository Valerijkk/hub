import { describe, expect, it } from "vitest";
import {
	applyComposerSuggestion,
	canSuggestComposer,
} from "@/hooks/use-composer-ai-suggest";

describe("use-composer-ai-suggest helpers", () => {
	it("canSuggestComposer requires min length", () => {
		expect(canSuggestComposer("short")).toBe(false);
		expect(canSuggestComposer("12345678")).toBe(true);
		expect(canSuggestComposer("  hello world  ")).toBe(true);
	});

	it("applyComposerSuggestion adds space when needed", () => {
		expect(applyComposerSuggestion("Привет", "команда")).toBe("Привет команда");
		expect(applyComposerSuggestion("Привет ", "команда")).toBe(
			"Привет команда",
		);
	});
});
