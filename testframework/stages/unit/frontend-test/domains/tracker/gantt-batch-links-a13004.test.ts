import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { listProjectIssueLinks } from "@/api/tracker";

/**
 * A13-004: Gantt must batch project links (1 HTTP), not N× listIssueLinks per row.
 * Mutation: re-add listIssueLinks useQuery in GanttRow → FAIL.
 */
describe("gantt batch links (A13-004)", () => {
	const here = dirname(fileURLToPath(import.meta.url));
	const repoRoot = join(here, "..", "..", "..", "..", "..", "..");
	const ganttPath = join(
		repoRoot,
		"frontend",
		"src",
		"modules",
		"tracker",
		"gantt-chart.tsx",
	);

	it("gantt-chart uses listProjectIssueLinks and not per-row listIssueLinks", () => {
		const src = readFileSync(ganttPath, "utf8");
		expect(src).toContain("listProjectIssueLinks");
		expect(src).toContain("project-issue-links");
		expect(src).not.toMatch(/listIssueLinks\s*\(/);
		expect(src).not.toMatch(/queryKey:\s*\[["']issue-links["']/);
	});

	it("exposes listProjectIssueLinks API helper", () => {
		expect(typeof listProjectIssueLinks).toBe("function");
	});
});
