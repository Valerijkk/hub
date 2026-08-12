import { describe, expect, it } from "vitest";
import { computeInsertRank } from "@/lib/board-rank";

describe("FE-P1-BOARD-DND · computeInsertRank (W10-A4-002)", () => {
	it("empty column gets a default rank", () => {
		expect(computeInsertRank([], 0)).toBe("0|hzzzzz:");
	});

	it("append after last produces distinct rank (not constant collision)", () => {
		const issues = [{ rank: "0|aaaaaa:" }, { rank: "0|aaaaab:" }];
		const r1 = computeInsertRank(issues, 2);
		const r2 = computeInsertRank([...issues, { rank: r1 }], 3);
		expect(r1).not.toBe("0|aaaaaa:");
		expect(r2).not.toBe(r1);
		expect(r1 > "0|aaaaab:").toBe(true);
	});

	it("insert between neighbors stays lexicographically between", () => {
		const issues = [{ rank: "0|aaaaaa:" }, { rank: "0|aaaaaz:" }];
		const mid = computeInsertRank(issues, 1);
		expect(mid > "0|aaaaaa:").toBe(true);
		expect(mid < "0|aaaaaz:").toBe(true);
	});

	it("mutation: constant rank would collide on two appends", () => {
		const constant = "0|aaaaaa:";
		expect(constant).toBe(constant);
		const a = computeInsertRank([{ rank: "0|aaaaaa:" }], 1);
		const b = computeInsertRank([{ rank: "0|aaaaaa:" }, { rank: a }], 2);
		expect(a).not.toBe(b);
	});
});
