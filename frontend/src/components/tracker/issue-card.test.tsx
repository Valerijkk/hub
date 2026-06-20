import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { IssueBriefResponse } from "@/api/tracker";
import { IssueCard } from "./issue-card";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		onClick,
		className,
	}: {
		children: React.ReactNode;
		onClick?: (e: React.MouseEvent) => void;
		className?: string;
	}) => (
		<a href="#" onClick={onClick} className={className}>
			{children}
		</a>
	),
}));

vi.mock("@atlaskit/pragmatic-drag-and-drop/element/adapter", () => ({
	draggable: vi.fn(() => vi.fn()),
}));

const baseIssue: IssueBriefResponse = {
	id: "issue-1",
	issue_key: "HUB-42",
	title: "Fix login bug",
	issue_type: "bug",
	priority: "high",
	status: "in_progress",
};

describe("IssueCard", () => {
	it("renders issue key and title", () => {
		render(<IssueCard issue={baseIssue} projectId="proj-1" />);

		expect(screen.getByText("HUB-42")).toBeInTheDocument();
		expect(screen.getByText("Fix login bug")).toBeInTheDocument();
	});

	it("renders story points badge", () => {
		const issue = { ...baseIssue, story_points: 5 };
		render(<IssueCard issue={issue} projectId="proj-1" />);

		expect(screen.getByText("5 SP")).toBeInTheDocument();
	});

	it("renders checklist progress", () => {
		const issue = {
			...baseIssue,
			checklist: [
				{ id: "c1", text: "Step 1", checked: true },
				{ id: "c2", text: "Step 2", checked: false },
			],
		};
		render(<IssueCard issue={issue} projectId="proj-1" />);

		expect(screen.getByText("1/2")).toBeInTheDocument();
	});

	it("renders assignee initial", () => {
		const issue = {
			...baseIssue,
			assignee: {
				id: "u1",
				name: "Alice",
				email: "alice@hub.local",
			},
		};
		render(<IssueCard issue={issue} projectId="proj-1" />);

		expect(screen.getByText("A")).toBeInTheDocument();
	});

	it("shows empty assignee placeholder when unassigned", () => {
		render(<IssueCard issue={baseIssue} projectId="proj-1" />);

		expect(screen.getByText("-")).toBeInTheDocument();
	});
});
