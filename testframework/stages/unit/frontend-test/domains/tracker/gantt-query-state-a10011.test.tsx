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
		useInfiniteQuery: vi.fn(),
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

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";

describe("GanttChart QueryStateView (A10-011)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.mocked(useQuery).mockReturnValue({
			data: [],
			isLoading: false,
			isError: false,
			refetch: vi.fn(),
		} as ReturnType<typeof useQuery>);
	});

	it("shows alert + retry instead of empty due-date copy when issues query fails", () => {
		vi.mocked(useInfiniteQuery).mockReturnValue({
			data: undefined,
			isLoading: false,
			isError: true,
			error: new Error("network"),
			refetch,
			hasNextPage: false,
			isFetchingNextPage: false,
			fetchNextPage: vi.fn(),
		} as unknown as ReturnType<typeof useInfiniteQuery>);

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
