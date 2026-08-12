import { describe, expect, it } from "vitest";
import { filterBoardIssue } from "@/lib/board-filters";

const mine = {
	title: "My bug",
	issue_key: "HUB-1",
	priority: "high",
	issue_type: "bug",
	assignee: { id: "user-me" },
};

const theirs = {
	title: "Other",
	issue_key: "HUB-2",
	priority: "low",
	issue_type: "task",
	assignee: { id: "user-other" },
};

describe("filterBoardIssue myTasks (P2-FE-06 / TASK-050)", () => {
	it("without myTasks keeps all issues", () => {
		expect(filterBoardIssue(mine, {}, "user-me")).toBe(true);
		expect(filterBoardIssue(theirs, {}, "user-me")).toBe(true);
	});

	it("myTasks hides issues not assigned to current user", () => {
		expect(filterBoardIssue(mine, { myTasks: true }, "user-me")).toBe(true);
		expect(filterBoardIssue(theirs, { myTasks: true }, "user-me")).toBe(false);
	});

	it("mutation: without myTasks branch foreign issues stay visible", () => {
		// Guard: if myTasks were ignored, this would pass for theirs — regression target.
		const foreignVisibleWithoutFilter = filterBoardIssue(
			theirs,
			{ myTasks: false },
			"user-me",
		);
		expect(foreignVisibleWithoutFilter).toBe(true);
		expect(filterBoardIssue(theirs, { myTasks: true }, "user-me")).toBe(false);
	});
});
