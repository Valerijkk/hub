import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { SprintBacklog } from "@/modules/tracker/sprint-backlog";
import { renderWithProviders } from "../support/render";

const listIssues = vi.fn();
const fetchSprints = vi.fn().mockResolvedValue(undefined);
const moveIssueToSprint = vi.fn().mockResolvedValue(undefined);

vi.mock("@/api/tracker", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@/api/tracker")>();
	return {
		...actual,
		listIssues: (...args: unknown[]) => listIssues(...args),
	};
});
vi.mock("@/stores/sprint-store", () => ({
	useSprintStore: () => ({
		sprints: [],
		fetchSprints,
		moveIssueToSprint,
	}),
}));
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, ...props }: { children: React.ReactNode }) => (
		<a {...props}>{children}</a>
	),
}));
vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("SprintBacklog error≠empty (A2W10-A3-002 / TASK-093)", () => {
	beforeEach(() => {
		listIssues.mockReset();
		fetchSprints.mockClear();
		moveIssueToSprint.mockClear();
	});

	it("shows alert + retry instead of «Бэклог пуст» when listIssues fails", async () => {
		listIssues.mockRejectedValue(new Error("network"));

		renderWithProviders(<SprintBacklog projectId="proj-1" />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить бэклог/i,
			);
		});
		expect(screen.queryByText(/бэклог пуст/i)).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
