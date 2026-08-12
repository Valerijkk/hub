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

	it("gantt-chart wraps issues query in QueryStateView (A10-011)", () => {
		const src = readFileSync(ganttPath, "utf8");
		expect(src).toContain("QueryStateView");
		expect(src).toContain("issuesQuery");
	});

	it("gantt-chart virtualizes rows at threshold and uses cursor pages (A10-014)", () => {
		const src = readFileSync(ganttPath, "utf8");
		expect(src).toContain("useInfiniteQuery");
		expect(src).toContain("useVirtualizer");
		expect(src).toContain("GANTT_VIRTUALIZE_THRESHOLD");
		expect(src).toMatch(/GANTT_VIRTUALIZE_THRESHOLD\s*=\s*24/);
		expect(src).not.toMatch(/listIssues\([^)]*\{\s*limit:\s*100[^}]*\}\s*\)/);
	});

	it("exposes listProjectIssueLinks API helper", () => {
		expect(typeof listProjectIssueLinks).toBe("function");
	});
});
