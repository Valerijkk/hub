import { describe, expect, it } from "vitest";
import fs from "node:fs";
import path from "node:path";

/**
 * Guard for W10-A4-001: board must not call moveIssue from both
 * monitorForElements.onDrop and BoardColumn dropTarget onDrop.
 */
describe("FE-P1-BOARD-DND · single moveIssue path (W10-A4-001)", () => {
	it("board page source has no monitorForElements dual mutate", () => {
		const file = path.join(
			process.cwd(),
			"src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx",
		);
		const text = fs.readFileSync(file, "utf8");
		expect(text).not.toContain("monitorForElements");
		expect(text).toContain("computeInsertRank");
		const mutateCount = (text.match(/moveCardMutation\.mutate/g) ?? []).length;
		expect(mutateCount).toBe(1);
	});
});
