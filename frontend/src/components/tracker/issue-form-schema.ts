import { z } from "zod";

/** Zod-схема формы задачи (FB-11): title 3–200, type required. */
export const issueFormSchema = z.object({
	title: z
		.string()
		.trim()
		.min(3, "form.titleMinLength")
		.max(200, "form.titleMaxLength"),
	issue_type: z.enum([
		"bug",
		"task",
		"story",
		"epic",
		"sub_task",
		"improvement",
		"tech_debt",
		"spike",
		"release",
	]),
	priority: z.enum(["critical", "high", "medium", "low", "trivial"]),
	assignee_id: z.string().optional(),
	estimate_minutes: z
		.string()
		.optional()
		.refine(
			(v) => !v || (!Number.isNaN(Number(v)) && Number(v) >= 0),
			"form.estimateInvalid",
		),
	story_points: z.string().optional(),
	due_date: z.string().optional(),
});

export type IssueFormValues = z.infer<typeof issueFormSchema>;
