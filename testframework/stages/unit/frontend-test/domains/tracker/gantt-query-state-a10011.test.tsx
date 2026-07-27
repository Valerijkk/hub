import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { GanttChart } from "@/modules/tracker/gantt-chart";
import { renderWithProviders } from "../../render";

const refetch = vi.fn();

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

import { useQuery } from "@tanstack/react-query";

describe("GanttChart QueryStateView (A10-011)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of empty due-date copy when issues query fails", () => {
		vi.mocked(useQuery).mockImplementation(((opts: { queryKey?: unknown[] }) => {
			const key = opts.queryKey?.[0];
			if (key === "gantt-issues") {
				return {
					data: undefined,
					isLoading: false,
					isError: true,
					error: new Error("network"),
					refetch,
				};
			}
			return {
				data: [],
				isLoading: false,
				isError: false,
				refetch: vi.fn(),
			};
		}) as typeof useQuery);

		renderWithProviders(<GanttChart projectId="proj-1" />);

		expect(screen.getByRole("alert")).toHaveTextContent(
			/не удалось загрузить данные/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/нет задач с due date/i),
		).not.toBeInTheDocument();
	});
});
