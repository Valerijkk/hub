import { render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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

/** Mirrors `VIRTUALIZE_THRESHOLD` in board-column.tsx (A13-014). */
const VIRTUALIZE_THRESHOLD = 24;

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

/** jsdom reports 0 scroll height — stub so @tanstack/react-virtual mounts rows. */
function stubScrollMetrics() {
	Object.defineProperty(HTMLElement.prototype, "clientHeight", {
		configurable: true,
		get() {
			return 480;
		},
	});
	Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
		configurable: true,
		get() {
			return 96;
		},
	});
}

describe("BoardColumn virtualization (A13-014)", () => {
	beforeEach(() => {
		stubScrollMetrics();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("renders all cards when below virtualize threshold", () => {
		const issues = Array.from({ length: 3 }, (_, i) => makeIssue(i + 1));
		const { container } = render(
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
		expect(screen.getAllByTestId(/^issue-/)).toHaveLength(3);
		expect(container.querySelectorAll("[data-index]")).toHaveLength(0);
	});

	it("mounts virtual path when issues >= VIRTUALIZE_THRESHOLD", () => {
		const count = VIRTUALIZE_THRESHOLD;
		const issues = Array.from({ length: count }, (_, i) => makeIssue(i + 1));
		const { container } = render(
			<BoardColumn
				status={status}
				issues={issues}
				projectId="proj-1"
				onDrop={() => {}}
			/>,
		);

		expect(screen.getByText("Todo")).toBeInTheDocument();
		expect(screen.getByText(String(count))).toBeInTheDocument();

		const virtualRows = container.querySelectorAll("[data-index]");
		expect(virtualRows.length).toBeGreaterThan(0);
		expect(virtualRows.length).toBeLessThan(count);

		const spacer = Array.from(
			container.querySelectorAll<HTMLElement>("div.relative.w-full"),
		).find((el) => el.style.height && Number.parseFloat(el.style.height) > 0);
		expect(spacer).toBeTruthy();
		expect(Number.parseFloat(spacer?.style.height ?? "0")).toBeGreaterThan(
			count * 50,
		);

		expect(screen.getAllByTestId(/^issue-/).length).toBeLessThan(count);
		expect(
			within(container).queryByText(`Task ${count}`),
		).not.toBeInTheDocument();
	});
});
