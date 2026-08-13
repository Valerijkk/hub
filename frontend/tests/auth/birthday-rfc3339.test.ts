import { beforeEach, describe, expect, it, vi } from "vitest";
import {
	birthdayForDateInput,
	completeOnboarding,
	toBirthdayRFC3339,
	updateProfile,
} from "@/api/auth";
import { api } from "@/api/client";

describe("birthday RFC3339 contract (A11-006)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("toBirthdayRFC3339 converts date-only to UTC midnight", () => {
		expect(toBirthdayRFC3339("1990-05-15")).toBe("1990-05-15T00:00:00Z");
		expect(toBirthdayRFC3339("1990-05-15T12:00:00Z")).toBe(
			"1990-05-15T12:00:00Z",
		);
		expect(toBirthdayRFC3339(undefined)).toBeUndefined();
		expect(toBirthdayRFC3339("")).toBeUndefined();
	});

	it("birthdayForDateInput slices ISO to YYYY-MM-DD", () => {
		expect(birthdayForDateInput("1990-05-15T00:00:00Z")).toBe("1990-05-15");
		expect(birthdayForDateInput("1990-05-15")).toBe("1990-05-15");
		expect(birthdayForDateInput(null)).toBe("");
	});

	it("updateProfile sends RFC3339 birthday to PUT /auth/me", async () => {
		const putSpy = vi.spyOn(api, "put").mockResolvedValue({} as never);

		await updateProfile({ name: "Ada", birthday: "1990-05-15" });

		expect(putSpy).toHaveBeenCalledWith("/auth/me", {
			name: "Ada",
			birthday: "1990-05-15T00:00:00Z",
		});
	});

	it("completeOnboarding normalizes profile.birthday", async () => {
		const postSpy = vi.spyOn(api, "post").mockResolvedValue({} as never);

		await completeOnboarding({
			profile: { birthday: "2001-01-02" },
		});

		expect(postSpy).toHaveBeenCalledWith("/auth/me/onboarding", {
			profile: { birthday: "2001-01-02T00:00:00Z" },
		});
	});
});
