import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CoverageReport } from "@/modules/tms/components/coverage-report";
import { renderWithProviders } from "../support/render";

const refetch = vi.fn();

vi.mock("@tanstack/react-query", async () => {
	const actual = await vi.importActual<typeof import("@tanstack/react-query")>(
		"@tanstack/react-query",
	);
	return {
		...actual,
		useQuery: vi.fn(),
	};
});

import { useQuery } from "@tanstack/react-query";

describe("CoverageReport error≠empty (W10-A2-007 / TASK-047)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("shows alert + retry instead of empty coverage copy when query fails", () => {
		vi.mocked(useQuery).mockReturnValue({
			data: undefined,
			isLoading: false,
			isError: true,
			error: new Error("network"),
			refetch,
		} as ReturnType<typeof useQuery>);

		renderWithProviders(<CoverageReport projectId="proj-1" />);

		expect(screen.getByRole("alert")).toHaveTextContent(
			/не удалось загрузить отчёт о покрытии/i,
		);
		expect(
			screen.getByRole("button", { name: /попробовать снова/i }),
		).toBeInTheDocument();
		expect(
			screen.queryByText(/нет задач с привязанными тест-кейсами/i),
		).not.toBeInTheDocument();
	});
});
