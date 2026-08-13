import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as trackerApi from "@/api/tracker";
import { LabelsSettingsPage } from "@/routes/_authenticated/projects/$projectId/settings/labels";
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

describe("LabelsSettingsPage error≠empty (A3W10-4-004 / TASK-127)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows alert, hides empty copy and create UI when listLabels fails", async () => {
		vi.spyOn(trackerApi, "listLabels").mockRejectedValue(new Error("network"));
		vi.spyOn(trackerApi, "listLabelGroups").mockResolvedValue([]);

		renderWithProviders(<LabelsSettingsPage />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить метки/i,
			);
		});
		expect(screen.queryByText(/нет меток без группы/i)).not.toBeInTheDocument();
		expect(
			screen.queryByRole("heading", { name: /добавить метку/i }),
		).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
