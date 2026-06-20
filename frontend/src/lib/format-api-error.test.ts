import { describe, expect, it, vi } from "vitest";

vi.mock("sonner", () => ({
	toast: { error: vi.fn() },
}));

import { toast } from "sonner";
import "@/i18n";
import { formatApiError, toastApiError } from "./format-api-error";

describe("formatApiError", () => {
	it("returns API message when present", () => {
		expect(formatApiError({ message: "WIP limit exceeded", code: 409 })).toBe(
			"WIP limit exceeded",
		);
	});

	it("maps HTTP status to i18n when message empty", () => {
		const msg = formatApiError({ message: "", code: 403 });
		expect(msg.length).toBeGreaterThan(0);
		expect(msg).not.toBe("");
	});

	it("uses Error.message for generic errors", () => {
		expect(formatApiError(new Error("Network failed"))).toBe("Network failed");
	});

	it("falls back to unknown for nullish input", () => {
		const msg = formatApiError(undefined);
		expect(msg.length).toBeGreaterThan(0);
	});

	it("toastApiError delegates to toast.error", () => {
		toastApiError({ message: "Boom", code: 500 });
		expect(toast.error).toHaveBeenCalledWith("Boom");
	});
});
