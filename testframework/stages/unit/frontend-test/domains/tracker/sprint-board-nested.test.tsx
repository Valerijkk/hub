import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SprintBoard } from "@/modules/tracker/sprint-board";
import { useSprintStore } from "@/stores/sprint-store";

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
		<a href="/sprint" {...props}>
			{children}
		</a>
	),
}));

describe("SprintBoard nested interactive (A2W10-A3-005 / TASK-093)", () => {
	beforeEach(() => {
		useSprintStore.setState({
			sprints: [
				{
					id: "sp1",
					name: "Sprint 1",
					goal: "Ship",
					status: "planning",
					start_date: "2026-07-01",
					end_date: "2026-07-14",
					issue_count: 3,
					created_at: "2026-07-01T00:00:00Z",
					updated_at: "2026-07-01T00:00:00Z",
				},
			],
			activeSprint: null,
			loading: false,
			error: null,
			fetchSprints: vi.fn().mockResolvedValue(undefined),
			startSprint: vi.fn(),
			completeSprint: vi.fn(),
		});
	});

	it("action button is not nested inside the sprint link", () => {
		render(<SprintBoard projectId="proj-1" />);

		const start = screen.getByRole("button", { name: /начать спринт/i });
		expect(start.closest("a")).toBeNull();
		expect(screen.getByRole("link")).toBeInTheDocument();
	});
});
