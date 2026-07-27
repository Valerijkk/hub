import { describe, expect, it } from "vitest";
import type { IssueResponse } from "@/api/tracker";
import type { PaginatedResponse } from "@/api/types";
import {
	flattenIssuePages,
	nextIssuesCursor,
} from "@/modules/tracker/issues-cursor";

function page(
	items: Array<Pick<IssueResponse, "id">>,
	opts: { has_more?: boolean; next_cursor?: string } = {},
): PaginatedResponse<IssueResponse> {
	return {
		items: items as IssueResponse[],
		total: items.length,
		has_more: opts.has_more,
		next_cursor: opts.next_cursor,
	};
}

describe("issues cursor helpers (A10-002)", () => {
	it("flattenIssuePages merges all loaded pages (no off-by-one)", () => {
		const pages = [
			page([{ id: "1" }, { id: "2" }], {
				has_more: true,
				next_cursor: "c2",
			}),
			page([{ id: "3" }], { has_more: false }),
		];

		expect(flattenIssuePages(pages).map((i) => i.id)).toEqual([
			"1",
			"2",
			"3",
		]);
	});

	it("flattenIssuePages returns [] for undefined/empty", () => {
		expect(flattenIssuePages(undefined)).toEqual([]);
		expect(flattenIssuePages([])).toEqual([]);
	});

	it("nextIssuesCursor returns cursor only when has_more", () => {
		expect(
			nextIssuesCursor(
				page([{ id: "1" }], { has_more: true, next_cursor: "abc" }),
			),
		).toBe("abc");
		expect(
			nextIssuesCursor(
				page([{ id: "1" }], { has_more: false, next_cursor: "abc" }),
			),
		).toBeUndefined();
		expect(
			nextIssuesCursor(page([{ id: "1" }], { has_more: true })),
		).toBeUndefined();
	});
});
