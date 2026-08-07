import { fireEvent, screen } from "@testing-library/react";
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

	function mockCoverageReport() {
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
	}

	it("expands from the row-wide pointer control outside the trailing action cell", () => {
		mockCoverageReport();

		renderWithProviders(<CoverageReport projectId="proj-1" />);

		const control = screen.getByRole("button", {
			name: /развернуть покрытие hub-1/i,
		});
		const trailingActionCell = screen.getByText("Развернуть").closest("td");

		expect(control.closest("td")).not.toBe(trailingActionCell);
		expect(control.className).toMatch(/absolute/);
		expect(control.className).toMatch(/inset-0/);
		fireEvent.click(control);

		expect(control).toHaveAttribute("aria-expanded", "true");
		expect(screen.getByText("TC-1")).toBeInTheDocument();
	});

	it("exposes a focusable native keyboard control", () => {
		mockCoverageReport();

		renderWithProviders(<CoverageReport projectId="proj-1" />);

		const control = screen.getByRole("button", {
			name: /развернуть покрытие hub-1/i,
		});
		control.focus();
		expect(control).toHaveFocus();
		expect(control.tagName).toBe("BUTTON");
		expect(control.tabIndex).toBe(0);
		expect(control).toHaveAttribute("aria-expanded", "false");
	});
});
