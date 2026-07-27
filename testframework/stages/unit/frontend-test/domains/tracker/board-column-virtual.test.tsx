import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		...props
	}: {
		children?: React.ReactNode;
		to?: string;
		params?: Record<string, string>;
		"aria-label"?: string;
		className?: string;
	}) => (
		<a href={props.to ?? "#"} aria-label={props["aria-label"]}>
			{children}
		</a>
	),
}));

vi.mock("@atlaskit/pragmatic-drag-and-drop/element/adapter", () => ({
	dropTargetForElements: () => () => {},
}));

vi.mock("@/components/tracker/issue-card", () => ({
	IssueCard: ({ issue }: { issue: { issue_key: string; title: string } }) => (
		<div data-testid={`issue-${issue.issue_key}`}>{issue.title}</div>
	),
}));

import { BoardColumn } from "@/components/tracker/board-column";
import type { IssueBriefResponse, WorkflowStatusResponse } from "@/api/tracker";

function makeIssue(i: number): IssueBriefResponse {
	return {
		id: `00000000-0000-4000-8000-${String(i).padStart(12, "0")}`,
		issue_key: `HUB-${i}`,
		title: `Task ${i}`,
		issue_type: "task",
		priority: "medium",
		status: "todo",
		rank: `0|${i}`,
	} as IssueBriefResponse;
}

const status: WorkflowStatusResponse = {
	id: "status-1",
	name: "Todo",
	color: "#888",
	category: "todo",
	sort_order: 0,
} as WorkflowStatusResponse;

describe("BoardColumn virtualization (A13-014)", () => {
	it("renders all cards when below virtualize threshold", () => {
		const issues = Array.from({ length: 3 }, (_, i) => makeIssue(i + 1));
		render(
			<BoardColumn
				status={status}
				issues={issues}
				projectId="proj-1"
				onDrop={() => {}}
			/>,
		);
		expect(screen.getByText("Task 1")).toBeInTheDocument();
		expect(screen.getByText("Task 3")).toBeInTheDocument();
		expect(screen.getByText("3")).toBeInTheDocument();
	});

	it("still shows column count when many cards (virtualized path)", () => {
		const issues = Array.from({ length: 30 }, (_, i) => makeIssue(i + 1));
		render(
			<BoardColumn
				status={status}
				issues={issues}
				projectId="proj-1"
				onDrop={() => {}}
			/>,
		);
		expect(screen.getByText("Todo")).toBeInTheDocument();
		expect(screen.getByText("30")).toBeInTheDocument();
	});
});
