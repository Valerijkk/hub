import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CoverageReport } from "@/modules/tms/components/coverage-report";
import { renderWithProviders } from "../../render";

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

describe("CoverageReport keyboard expand (A2W10-A2-007 / TASK-093)", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it("expandable issue row is keyboard-reachable", () => {
		vi.mocked(useQuery).mockReturnValue({
			data: {
				issues: [
					{
						issue_id: "i1",
						issue_key: "HUB-1",
						issue_title: "Login",
						test_cases: [
							{
								case_key: "TC-1",
								case_title: "ok",
								last_status: "passed",
							},
						],
					},
				],
			},
			isLoading: false,
			isError: false,
			error: null,
			refetch: vi.fn(),
		} as unknown as ReturnType<typeof useQuery>);

		renderWithProviders(<CoverageReport projectId="proj-1" />);

		const row = screen.getByRole("button", {
			name: /развернуть покрытие hub-1/i,
		});
		expect(row).toHaveAttribute("tabIndex", "0");
		expect(row).toHaveAttribute("aria-expanded", "false");
	});
});
