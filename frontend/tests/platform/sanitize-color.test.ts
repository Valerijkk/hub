import { describe, expect, it } from "vitest";
import { sanitizeHexColor } from "@/lib/sanitize-color";

describe("sanitizeHexColor", () => {
	it("accepts valid hex", () => {
		expect(sanitizeHexColor("#ff00aa")).toBe("#ff00aa");
	});

	it("rejects injection attempts", () => {
		expect(sanitizeHexColor("red; position:fixed")).toBe("#808080");
		expect(sanitizeHexColor("#abc")).toBe("#808080");
	});

	it("uses fallback when empty", () => {
		expect(sanitizeHexColor(undefined, "#111111")).toBe("#111111");
	});
});
