import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SprintBacklog } from "@/modules/tracker/sprint-backlog";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../support/render";

const listIssues = vi.fn();
const moveIssueToSprint = vi.fn().mockResolvedValue(undefined);

vi.mock("@/api/tracker", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tracker")>();
	return {
		...actual,
		listIssues: (...args: unknown[]) => listIssues(...args),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

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

function issue(partial: {
	id: string;
	issue_key: string;
	title: string;
	sprint_id?: string | null;
}) {
	return {
		id: partial.id,
		issue_number: 1,
		issue_key: partial.issue_key,
		title: partial.title,
		issue_type: "task",
		priority: "medium",
		version: 1,
		time_spent_minutes: 0,
		children_count: 0,
		labels: [],
		status: {
			id: "st-1",
			name: "Todo",
			category: "todo",
			color: "#888",
			sort_order: 0,
			is_default: true,
		},
		sprint_id: partial.sprint_id ?? null,
		created_at: "2026-01-01T00:00:00Z",
		updated_at: "2026-01-01T00:00:00Z",
	};
}

const planningSprint = {
	id: "sprint-plan-1",
	name: "Sprint A",
	status: "planning" as const,
	start_date: "2026-01-01",
	end_date: "2026-01-14",
	issue_count: 0,
	created_at: "2026-01-01T00:00:00Z",
	updated_at: "2026-01-01T00:00:00Z",
};

describe("FE · SprintBacklog selection leak (A3W10-3-004)", () => {
	beforeEach(() => {
		listIssues.mockReset();
		moveIssueToSprint.mockClear();
		listIssues.mockResolvedValue({
			items: [
				issue({
					id: "old-uuid",
					issue_key: "HUB-1",
					title: "Old project issue",
				}),
			],
			total: 1,
			has_more: false,
		});
		useSprintStore.setState({
			sprints: [planningSprint],
			fetchSprints: vi.fn().mockResolvedValue(undefined),
			moveIssueToSprint,
		});
	});

	it("clears selectedIssues when projectId changes", async () => {
		const { rerender } = renderWithProviders(
			<SprintBacklog projectId="proj-old" />,
		);

		await waitFor(() => {
			expect(screen.getByText("Old project issue")).toBeInTheDocument();
		});

		fireEvent.click(
			screen.getByRole("checkbox", { name: /Выбрать HUB-1/i }),
		);
		expect(screen.getByText(/Выбрано:\s*1/)).toBeInTheDocument();

		listIssues.mockResolvedValue({
			items: [
				issue({
					id: "new-uuid",
					issue_key: "HUB-9",
					title: "New project issue",
				}),
			],
			total: 1,
			has_more: false,
		});

		rerender(<SprintBacklog projectId="proj-new" />);

		await waitFor(() => {
			expect(screen.getByText("New project issue")).toBeInTheDocument();
		});

		expect(screen.queryByText(/Выбрано:/)).not.toBeInTheDocument();
		expect(
			screen.getByRole("checkbox", { name: /Выбрать HUB-9/i }),
		).not.toBeChecked();
		expect(
			screen.queryByRole("button", { name: /В Sprint A/i }),
		).not.toBeInTheDocument();
		expect(moveIssueToSprint).not.toHaveBeenCalled();

		// Mutation: без setSelectedIssues(new Set()) на projectId → «Выбрано: 1» / FAIL.
	});
});
