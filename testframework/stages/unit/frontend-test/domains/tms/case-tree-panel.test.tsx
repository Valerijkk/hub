import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CaseTreePanel } from "@/modules/tms/components/case-tree";

describe("CaseTreePanel live export (A2W10-A2-008 / TASK-093)", () => {
	it("renders suite tree and all-cases control", () => {
		const onSelect = vi.fn();
		render(
			<CaseTreePanel
				suites={[
					{
						id: "su1",
						name: "Auth",
						sort_order: 0,
						created_at: "2026-01-01T00:00:00Z",
						updated_at: "2026-01-01T00:00:00Z",
						children: [],
					},
				]}
				selectedSuiteId={undefined}
				onSelectSuite={onSelect}
				total={4}
			/>,
		);

		expect(screen.getByText("Тест-сьюты")).toBeInTheDocument();
		expect(
			screen.getByRole("button", { name: /все кейсы/i }),
		).toBeInTheDocument();
		expect(screen.getByText("Auth")).toBeInTheDocument();
		expect(screen.getByText("4")).toBeInTheDocument();
	});
});
