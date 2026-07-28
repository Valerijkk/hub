import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as trackerApi from "@/api/tracker";
import { SprintBacklog } from "@/modules/tracker/sprint-backlog";
import { renderWithProviders } from "../../render";

vi.mock("@/api/tracker");
vi.mock("@/stores/sprint-store", () => ({
	useSprintStore: () => ({
		sprints: [],
		fetchSprints: vi.fn(),
		moveIssueToSprint: vi.fn(),
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
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of «Бэклог пуст» when listIssues fails", async () => {
		vi.mocked(trackerApi.listIssues).mockRejectedValue(new Error("network"));

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
