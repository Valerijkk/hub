import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { IssueAttachments } from "@/components/tracker/issue-attachments";

vi.mock("@/api/files", () => ({
	uploadFile: vi.fn(),
}));

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("FE · IssueAttachments sanitizeUrl (FE-TA-002 / TASK-187)", () => {
	it("mutation: javascript: att.url → нет <a href>", () => {
		render(
			<IssueAttachments
				projectId="proj-1"
				issueId="issue-1"
				attachments={[
					{
						file_id: "f-1",
						file_name: "evil.txt",
						file_size: 10,
						mime_type: "text/plain",
						url: "javascript:alert(1)",
					},
				]}
			/>,
		);

		expect(screen.getByText("evil.txt")).toBeInTheDocument();
		expect(document.querySelector('a[href^="javascript:"]')).toBeNull();
		expect(screen.queryByRole("link")).toBeNull();
		// Mutation: href={att.url} → assert FAIL.
	});

	it("keeps relative /api/files download href", () => {
		render(
			<IssueAttachments
				projectId="proj-1"
				issueId="issue-1"
				attachments={[
					{
						file_id: "f-2",
						file_name: "ok.pdf",
						file_size: 2048,
						mime_type: "application/pdf",
					},
				]}
			/>,
		);

		const link = screen.getByRole("link");
		expect(link.getAttribute("href")).toBe("/api/files/f-2/download");
	});
});
