import { describe, expect, it } from "vitest";
import {
	buildProfileDisplayName,
	validateOnboardingProfileStep,
} from "@/lib/onboarding/profile-schema";

describe("validateOnboardingProfileStep", () => {
	it("accepts valid profile", () => {
		expect(
			validateOnboardingProfileStep({
				name: "Alex",
				surname: "Ivanov",
				position: "Developer",
				department: "Backend",
				timezone: "Europe/Moscow",
				language: "ru",
			}),
		).toBeNull();
	});

	it("rejects short name", () => {
		const errors = validateOnboardingProfileStep({
			name: "A",
			surname: "",
			position: "",
			department: "",
			timezone: "UTC",
			language: "en",
		});
		expect(errors?.name).toBeTruthy();
	});

	it("rejects overlong position", () => {
		const errors = validateOnboardingProfileStep({
			name: "Valid Name",
			surname: "",
			position: "x".repeat(101),
			department: "",
			timezone: "UTC",
			language: "ru",
		});
		expect(errors?.position).toBeTruthy();
	});
});

describe("buildProfileDisplayName", () => {
	it("joins name and surname", () => {
		expect(
			buildProfileDisplayName({
				name: "Alex",
				surname: "Ivanov",
				position: "",
				department: "",
				timezone: "UTC",
				language: "ru",
			}),
		).toBe("Alex Ivanov");
	});

	it("uses name only when surname empty", () => {
		expect(
			buildProfileDisplayName({
				name: "Alex",
				surname: "",
				position: "",
				department: "",
				timezone: "UTC",
				language: "ru",
			}),
		).toBe("Alex");
	});
});
