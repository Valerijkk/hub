import { z } from "zod";

/** Zod-схема шага 4 онбординга — создание проекта (FB-12). */
export const onboardingCreateProjectSchema = z.object({
	name: z
		.string()
		.trim()
		.min(2, "onboarding:validation.projectNameMin")
		.max(100, "onboarding:validation.projectNameMax"),
	key: z
		.string()
		.trim()
		.min(2, "onboarding:validation.projectKeyMin")
		.max(10, "onboarding:validation.projectKeyMax")
		.regex(/^[A-Z0-9]+$/, "onboarding:validation.projectKeyFormat"),
});

/** Zod-схема шага 4 — присоединение по invite-коду. */
export const onboardingJoinProjectSchema = z.object({
	inviteCode: z
		.string()
		.trim()
		.min(4, "onboarding:validation.inviteCodeMin")
		.max(64, "onboarding:validation.inviteCodeMax"),
});

export type OnboardingCreateProjectValues = z.infer<
	typeof onboardingCreateProjectSchema
>;
export type OnboardingJoinProjectValues = z.infer<
	typeof onboardingJoinProjectSchema
>;

export function validateOnboardingProjectStep(
	mode: "create" | "join",
	values: { projectName: string; projectKey: string; inviteCode: string },
): Record<string, string> | null {
	if (mode === "create") {
		const result = onboardingCreateProjectSchema.safeParse({
			name: values.projectName,
			key: values.projectKey,
		});
		if (result.success) return null;
		const errors: Record<string, string> = {};
		for (const issue of result.error.issues) {
			const field = issue.path[0];
			if (field === "name") errors.projectName = issue.message;
			if (field === "key") errors.projectKey = issue.message;
		}
		return errors;
	}
	const result = onboardingJoinProjectSchema.safeParse({
		inviteCode: values.inviteCode,
	});
	if (result.success) return null;
	return {
		inviteCode: result.error.issues[0]?.message ?? "validation.invalid",
	};
}
