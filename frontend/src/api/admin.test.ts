import { beforeEach, describe, expect, it, vi } from "vitest";
import { banUser, deactivateUser } from "./admin";
import { api } from "./client";

describe("admin API client", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("banUser posts reason to /admin/users/:id/ban", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue({
			success: true,
			message: "User banned",
		});

		const result = await banUser("user-42", "spam");

		expect(postSpy).toHaveBeenCalledWith("/admin/users/user-42/ban", {
			reason: "spam",
		});
		expect(result.success).toBe(true);
	});

	it("deactivateUser posts to /admin/users/:id/deactivate", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue({
			success: true,
			message: "User deactivated",
		});

		await deactivateUser("user-99");

		expect(postSpy).toHaveBeenCalledWith("/admin/users/user-99/deactivate");
	});
});
