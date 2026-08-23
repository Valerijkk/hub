import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import * as tmsApi from "@/api/tms";
import { TestCasesPage } from "@/routes/_authenticated/projects/$projectId/tms/cases.lazy";
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
	};
});

vi.mock("@/stores/tms-store", () => ({
	useTmsStore: () => ({
		filters: {
			suiteId: undefined,
			status: undefined,
			priority: undefined,
			search: undefined,
		},
		updateFilter: vi.fn(),
		resetFilters: vi.fn(),
	}),
}));

vi.mock("@/lib/format-api-error", () => ({
	toastApiError: vi.fn(),
}));

vi.mock("@/modules/tms/components/case-tree", () => ({
	CaseTreePanel: () => <div data-testid="case-tree" />,
}));

describe("TestCasesPage error≠empty (A3W10-4-003 / TASK-127)", () => {
	beforeEach(() => {
		vi.restoreAllMocks();
	});

	it("shows alert + retry instead of «Нет тест-кейсов» when listTestCases fails", async () => {
		vi.spyOn(tmsApi, "listTestCases").mockRejectedValue(new Error("network"));
		vi.spyOn(tmsApi, "listSuites").mockResolvedValue([]);

		renderWithProviders(<TestCasesPage />);

		await waitFor(() => {
			expect(screen.getByRole("alert")).toHaveTextContent(
				/не удалось загрузить тест-кейсы/i,
			);
		});
		expect(screen.queryByText(/нет тест-кейсов/i)).not.toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
	});
});
