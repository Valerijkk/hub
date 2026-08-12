import { beforeEach, describe, expect, it, vi } from "vitest";
import { getActiveCall } from "@/api/calls";
import { api } from "@/api/client";

vi.mock("@/api/client", () => ({
	api: {
		get: vi.fn(),
		post: vi.fn(),
	},
}));

describe("getActiveCall error mapping (W11-A3-002)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("returns null for 404 / empty call", async () => {
		vi.mocked(api.get).mockRejectedValue({ code: 404, message: "not found" });
		await expect(getActiveCall("ch-1")).resolves.toBeNull();

		vi.mocked(api.get).mockResolvedValue({ call: null });
		await expect(getActiveCall("ch-1")).resolves.toBeNull();
	});

	it("rethrows non-404 errors instead of masking as null", async () => {
		vi.mocked(api.get).mockRejectedValue({
			code: 403,
			message: "forbidden",
		});
		await expect(getActiveCall("ch-1")).rejects.toMatchObject({ code: 403 });
	});
});
