import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SprintBacklog } from "@/modules/tracker/sprint-backlog";
import { useSprintStore } from "@/stores/sprint-store";
import { renderWithProviders } from "../support/render";

const listIssues = vi.fn();
const toastApiError = vi.fn();

vi.mock("@/api/tracker", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tracker")>();
	return {
		...actual,
		listIssues: (...args: unknown[]) => listIssues(...args),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: (...args: unknown[]) => toastApiError(...args),
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
		sprint_id: partial.sprint_id,
		created_at: "2026-01-01T00:00:00Z",
		updated_at: "2026-01-01T00:00:00Z",
	};
}

describe("FE · SprintBacklog without-sprint filter (TASK-082)", () => {
	beforeEach(() => {
		listIssues.mockReset();
		toastApiError.mockReset();
		useSprintStore.setState({
			sprints: [],
			fetchSprints: vi.fn().mockResolvedValue(undefined),
			moveIssueToSprint: vi.fn().mockResolvedValue(undefined),
		});
	});

	it("requests sprint_id=none and hides issues already in a sprint", async () => {
		listIssues.mockResolvedValue({
			items: [
				issue({
					id: "b1",
					issue_key: "HUB-1",
					title: "Backlog only",
					sprint_id: null,
				}),
				issue({
					id: "s1",
					issue_key: "HUB-2",
					title: "In sprint",
					sprint_id: "sprint-1",
				}),
			],
			total: 2,
			has_more: false,
		});

		renderWithProviders(<SprintBacklog projectId="proj-1" />);

		await waitFor(() => {
			expect(listIssues).toHaveBeenCalledWith("proj-1", {
				limit: 100,
				sprint_id: "none",
			});
		});

		await waitFor(() => {
			expect(screen.getByText("1 задач без спринта")).toBeInTheDocument();
		});
		expect(screen.getByText("Backlog only")).toBeInTheDocument();
		expect(screen.queryByText("In sprint")).not.toBeInTheDocument();
	});

	it("shows error state instead of empty backlog on fail", async () => {
		listIssues.mockRejectedValue(new Error("network"));

		renderWithProviders(<SprintBacklog projectId="proj-1" />);

		await waitFor(() => {
			expect(
				screen.getByText("Не удалось загрузить бэклог"),
			).toBeInTheDocument();
		});
		expect(screen.queryByText("Бэклог пуст")).not.toBeInTheDocument();
		expect(toastApiError).toHaveBeenCalled();
	});
});
