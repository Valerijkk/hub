import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as projectsApi from "@/api/projects";
import { ProjectPage } from "@/routes/_authenticated/projects/$projectId/index.lazy";
import { renderWithProviders } from "../support/render";

vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		createLazyFileRoute: () => (opts: { component: React.ComponentType }) => ({
			...opts,
			useParams: () => ({ projectId: "proj-1" }),
		}),
		Link: ({ children, ...props }: { children: React.ReactNode }) => (
			<a {...props}>{children}</a>
		),
		useNavigate: () => vi.fn(),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

describe("ProjectPage error≠empty (A3W10-4-002 / TASK-127)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows alert + retry instead of «Проект не найден» when getProject fails", async () => {
		vi.spyOn(projectsApi, "getProject").mockRejectedValue(new Error("network"));
		vi.spyOn(projectsApi, "listMembers").mockResolvedValue([]);

		renderWithProviders(<ProjectPage />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить проект/i,
			);
		});
		expect(screen.queryByText(/проект не найден/i)).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
