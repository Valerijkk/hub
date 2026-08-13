import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as trackerApi from "@/api/tracker";
import { WorkflowSettingsPage } from "@/routes/_authenticated/projects/$projectId/settings/workflow";
import { renderWithProviders } from "../support/render";

vi.mock("@tanstack/react-router", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-router")>(
		"@tanstack/react-router",
	);
	return {
		...actual,
		createFileRoute: () => (opts: { component: React.ComponentType }) => ({
			...opts,
			useParams: () => ({ projectId: "proj-1" }),
		}),
		Link: ({ children, ...props }: { children: React.ReactNode }) => (
			<a {...props}>{children}</a>
		),
	};
});

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

vi.mock("@/components/tracker/automation-rules-panel", () => ({
	AutomationRulesPanel: () => <div data-testid="automation-rules" />,
}));

describe("WorkflowSettingsPage error≠empty (A3W10-4-005 / TASK-127)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows alert and blocks create when statuses query fails", async () => {
		vi.spyOn(trackerApi, "listWorkflowStatuses").mockRejectedValue(
			new Error("network"),
		);
		vi.spyOn(trackerApi, "listWorkflowTransitions").mockResolvedValue([]);

		renderWithProviders(<WorkflowSettingsPage />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить статусы/i,
			);
		});
		expect(
			screen.queryByText(/нет статусов\. создайте первый/i),
		).not.toBeInTheDocument();
		expect(screen.queryByText(/добавить статус/i)).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
