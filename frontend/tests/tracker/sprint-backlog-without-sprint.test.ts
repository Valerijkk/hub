import { describe, expect, it } from "vitest";
import { filterIssuesWithoutSprint } from "@/api/tracker";

describe("TASK-082 · filterIssuesWithoutSprint (A2W10-A3-001)", () => {
	const mixed = [
		{ id: "1", sprint_id: null as string | null },
		{ id: "2", sprint_id: "sprint-aa" },
		{ id: "3" },
		{ id: "4", sprint_id: "" },
	];

	it("keeps only issues without sprint_id", () => {
		const backlog = filterIssuesWithoutSprint(mixed);
		expect(backlog.map((i) => i.id)).toEqual(["1", "3", "4"]);
	});

	it("mutation: without filter UI would show sprint-assigned issues", () => {
		// Mutation note: убрать filterIssuesWithoutSprint в SprintBacklog →
		// в списке снова появятся id "2" (задача уже в спринте).
		expect(mixed).toHaveLength(4);
		expect(filterIssuesWithoutSprint(mixed)).toHaveLength(3);
		expect(mixed.some((i) => i.sprint_id === "sprint-aa")).toBe(true);
	});
});
