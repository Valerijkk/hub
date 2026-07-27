import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	getRegistrationMode,
	setRegistrationMode,
} from "@/api/admin-org";
import { api } from "@/api/client";

describe("admin registration-mode API (A11-013)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("getRegistrationMode GETs /admin/registration-mode", async () => {
		const getSpy = vi
			.spyOn(api, "get")
			.mockResolvedValue({ mode: "invite_only" });

		const result = await getRegistrationMode();

		expect(getSpy).toHaveBeenCalledWith("/admin/registration-mode");
		expect(result.mode).toBe("invite_only");
	});

	it("setRegistrationMode PUTs mode body", async () => {
		const putSpy = vi
			.spyOn(api, "put")
			.mockResolvedValue({ mode: "open" });

		const result = await setRegistrationMode("open");

		expect(putSpy).toHaveBeenCalledWith("/admin/registration-mode", {
			mode: "open",
		});
		expect(result.mode).toBe("open");
	});
});
