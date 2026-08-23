import { describe, expect, it } from "vitest";
import { validateOnboardingProjectStep } from "@/lib/onboarding/project-schema";

describe("validateOnboardingProjectStep", () => {
	it("accepts valid create project", () => {
		expect(
			validateOnboardingProjectStep("create", {
				projectName: "My Team",
				projectKey: "MT",
				inviteCode: "",
			}),
		).toBeNull();
	});

	it("rejects short project name", () => {
		const errors = validateOnboardingProjectStep("create", {
			projectName: "A",
			projectKey: "AB",
			inviteCode: "",
		});
		expect(errors?.projectName).toBeTruthy();
	});

	it("rejects invalid project key", () => {
		const errors = validateOnboardingProjectStep("create", {
			projectName: "Valid Name",
			projectKey: "ab",
			inviteCode: "",
		});
		expect(errors?.projectKey).toBeTruthy();
	});

	it("accepts valid invite code", () => {
		expect(
			validateOnboardingProjectStep("join", {
				projectName: "",
				projectKey: "",
				inviteCode: "abcd1234",
			}),
		).toBeNull();
	});

	it("rejects short invite code", () => {
		const errors = validateOnboardingProjectStep("join", {
			projectName: "",
			projectKey: "",
			inviteCode: "ab",
		});
		expect(errors?.inviteCode).toBeTruthy();
	});
});
