import { z } from "zod";

/** Zod-схема шага 1 онбординга — профиль (UX-P1-07 / FB-12). */
export const onboardingProfileSchema = z.object({
	name: z
		.string()
		.trim()
		.min(2, "onboarding:validation.nameMin")
		.max(100, "onboarding:validation.nameMax"),
	surname: z
		.string()
		.trim()
		.max(100, "onboarding:validation.surnameMax")
		.optional()
		.or(z.literal("")),
	position: z
		.string()
		.trim()
		.max(100, "onboarding:validation.positionMax")
		.optional()
		.or(z.literal("")),
	department: z
		.string()
		.trim()
		.max(100, "onboarding:validation.departmentMax")
		.optional()
		.or(z.literal("")),
	language: z.enum(["ru", "en"]),
});

export type OnboardingProfileValues = z.infer<typeof onboardingProfileSchema>;

/** Собирает display name из имени и фамилии для API. */
export function buildProfileDisplayName(
	values: OnboardingProfileValues,
): string {
	const parts = [values.name.trim(), values.surname?.trim() ?? ""].filter(
		Boolean,
	);
	return parts.join(" ");
}

export function validateOnboardingProfileStep(
	values: OnboardingProfileValues,
): Record<string, string> | null {
	const result = onboardingProfileSchema.safeParse(values);
	if (result.success) return null;
	const errors: Record<string, string> = {};
	for (const issue of result.error.issues) {
		const field = issue.path[0];
		if (typeof field === "string") {
			errors[field] = issue.message;
		}
	}
	return errors;
}
