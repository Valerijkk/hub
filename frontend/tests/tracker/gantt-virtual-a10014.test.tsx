import { screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { IssueResponse } from "@/api/tracker";
import {
	GANTT_VIRTUALIZE_THRESHOLD,
	GanttChart,
} from "@/modules/tracker/gantt-chart";
import { renderWithProviders } from "../support/render";
import { querySuccess } from "../support/query-result";

/** Mirrors `GANTT_VIRTUALIZE_THRESHOLD` export (A10-014). */
const VIRTUALIZE_THRESHOLD = GANTT_VIRTUALIZE_THRESHOLD;

vi.mock("@/api/tracker", () => ({
	listIssues: vi.fn(),
	listProjectIssueLinks: vi.fn(),
}));

vi.mock("@tanstack/react-query", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-query")>(
		"@tanstack/react-query",
	);
	return {
		...actual,
		useQuery: vi.fn(),
		useInfiniteQuery: vi.fn(),
	};
});

vi.mock("@tanstack/react-router", () => ({
	Link: ({
		children,
		...props
	}: {
		children: React.ReactNode;
		to?: string;
		params?: Record<string, string>;
		className?: string;
	}) => (
		<a href={props.to ?? "#"} className={props.className}>
			{children}
		</a>
	),
}));

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";

function makeIssue(i: number): IssueResponse {
	const day = String(i).padStart(2, "0");
	return {
		id: `00000000-0000-4000-8000-${String(i).padStart(12, "0")}`,
		issue_number: i,
		issue_key: `HUB-${i}`,
		title: `Task ${i}`,
		issue_type: "task",
		priority: "medium",
		version: 1,
		time_spent_minutes: 0,
		children_count: 0,
		labels: [],
		status: {
			id: "st-1",
			name: "Todo",
			color: "#888",
			category: "todo",
			sort_order: 0,
			is_default: true,
		},
		due_date: `2026-08-${day.length > 2 ? "28" : day}T00:00:00Z`,
		created_at: "2026-07-01T00:00:00Z",
		updated_at: "2026-07-01T00:00:00Z",
	};
}

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
			return 56;
		},
	});
}

function mockIssues(items: IssueResponse[], total = items.length) {
	vi.mocked(useInfiniteQuery).mockReturnValue({
		data: {
			pages: [{ items, total, has_more: false }],
			pageParams: [undefined],
		},
		isLoading: false,
		isError: false,
		refetch: vi.fn(),
		hasNextPage: false,
		isFetchingNextPage: false,
		fetchNextPage: vi.fn(),
	} as unknown as ReturnType<typeof useInfiniteQuery>);
}

describe("GanttChart virtualization (A10-014)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		stubScrollMetrics();
		vi.mocked(useQuery).mockReturnValue(querySuccess([]));
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("renders all rows when below virtualize threshold", () => {
		const items = Array.from({ length: 3 }, (_, i) => makeIssue(i + 1));
		mockIssues(items);

		const { container } = renderWithProviders(
			<GanttChart projectId="proj-1" />,
		);

		expect(screen.getByText("HUB-1")).toBeInTheDocument();
		expect(screen.getByText("HUB-3")).toBeInTheDocument();
		expect(screen.getByText(/3 задач с дедлайном/)).toBeInTheDocument();
		expect(container.querySelectorAll("[data-index]")).toHaveLength(0);
		expect(VIRTUALIZE_THRESHOLD).toBe(24);
	});

	it("mounts virtual path when bars >= GANTT_VIRTUALIZE_THRESHOLD", () => {
		const count = VIRTUALIZE_THRESHOLD;
		const items = Array.from({ length: count }, (_, i) => makeIssue(i + 1));
		mockIssues(items);

		const { container } = renderWithProviders(
			<GanttChart projectId="proj-1" />,
		);

		const list = screen.getByTestId("gantt-bars-list");
		expect(list).toBeInTheDocument();
		expect(screen.getByText(new RegExp(`${count} задач с дедлайном`))).toBeInTheDocument();

		const virtualRows = container.querySelectorAll("[data-index]");
		expect(virtualRows.length).toBeGreaterThan(0);
		expect(virtualRows.length).toBeLessThan(count);

		const spacer = Array.from(
			container.querySelectorAll<HTMLElement>("div.relative.w-full"),
		).find((el) => el.style.height && Number.parseFloat(el.style.height) > 0);
		expect(spacer).toBeTruthy();
		expect(Number.parseFloat(spacer?.style.height ?? "0")).toBeGreaterThan(
			count * 40,
		);

		expect(within(list).queryByText(`HUB-${count}`)).not.toBeInTheDocument();
		expect(screen.getByText("HUB-1")).toBeInTheDocument();
	});

	it("shows truncated caption when loaded < total (cursor pages)", () => {
		const items = Array.from({ length: 2 }, (_, i) => makeIssue(i + 1));
		mockIssues(items, 250);

		renderWithProviders(<GanttChart projectId="proj-1" />);

		expect(screen.getByText(/показано 2 из 250/)).toBeInTheDocument();
	});
});
