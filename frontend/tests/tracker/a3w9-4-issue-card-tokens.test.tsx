import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { IssueBriefResponse } from "@/api/tracker";
import { IssueCard } from "@/components/tracker/issue-card";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		className,
	}: {
		children: React.ReactNode;
		className?: string;
	}) => (
		<a href="#" className={className}>
			{children}
		</a>
	),
}));

vi.mock("@atlaskit/pragmatic-drag-and-drop/element/adapter", () => ({
	draggable: vi.fn(() => vi.fn()),
}));

const bugIssue: IssueBriefResponse = {
	id: "issue-1",
	issue_key: "HUB-42",
	title: "Fix login bug",
	issue_type: "bug",
	priority: "high",
	status: "in_progress",
};

describe("IssueCard tokens (A3W9-4-004 / TASK-175)", () => {
	it("type icon uses semantic var(--red), not text-red-500", () => {
		const { container } = render(
			<IssueCard issue={bugIssue} projectId="proj-1" />,
		);

		expect(screen.getByText("HUB-42")).toBeInTheDocument();
		const html = container.innerHTML;
		expect(html).toMatch(/text-\[var\(--red\)\]/);
		expect(html).not.toMatch(/text-red-500/);
		expect(html).not.toMatch(/text-blue-500/);
	});
});
