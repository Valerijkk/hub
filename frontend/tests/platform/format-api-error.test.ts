import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { toastError, toastDismiss } = vi.hoisted(() => ({
	toastError: vi.fn(),
	toastDismiss: vi.fn(),
}));

vi.mock("sonner", () => ({
	toast: {
		error: toastError,
		dismiss: toastDismiss,
	},
}));

import "@/i18n";
import {
	formatApiError,
	formatRateLimitMessage,
	resetRateLimitToastForTests,
	toastApiError,
} from "@/lib/format-api-error";

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

	it("formats 429 with retry_after seconds", () => {
		const msg = formatApiError({
			message: "too many",
			code: 429,
			retry_after: 42,
		});
		expect(msg).toContain("42");
	});

	it("uses Error.message for generic errors", () => {
		expect(formatApiError(new Error("Network failed"))).toBe("Network failed");
	});

	it("falls back to unknown for nullish input", () => {
		const msg = formatApiError(undefined);
		expect(msg.length).toBeGreaterThan(0);
	});

	it("formatRateLimitMessage includes seconds", () => {
		expect(formatRateLimitMessage(15)).toContain("15");
	});
});

describe("toastApiError", () => {
	beforeEach(() => {
		vi.useFakeTimers();
		toastError.mockClear();
		toastDismiss.mockClear();
	});

	afterEach(() => {
		resetRateLimitToastForTests();
		vi.useRealTimers();
	});

	it("delegates to toast.error for generic errors", () => {
		toastApiError({ message: "Boom", code: 500 });
		expect(toastError).toHaveBeenCalledWith("Boom");
	});

	it("shows request_id in description for non-429 errors", () => {
		toastApiError({
			message: "Forbidden",
			code: 403,
			request_id: "req-abc-123",
		});
		expect(toastError).toHaveBeenCalledWith(
			"Forbidden",
			expect.objectContaining({
				description: expect.stringContaining("req-abc-123"),
			}),
		);
	});

	it("starts countdown toast for 429 with retry_after", () => {
		toastApiError({
			message: "too many",
			code: 429,
			retry_after: 3,
			request_id: "req-rate",
		});

		expect(toastError).toHaveBeenCalledWith(
			expect.stringContaining("3"),
			expect.objectContaining({
				id: "hub-api-rate-limit",
				description: expect.stringContaining("req-rate"),
			}),
		);

		vi.advanceTimersByTime(1000);
		expect(toastError).toHaveBeenCalledWith(
			expect.stringContaining("2"),
			expect.objectContaining({ id: "hub-api-rate-limit" }),
		);

		vi.advanceTimersByTime(2000);
		expect(toastDismiss).toHaveBeenCalledWith("hub-api-rate-limit");
	});
});
