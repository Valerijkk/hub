import { describe, expect, it } from "vitest";
import { isSafeUrl, sanitizeUrl } from "@/lib/safe-url";

describe("FE · sanitizeUrl (A3W10-2-001)", () => {
	it("allows https, http, mailto, absolute path, hash", () => {
		expect(sanitizeUrl("https://github.com/org/repo/pull/1")).toBe(
			"https://github.com/org/repo/pull/1",
		);
		expect(sanitizeUrl("http://example.com")).toBe("http://example.com");
		expect(sanitizeUrl("mailto:a@b.c")).toBe("mailto:a@b.c");
		expect(sanitizeUrl("/projects/1")).toBe("/projects/1");
		expect(sanitizeUrl("#section")).toBe("#section");
		expect(isSafeUrl("https://ok")).toBe(true);
	});

	it("blocks javascript: and data: schemes → #", () => {
		expect(sanitizeUrl("javascript:alert(1)")).toBe("#");
		expect(sanitizeUrl("data:text/html,<script>alert(1)</script>")).toBe("#");
		expect(sanitizeUrl("vbscript:msgbox(1)")).toBe("#");
		expect(isSafeUrl("javascript:alert(1)")).toBe(false);
	});

	it("rejects non-string → #", () => {
		expect(sanitizeUrl(null)).toBe("#");
		expect(sanitizeUrl(undefined)).toBe("#");
		expect(sanitizeUrl(42)).toBe("#");
		expect(sanitizeUrl({ url: "https://x" })).toBe("#");
	});

	// Mutation: ослабить SAFE_URL_RE (добавить javascript:) → assert на "#" FAIL.
});
