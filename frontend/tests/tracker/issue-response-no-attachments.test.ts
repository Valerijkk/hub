import { describe, expect, it } from "vitest";
import type { IssueResponse } from "@/api/tracker";

describe("IssueResponse attachments field (A11-008)", () => {
	it("does not declare attachments (BE IssueResponse has no such field)", () => {
		type HasAttachments = "attachments" extends keyof IssueResponse
			? true
			: false;
		const hasAttachments: HasAttachments = false;
		expect(hasAttachments).toBe(false);
	});
});
